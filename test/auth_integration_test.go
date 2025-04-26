package test_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"wonk/cmd/server"
)

func IntegrationTest(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("skipping integration tests: set INTEGRATION environment variable")
	}
}

// Test handles the following flow: New User signs up, logins in,
// requests a page with auth enabled and user is able to get page
// This test case handles the happy path
func TestAuthHandlers(t *testing.T) {
	IntegrationTest(t)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	endpoint := "http://localhost:8070"
	go server.Run(ctx, getTestSecrets, nil, []string{"--exclude-env", "-logfmt=devlog", "--test-db"})
	waitForReady(ctx, time.Second*5, endpoint+"/health")

	// Begin Testing workflow
	mockUsername := "testUser"
	mockPassword := "mockPassword!"
	// Sign up and create first user
	resp, err := http.DefaultClient.PostForm(endpoint+"/signup", url.Values{
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
	resp, err = http.DefaultClient.PostForm(endpoint+"/login", url.Values{
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
	req, err := http.NewRequest(http.MethodGet, endpoint+"/finance", nil)
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

// waitForReady calls the specified endpoint until it gets a 200
// response or until the context is cancelled or the timeout is
// reached.
func waitForReady(
	ctx context.Context,
	timeout time.Duration,
	endpoint string,
) error {
	client := http.Client{}
	startTime := time.Now()
	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			endpoint,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %s\n", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Endpoint is ready!")
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) >= timeout {
				return fmt.Errorf("timeout reached while waiting for endpoint")
			}
			// wait a little while between checks
			time.Sleep(250 * time.Millisecond)
		}
	}
}
