package vk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
)

func newTestClient(serverURL string) *Client {
	return &Client{
		enabled:    true,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    serverURL,
	}
}

func serveGroups(t *testing.T, w http.ResponseWriter, count int, items []map[string]any) {
	t.Helper()
	if items == nil {
		items = []map[string]any{}
	}
	if err := json.NewEncoder(w).Encode(map[string]any{
		"response": map[string]any{"count": count, "items": items},
	}); err != nil {
		t.Errorf("encode: %v", err)
	}
}

func serveAPIError(t *testing.T, w http.ResponseWriter, code int, msg string) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"error_code": code, "error_msg": msg},
	}); err != nil {
		t.Errorf("encode: %v", err)
	}
}

func TestImportInterestsDisabled(t *testing.T) {
	t.Parallel()

	client := &Client{enabled: false}
	_, err := client.ImportInterests(context.Background(), vkintegrationusecase.ImportInterestsRequest{})

	if !errors.Is(err, vkintegrationusecase.ErrInterestImportUnavailable) {
		t.Fatalf("expected ErrInterestImportUnavailable, got %v", err)
	}
}

func TestImportInterestsSinglePage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveGroups(t, w, 2, []map[string]any{
			{"id": 1, "name": "Gaming", "description": "Games", "activity": "Games"},
			{"id": 2, "name": "Music", "description": "Music fans", "activity": "Music"},
		})
	}))
	defer server.Close()

	result, err := newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{AccessToken: "token"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Interests) != 2 {
		t.Fatalf("expected 2 interests, got %d", len(result.Interests))
	}
	if result.Interests[0].Name != "Gaming" {
		t.Fatalf("expected name 'Gaming', got %q", result.Interests[0].Name)
	}
	if result.Interests[0].SourceLabel != "vk_group" {
		t.Fatalf("expected source_label 'vk_group', got %q", result.Interests[0].SourceLabel)
	}
	if result.Interests[0].Position != 1 || result.Interests[1].Position != 2 {
		t.Fatalf("wrong positions: %d, %d", result.Interests[0].Position, result.Interests[1].Position)
	}
}

func TestImportInterestsPagination(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch int(callCount.Add(1)) {
		case 1:
			serveGroups(t, w, 3, []map[string]any{
				{"id": 1, "name": "Group A"},
				{"id": 2, "name": "Group B"},
			})
		case 2:
			serveGroups(t, w, 3, []map[string]any{
				{"id": 3, "name": "Group C"},
			})
		default:
			t.Error("unexpected third API call")
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	result, err := newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(callCount.Load()) != 2 {
		t.Fatalf("expected 2 API calls, got %d", callCount.Load())
	}
	if len(result.Interests) != 3 {
		t.Fatalf("expected 3 interests, got %d", len(result.Interests))
	}
	if result.Interests[2].Name != "Group C" {
		t.Fatalf("expected 'Group C', got %q", result.Interests[2].Name)
	}
}

func TestImportInterestsEmptyList(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveGroups(t, w, 0, nil)
	}))
	defer server.Close()

	result, err := newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Interests) != 0 {
		t.Fatalf("expected 0 interests, got %d", len(result.Interests))
	}
}

func TestImportInterestsAPIErrorGroupsAccessDenied(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveAPIError(t, w, 260, "Access to groups list denied")
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{},
	)
	if !errors.Is(err, vkintegrationusecase.ErrVKGroupsAccessDenied) {
		t.Fatalf("expected ErrVKGroupsAccessDenied, got %v", err)
	}
}

func TestImportInterestsAPIErrorGroupsScopeRequired(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveAPIError(t, w, 1051, "Method is not available with this access token")
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{},
	)
	if !errors.Is(err, vkintegrationusecase.ErrVKGroupsScopeRequired) {
		t.Fatalf("expected ErrVKGroupsScopeRequired, got %v", err)
	}
}

func TestImportInterestsAPIErrorInvalidToken(t *testing.T) {
	t.Parallel()

	for _, code := range []int{5, 1117} {
		code := code
		t.Run(fmt.Sprintf("error_code_%d", code), func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serveAPIError(t, w, code, "User authorization failed")
			}))
			defer server.Close()

			_, err := newTestClient(server.URL).ImportInterests(
				context.Background(), vkintegrationusecase.ImportInterestsRequest{},
			)
			if !errors.Is(err, vkintegrationusecase.ErrVKTokenInvalid) {
				t.Fatalf("error_code=%d: expected ErrVKTokenInvalid, got %v", code, err)
			}
		})
	}
}

func TestImportInterestsAPIErrorRateLimit(t *testing.T) {
	t.Parallel()

	for _, code := range []int{6, 29} {
		code := code
		t.Run(fmt.Sprintf("error_code_%d", code), func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serveAPIError(t, w, code, "Too many requests per second")
			}))
			defer server.Close()

			_, err := newTestClient(server.URL).ImportInterests(
				context.Background(), vkintegrationusecase.ImportInterestsRequest{},
			)
			if !errors.Is(err, vkintegrationusecase.ErrVKRateLimited) {
				t.Fatalf("error_code=%d: expected ErrVKRateLimited, got %v", code, err)
			}
		})
	}
}

func TestImportInterestsMalformedJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not valid json {"))
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{},
	)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestImportInterestsHTTPNon2xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{},
	)
	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
}

func TestImportInterestsUsesPostMethod(t *testing.T) {
	t.Parallel()

	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		serveGroups(t, w, 0, nil)
	}))
	defer server.Close()

	_, _ = newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{},
	)
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %q", gotMethod)
	}
}

func TestImportInterestsTokenNotInQueryString(t *testing.T) {
	t.Parallel()

	const secret = "super-secret-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, secret) {
			t.Errorf("access_token must not appear in query string, got: %q", r.URL.RawQuery)
		}
		serveGroups(t, w, 0, nil)
	}))
	defer server.Close()

	_, _ = newTestClient(server.URL).ImportInterests(
		context.Background(), vkintegrationusecase.ImportInterestsRequest{AccessToken: secret},
	)
}
