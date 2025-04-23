package test_test

import (
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"wonk/app/secret"
	application "wonk/app/service"
	"wonk/business"
	"wonk/cmd/server"
	database "wonk/storage"
)

func IntegrationTest(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration tests: set INTEGRATION environment variable")
	}
}

// Test handles the following flow: New User signs up for account, logins in, and
// goes to a page with auth enabled and user is able to get in
// This test case handles the happy path
func TestAuthHandlers(t *testing.T) {
	IntegrationTest(t)

	// Start API server
	srv, err := setUpApiServer()
	if err != nil {
		t.Error("Error starting business:", err)
		return
	}
	t.Cleanup(srv.Close)
	time.Sleep(3 * time.Second)

	// Check if server started by making a request to health endpoint
	resp, err := http.DefaultClient.Get(srv.URL + "/health")
	if err != nil {
		t.Error("health resp error: ", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("health status isnt 200")
		return
	}

	// Begin workflow
	mockUsername := "testUser"
	mockPassword := "mockPassword!"
	// Sign up and create first user
	resp, err = http.DefaultClient.PostForm(srv.URL+"/signup", url.Values{
		"username": []string{mockUsername},
		"password": []string{mockPassword},
	})
	if err != nil {
		t.Error("sign up resp error:", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("signup status isnt 200")
		return
	}

	// Login with user
	resp, err = http.DefaultClient.PostForm(srv.URL+"/login", url.Values{
		"username": []string{mockUsername},
		"password": []string{mockPassword},
	})
	if err != nil {
		t.Error("login resp error:", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("login status isnt 200")
		return
	}

	// Save cookie to use for auth
	if len(resp.Cookies()) < 1 {
		t.Error("login: missing auth cookie:")
		return
	}
	authCookie := resp.Cookies()[0]
	if authCookie.Name != "WonkAuth" {
		t.Error("login: cookie is not auth")
		return
	}

	// Hit endpoint with auth
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/finance", nil)
	if err != nil {
		t.Error("finance req error:", err)
		return
	}
	req.AddCookie(authCookie)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error("finance resp error:", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("finance status isnt 200")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error("finance body error:", err)
		return
	}

	respHtml := string(body)
	if !strings.Contains(respHtml, "html") {
		t.Error("finance respinse isnt html")
		return
	}
}

func setUpApiServer() (*httptest.Server, error) {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	d, err := database.InitDb(":memory:")
	if err != nil {
		return nil, fmt.Errorf("setUpApiServer: db: %w", err)
	}

	err = d.InitTablesForTesting()
	if err != nil {
		return nil, fmt.Errorf("setUpApiServer: set up tables: %w", err)
	}

	s, err := secret.InitSecret(getTestSecrets)
	if err != nil {
		return nil, fmt.Errorf("setUpApiServer: init secrets: %w", err)
	}

	b, err := business.InitServices(s, l, d)
	if err != nil {
		return nil, fmt.Errorf("setUpApiServer: init business: %w", err)
	}
	a, err := application.InitServices(s, l, b)
	if err != nil {
		return nil, fmt.Errorf("setUpApiServer: init app: %w", err)
	}
	return httptest.NewServer(server.NewServer(l, d, a)), nil
}

func getTestSecrets(s string) string {
	switch s {
	case "COOKIE_SECRET_KEY":
		return hex.EncodeToString([]byte("RANDOM_SECRET"))
	case "JWT_SECRET_KEY":
		return "RANDOM_SECRET"
	default:
		return ""
	}
}
