package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"wonk/app/cuserr"
	"wonk/app/secret"
	"wonk/app/templates/views"
	"wonk/business/user"

	"github.com/golang-jwt/jwt/v5"
)

const (
	COOKIE_NAME = "WonkAuth"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

type UserInfo struct {
	UserName string
	UserId   int
}

type AuthService interface {
	HandleLogin() http.Handler
	HandleSignUp() http.Handler
	AuthMiddleware(http.Handler) http.Handler
}

type Auth struct {
	Logger          *slog.Logger
	JwtSecretKey    string
	CookieSecretKey string
	User            user.User
}

func InitAuthService(s *secret.Secret, l *slog.Logger, u user.User) AuthService {
	return &Auth{
		Logger:          l,
		JwtSecretKey:    s.JwtKey,
		CookieSecretKey: s.CookieKey,
		User:            u,
	}
}

// IDEA: Look into kid for keys
func (a *Auth) CreateToken(username string, userId int) (string, error) {
	secretKey := []byte(a.JwtSecretKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"userId":   strconv.Itoa(userId),
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		},
	)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("CreateToken: %w", err)
	}

	return tokenString, nil
}

func (a *Auth) VerifyToken(tokenString string) error {
	secretKey := []byte(a.JwtSecretKey)
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return secretKey, nil
	})

	if err != nil {
		return fmt.Errorf("VerifyToken: %w", err)
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}

func (a *Auth) ReadTokenUserName(tokenString string) (string, int, error) {
	secretKey := []byte(a.JwtSecretKey)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return secretKey, nil
	})
	if err != nil {
		return "", -1, fmt.Errorf("ReadTokenUserName: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"]
		if !ok {
			return "", -1, errors.New("ReadTokenUserName: username not found in jwt")
		}
		usernameStr, ok := username.(string)
		if !ok {
			return "", -1, errors.New("ReadTokenUserName: username type conversion err")
		}
		userId, ok := claims["userId"]
		if !ok {
			return "", -1, errors.New("ReadTokenUserName: userId not found in jwt")
		}
		userIdStr, ok := userId.(string)
		if !ok {
			return "", -1, errors.New("ReadTokenUserName: userId type conversion err")
		}
		userIdInt, err := strconv.Atoi(userIdStr)
		if err != nil {
			return "", -1, fmt.Errorf("ReadTokenUserName: userId strconv: %w", err)
		}
		return usernameStr, userIdInt, nil
	}
	return "", -1, errors.New("ReadTokenUserName: claims or vaild token error")
}

// IDEA: Encrypt cookie
func (a *Auth) CreateSignedCookie(token string) (*http.Cookie, error) {
	cookie := http.Cookie{
		Name:     COOKIE_NAME,
		Value:    token,
		Path:     "/",
		MaxAge:   60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	cookieSecretKey, err := hex.DecodeString(a.CookieSecretKey)
	if err != nil {
		return nil, fmt.Errorf("CreateSignedCookie: hex: %w", err)
	}
	mac := hmac.New(sha256.New, cookieSecretKey)
	mac.Write([]byte(cookie.Name))
	mac.Write([]byte(cookie.Value))
	signature := mac.Sum(nil)

	// Prepend the cookie value with the HMAC signature.
	cookie.Value = string(signature) + cookie.Value

	// Encode the cookie
	cookie.Value = base64.URLEncoding.EncodeToString([]byte(cookie.Value))

	return &cookie, nil
}

func (a *Auth) ReadSignedCookie(cookie *http.Cookie) (string, error) {
	// Decode the cookie
	decodedValue, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("Read: encoding: %w", err)
	}

	// Read in the signed value from the cookie. This should be in the format "{signature}{original value}"
	signedValue := string(decodedValue)

	// A SHA256 HMAC signature has a fixed length of 32 bytes
	if len(signedValue) < sha256.Size {
		return "", errors.New("readSigned: sha256: invalid value")
	}

	// Split apart the signature and original cookie value.
	signature := signedValue[:sha256.Size]
	value := signedValue[sha256.Size:]

	// Recalculate the HMAC signature of the cookie name and original value.
	cookieSecretKey, err := hex.DecodeString(a.CookieSecretKey)
	if err != nil {
		return "", fmt.Errorf("CreateSignedCookie: hex: %w", err)
	}
	mac := hmac.New(sha256.New, cookieSecretKey)
	mac.Write([]byte(COOKIE_NAME))
	mac.Write([]byte(value))
	expectedSignature := mac.Sum(nil)

	// Check that the recalculated signature matches the signature we received
	// in the cookie. If they match, we can be confident that the cookie name
	// and value haven't been edited by the client.
	if !hmac.Equal([]byte(signature), expectedSignature) {
		return "", errors.New("readSigned: hmac: invalid value")
	}

	// Return the original cookie value.
	return value, nil
}

