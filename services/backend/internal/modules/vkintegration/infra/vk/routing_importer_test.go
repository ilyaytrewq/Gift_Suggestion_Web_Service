package vk

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

func TestRoutingImporterUsesProfileWithoutGroupsScope(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"user_id":"1","first_name":"A","last_name":"B"}}`))
	}))
	t.Cleanup(server.Close)

	importer := NewRoutingImporter(config.VKConfig{
		Enabled:          true,
		RequestTimeout:   3 * time.Second,
		AppID:            "app",
		OAuthUserInfoURL: server.URL,
	})

	result, err := importer.ImportInterests(context.Background(), vkintegrationusecase.ImportInterestsRequest{
		ProviderUserID: "1",
		AccessToken:    "token",
		Scopes:         []string{"vkid.personal_info"},
	})
	if err != nil {
		t.Fatalf("ImportInterests() error = %v", err)
	}
	if result.ProfileScreenName == nil || *result.ProfileScreenName != "A B" {
		t.Fatalf("ProfileScreenName = %v, want %q", result.ProfileScreenName, "A B")
	}
}

func TestRoutingImporterFallsBackWhenGroupsAPIUnavailable(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path == "/groups.get" || calls == 1 {
			serveAPIError(t, w, 1051, "Method is not available with this access token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"user_id":"1","first_name":"Fallback","last_name":"User"}}`))
	}))
	t.Cleanup(server.Close)

	groupsClient := &Client{
		enabled:    true,
		httpClient: server.Client(),
		baseURL:    server.URL,
	}
	profileImporter := NewRoutingImporter(config.VKConfig{
		Enabled:          true,
		RequestTimeout:   3 * time.Second,
		AppID:            "app",
		OAuthUserInfoURL: server.URL,
	}).profile

	importer := &RoutingImporter{groups: groupsClient, profile: profileImporter}

	result, err := importer.ImportInterests(context.Background(), vkintegrationusecase.ImportInterestsRequest{
		ProviderUserID: "1",
		AccessToken:    "token",
		Scopes:         []string{"groups"},
	})
	if err != nil {
		t.Fatalf("ImportInterests() error = %v", err)
	}
	if result.ProfileScreenName == nil || *result.ProfileScreenName != "Fallback User" {
		t.Fatalf("ProfileScreenName = %v, want %q", result.ProfileScreenName, "Fallback User")
	}
}

func TestShouldFallbackFromGroups(t *testing.T) {
	t.Parallel()

	if !shouldFallbackFromGroups(vkintegrationusecase.ErrVKGroupsScopeRequired) {
		t.Fatal("expected fallback for ErrVKGroupsScopeRequired")
	}
	if !shouldFallbackFromGroups(&APIError{Code: 1051, Message: "unavailable"}) {
		t.Fatal("expected fallback for API error 1051")
	}
	if shouldFallbackFromGroups(errors.New("other")) {
		t.Fatal("did not expect fallback for unrelated error")
	}
}
