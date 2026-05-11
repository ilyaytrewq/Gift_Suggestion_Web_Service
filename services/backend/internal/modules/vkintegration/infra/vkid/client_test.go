package vkid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

)

func TestClientExchangeAuthorizationCodeSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if r.Form.Get("grant_type") != "authorization_code" {
			t.Fatalf("grant_type = %q", r.Form.Get("grant_type"))
		}
		if r.Form.Get("code_verifier") != "verifier-123" {
			t.Fatalf("code_verifier = %q", r.Form.Get("code_verifier"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token":"token",
			"expires_in":3600,
			"user_id":12345,
			"scope":"groups"
		}`))
	}))
	t.Cleanup(server.Close)

	client := &Client{
		httpClient: server.Client(),
		tokenURL:   server.URL,
	}

	response, err := client.ExchangeAuthorizationCode(context.Background(), ExchangeInput{
		ClientID:     "app",
		Code:         "code",
		CodeVerifier: "verifier-123",
		RedirectURI:  "https://example.com/callback",
		DeviceID:     "device",
		State:        "state",
	})
	if err != nil {
		t.Fatalf("ExchangeAuthorizationCode() error = %v", err)
	}
	if response.AccessToken != "token" {
		t.Fatalf("AccessToken = %q, want token", response.AccessToken)
	}
	if response.UserID != "12345" {
		t.Fatalf("UserID = %q, want 12345", response.UserID)
	}
}

func TestClientExchangeAuthorizationCodeMapsOAuthError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":"invalid_request","error_description":"bad redirect"}`))
	}))
	t.Cleanup(server.Close)

	client := &Client{
		httpClient: server.Client(),
		tokenURL:   server.URL,
	}

	_, err := client.ExchangeAuthorizationCode(context.Background(), ExchangeInput{
		ClientID:     "app",
		Code:         "code",
		CodeVerifier: "verifier-123",
		RedirectURI:  "https://example.com/callback",
		DeviceID:     "device",
		State:        "state",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
