package vkid

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

const defaultUserInfoURL = "https://id.vk.ru/oauth2/user_info"

// ProfileImporter loads VK ID profile data when groups scope is unavailable.
// It always completes without error: interests list may be empty until groups access is granted.
type ProfileImporter struct {
	enabled    bool
	httpClient *http.Client
	userInfoURL string
	appID      string
}

func NewProfileImporter(cfg config.VKConfig) *ProfileImporter {
	userInfoURL := strings.TrimSpace(cfg.OAuthUserInfoURL)
	if userInfoURL == "" {
		userInfoURL = defaultUserInfoURL
	}

	return &ProfileImporter{
		enabled:     cfg.Enabled,
		httpClient:  &http.Client{Timeout: cfg.RequestTimeout},
		userInfoURL: userInfoURL,
		appID:       strings.TrimSpace(cfg.AppID),
	}
}

type userInfoResponse struct {
	User *userInfoPayload `json:"user"`
}

type userInfoPayload struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (p *ProfileImporter) ImportInterests(
	ctx context.Context,
	input vkintegrationusecase.ImportInterestsRequest,
) (vkintegrationusecase.ImportInterestsResult, error) {
	if !p.enabled {
		return vkintegrationusecase.ImportInterestsResult{}, vkintegrationusecase.ErrInterestImportUnavailable
	}

	result := vkintegrationusecase.ImportInterestsResult{
		Interests: []vkintegrationusecase.ImportedInterestRecord{},
	}

	user, err := p.fetchUserInfo(ctx, input.AccessToken)
	if err != nil || user == nil {
		return result, nil
	}

	screenName := strings.TrimSpace(strings.TrimSpace(user.FirstName + " " + user.LastName))
	if screenName != "" {
		result.ProfileScreenName = &screenName
	}

	profileUserID := strings.TrimSpace(user.UserID)
	if profileUserID == "" {
		profileUserID = strings.TrimSpace(input.ProviderUserID)
	}
	if profileUserID != "" {
		profileURL := "https://vk.com/id" + profileUserID
		result.ProfileURL = &profileURL
	}

	return result, nil
}

func (p *ProfileImporter) fetchUserInfo(ctx context.Context, accessToken string) (*userInfoPayload, error) {
	if strings.TrimSpace(accessToken) == "" || p.appID == "" {
		return nil, nil
	}

	form := url.Values{}
	form.Set("client_id", p.appID)
	form.Set("access_token", strings.TrimSpace(accessToken))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.userInfoURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil
	}

	var payload userInfoResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return payload.User, nil
}
