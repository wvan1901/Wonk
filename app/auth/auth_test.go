package auth_test

import (
	"encoding/hex"
	"log/slog"
	"os"
	"testing"
	"wonk/app/auth"
)

// Testing funcs: CreateToken & Verify Token
// Testing that jwt token is created and validated
func TestCreateAndValidateJwt(t *testing.T) {
	// Starting Auth Service
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	jwtMockSecret := "RANDOM_SECRET"
	cookieMockSecret := hex.EncodeToString([]byte("RANDOM_SECRET"))

	authService := auth.Auth{
		Logger:          logger,
		JwtSecretKey:    jwtMockSecret,
		CookieSecretKey: cookieMockSecret,
		User:            &mockUserService{},
	}

	tests := []struct {
		name          string
		inputUserName string
		inputId       int
		expectedErr   bool
	}{
		{name: "Test 1", inputUserName: "jbil12", inputId: 1, expectedErr: false},
		{name: "Test 2", inputUserName: "wwva27", inputId: 22, expectedErr: false},
		{name: "Test 3", inputUserName: "s0meUsr", inputId: 333, expectedErr: false},
		{name: "Test 4", inputUserName: "hacker21", inputId: 4444, expectedErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtToken, err := authService.CreateToken(tt.inputUserName, tt.inputId)
			if err != nil {
				t.Errorf("unexpected error in creating jwt token: err: %v", err)
			}
			err = authService.VerifyToken(jwtToken)
			if tt.expectedErr && err == nil {
				t.Errorf("expected an error but didnt get one")
			} else if !tt.expectedErr && err != nil {
				t.Errorf("didn't expected an error but did get one, err: %v", err)
			}
		})
	}
}

// Test Func: ReadTokenUserName
// Testing retrieving values from JWT is correct
// NOTE: Input Jwts were created manually, go to https://jwt.io/ for encoding jwts
/* Example Jwt data used
   Header: {"alg": "HS256","typ": "JWT"}
   Body: {"exp": 10000000000 ,"userId": "21","username": "testUser"}
*/
func TestGetJwtValues(t *testing.T) {
	// Starting Auth Service
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// NOTE: DON'T change this JWT Secret, otherwise we will have to recreate the input Jwt tokone since this is used for the signature
	jwtMockSecret := "RANDOM_SECRET2"
	cookieMockSecret := hex.EncodeToString([]byte("RANDOM_SECRET2"))

	authService := auth.Auth{
		Logger:          logger,
		JwtSecretKey:    jwtMockSecret,
		CookieSecretKey: cookieMockSecret,
		User:            &mockUserService{},
	}
	tests := []struct {
		name             string
		inputJwt         string
		expectedId       int
		expectedUserName string
		expectedErr      bool
	}{
		// NOTE: all valid token expirations are set to year 2286
		{
			name:             "Testing valid jwt 1",
			inputJwt:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDAwLCJ1c2VySWQiOiIyMSIsInVzZXJuYW1lIjoic29tZVVzZXIxMiJ9.WGIwXBYajAx8GosUMnkqL3MOimaOxbEvoRXgF5nQUj8",
			expectedId:       21,
			expectedUserName: "someUser12",
			expectedErr:      false,
		},
		{
			name:             "Testing valid jwt 2",
			inputJwt:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDAwLCJ1c2VySWQiOiIzMTU0NTMxNTQzMTU0MTcyNyIsInVzZXJuYW1lIjoidHdvTWFueUtuZWVzIn0.g2kb4g2bu7ge34WCiXxNs0qV02LIZn2k7x0fQU82E7U",
			expectedId:       31545315431541727,
			expectedUserName: "twoManyKnees",
			expectedErr:      false,
		},
		{
			name:             "Testing invalid jwt id value returns error",
			inputJwt:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMDAwLCJ1c2VySWQiOiJ3cm9uZyIsInVzZXJuYW1lIjoidXNlciJ9.HCxiFQw_Oj9q_LAq0bC4jQy-36CLb_HmDGJ2fPe8NvI",
			expectedId:       -1,
			expectedUserName: "",
			expectedErr:      true,
		},
		{
			name:             "Testing expired jwt returns error",
			inputJwt:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjEwMDAwMDAwMCwidXNlcklkIjoid3JvbmciLCJ1c2VybmFtZSI6InVzZXIifQ.urfRiR12LzKFgJTjj1mDXT-PjUuv_81YWqWYCSnsMys",
			expectedId:       -1,
			expectedUserName: "",
			expectedErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputUserName, outputUserId, err := authService.ReadTokenUserName(tt.inputJwt)
			if tt.expectedErr && err == nil {
				t.Errorf("expected an error but didnt get one")
			} else if !tt.expectedErr && err != nil {
				t.Errorf("didn't expected an error but did get one, err: %v", err)
			}
			if tt.expectedUserName != outputUserName {
				t.Errorf("userName: expected %s, got %s", tt.expectedUserName, outputUserName)
			}
			if tt.expectedId != outputUserId {
				t.Errorf("userName: expected %d, got %d", tt.expectedId, outputUserId)
			}
		})
	}

}

type mockUserService struct{}

func (m *mockUserService) Login(string, string) (int, error) {
	return 0, nil
}
func (m *mockUserService) CreateUser(string, string) (int, error) {
	return 0, nil
}