func (a *Auth) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRedirect := func() {
			http.Redirect(w, r, "/login", 302)
		}
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if isHtmxRequest {
			handleRedirect = func() {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(302)
			}
		}
		c, err := r.Cookie(COOKIE_NAME)
		if err != nil {
			// Missing cookie so redirect to login
			a.Logger.Error("AuthMiddleware: cookie", slog.Any("error", err), slog.String("devMsg", "no auth cookie found"))
			handleRedirect()
			return
		}
		value, err := a.ReadSignedCookie(c)
		if err != nil {
			// cookie is invalid so remove cookie & redirect to login
			a.Logger.Error("AuthMiddleware: signed cookie", slog.Any("error", err), slog.String("devMsg", "auth cookie currupted"))
			handleRedirect()
			return
		}
		err = a.VerifyToken(value)
		if err != nil {
			// token is invalid so remove cookie & redirect to login
			a.Logger.Error("AuthMiddleware: cookie token", slog.Any("error", err), slog.String("devMsg", "auth cookie invalid"))
			handleRedirect()
			return
		}

		username, userId, err := a.ReadTokenUserName(value)
		if err != nil {
			// getting username error from jwt token
			a.Logger.Error("AuthMiddleware: jwt read token", slog.Any("error", err), slog.String("devMsg", "read username err"))
			handleRedirect()
			return
		}
		userInfo := UserInfo{UserName: username, UserId: userId}
		ctx := context.WithValue(r.Context(), userCtxKey, userInfo)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Auth) HandleLogin() http.Handler {
	funcName := "HandleLogin"
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*20)
			defer cancel()
			switch r.Method {
			case "GET":
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					signUpDiv := views.Login(views.LoginFormData{})
					err := signUpDiv.Render(ctx, w)
					if err != nil {
						a.Logger.Error(funcName, slog.String("HttpMethod", "GET"), slog.Any("error", err), slog.String("DevNote", "div render"))
					}
					return
				}
				loginPage := views.LoginPage(views.LoginFormData{})
				err := loginPage.Render(ctx, w)
				if err != nil {
					a.Logger.Error(funcName, slog.String("HttpMethod", "GET"), slog.Any("error", err))
				}
				return
			case "POST":
				err := r.ParseForm()
				if err != nil {
					a.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.Any("error", err))
					w.WriteHeader(502)
					return
				}

				userName := r.FormValue("username")
				password := r.FormValue("password")
				userId, err := a.User.Login(userName, password)
				if err != nil {
					a.Logger.Info(funcName, slog.String("HttpMethod", "POST"), slog.Any("error", err))
					w.WriteHeader(422)
					errMsg := "ERROR: " + loginConvertErrorMsg(err)
					formData := views.LoginFormData{
						FormErr: &errMsg,
					}
					loginForm := views.LoginForm(formData)
					err := loginForm.Render(ctx, w)
					if err != nil {
						a.Logger.Error(funcName, slog.String("Method", "POST"), slog.Any("error", err))
					}
					return
				}
				token, err := a.CreateToken(userName, userId)
				if err != nil {
					a.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.Any("error", err))
					w.WriteHeader(500)
					return
				}
				cookie, err := a.CreateSignedCookie(token)
				if err != nil {
					a.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.Any("error", err))
					w.WriteHeader(500)
					return
				}
				http.SetCookie(w, cookie)
				w.Header().Set("HX-Redirect", "/home")
				w.WriteHeader(200)
			default:
				w.WriteHeader(404)
			}
		},
	)
}

func (a *Auth) HandleSignUp() http.Handler {
	funcName := "HandleSignUp"
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*20)
			defer cancel()
			switch r.Method {
			case "GET":
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					signUpDiv := views.SignUp(views.LoginFormData{})
					err := signUpDiv.Render(ctx, w)
					if err != nil {
						a.Logger.Error(funcName, slog.String("HttpMethod", "GET"), slog.Any("error", err), slog.String("DevNote", "div render"))
					}
					return
				}
				loginPage := views.SignUpPage(views.LoginFormData{})
				err := loginPage.Render(ctx, w)
				if err != nil {
					a.Logger.Error(funcName, slog.String("HttpMethod", "GET"), slog.Any("error", err), slog.String("DevNote", "full page render"))
				}
				return
			case "POST":
				err := r.ParseForm()
				if err != nil {
					a.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.Any("error", err))
					w.WriteHeader(502)
					return
				}

				userName := r.FormValue("username")
				password := r.FormValue("password")
				_, err = a.User.CreateUser(userName, password)
				if err != nil {
					a.Logger.Info(funcName, slog.String("HttpMethod", "POST"), slog.Any("error", err))
					w.WriteHeader(422)
					errMsg := "ERROR: " + signUpConvertErrorMsg(err)
					formData := views.LoginFormData{
						FormErr: &errMsg,
					}
					signUpDiv := views.SignUpForm(formData)
					err := signUpDiv.Render(ctx, w)
					if err != nil {
						a.Logger.Error(funcName, slog.String("HttpMethod", "GET"), slog.Any("error", err), slog.String("DevNote", "div render"))
					}
					return
				}
				// NOTE: We should hash the password in the client for added security
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(200)
				return
			default:
				w.WriteHeader(404)
			}
		},
	)
}

func UserCtx(ctx context.Context) (*UserInfo, error) {
	user, ok := ctx.Value(userCtxKey).(UserInfo)
	if !ok {
		return nil, errors.New("UserCtx: userInfo not found")
	}
	return &user, nil
}

func signUpConvertErrorMsg(err error) string {
	invalidInputErr := &cuserr.InvalidInput{}
	if errors.As(err, invalidInputErr) {
		return invalidInputErr.Error()
	}

	itemAlreadyExistsErr := &cuserr.ItemAlreadyExists{}
	if errors.As(err, itemAlreadyExistsErr) {
		return itemAlreadyExistsErr.Error()
	}

	return "internal error"
}

func loginConvertErrorMsg(err error) string {
	invalidInputErr := &cuserr.InvalidInput{}
	if errors.As(err, invalidInputErr) {
		return invalidInputErr.Error()
	}

	notFoundErr := &cuserr.NotFound{}
	if errors.As(err, notFoundErr) {
		return notFoundErr.Error()
	}

	invalidCred := &cuserr.InvalidCred{}
	if errors.As(err, invalidCred) {
		return invalidCred.Error()
	}

	return "internal error"
}
