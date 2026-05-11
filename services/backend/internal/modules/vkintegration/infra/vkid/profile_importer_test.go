package vkid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

func TestProfileImporterImportInterests(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if r.Form.Get("client_id") != "54584723" {
			t.Fatalf("client_id = %q", r.Form.Get("client_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"user": {
				"user_id": "123",
				"first_name": "Иван",
				"last_name": "Петров"
			}
		}`))
	}))
	t.Cleanup(server.Close)

	importer := NewProfileImporter(config.VKConfig{
		Enabled:          true,
		RequestTimeout:     3 * time.Second,
		AppID:            "54584723",
		OAuthUserInfoURL: server.URL,
	})

	result, err := importer.ImportInterests(context.Background(), vkintegrationusecase.ImportInterestsRequest{
		ProviderUserID: "123",
		AccessToken:    "token",
	})
	if err != nil {
		t.Fatalf("ImportInterests() error = %v", err)
	}
	if len(result.Interests) != 0 {
		t.Fatalf("expected no interests, got %d", len(result.Interests))
	}
	if result.ProfileScreenName == nil || *result.ProfileScreenName != "Иван Петров" {
		t.Fatalf("ProfileScreenName = %v, want %q", result.ProfileScreenName, "Иван Петров")
	}
	if result.ProfileURL == nil || *result.ProfileURL != "https://vk.com/id123" {
		t.Fatalf("ProfileURL = %v, want https://vk.com/id123", result.ProfileURL)
	}
}

func TestProfileImporterImportInterestsIgnoresUserInfoFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	importer := NewProfileImporter(config.VKConfig{
		Enabled:          true,
		RequestTimeout:     3 * time.Second,
		AppID:            "app",
		OAuthUserInfoURL: server.URL,
	})

	result, err := importer.ImportInterests(context.Background(), vkintegrationusecase.ImportInterestsRequest{
		AccessToken: "token",
	})
	if err != nil {
		t.Fatalf("ImportInterests() error = %v", err)
	}
	if len(result.Interests) != 0 {
		t.Fatalf("expected no interests, got %d", len(result.Interests))
	}
}
