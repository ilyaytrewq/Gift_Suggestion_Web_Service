package vkid

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

const (
	defaultTokenURL = "https://id.vk.ru/oauth2/auth"
	maxResponseBody = 1 << 20
)

type Client struct {
	httpClient *http.Client
	tokenURL   string
}

func NewClient(cfg config.VKConfig) *Client {
	tokenURL := strings.TrimSpace(cfg.OAuthTokenURL)
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}

	return &Client{
		httpClient: &http.Client{Timeout: cfg.RequestTimeout},
		tokenURL:   tokenURL,
	}
}

type ExchangeInput struct {
	ClientID     string
	Code         string
	CodeVerifier string
	RedirectURI  string
	DeviceID     string
	State        string
}

type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	UserID       string
	Scope        string
	State        string
}

type tokenResponsePayload struct {
	AccessToken      string          `json:"access_token"`
	RefreshToken     string          `json:"refresh_token"`
	ExpiresIn        int             `json:"expires_in"`
	UserID           json.RawMessage `json:"user_id"`
	Scope            string          `json:"scope"`
	State            string          `json:"state"`
	Error            string          `json:"error"`
	ErrorDescription string          `json:"error_description"`
}

func (c *Client) ExchangeAuthorizationCode(ctx context.Context, input ExchangeInput) (TokenResponse, error) {
	if c == nil || c.httpClient == nil {
		return TokenResponse{}, apperrors.New(
			apperrors.KindInternal,
			"vk_oauth_client_unavailable",
			"vk oauth client is not configured",
		)
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", strings.TrimSpace(input.ClientID))
	form.Set("code", strings.TrimSpace(input.Code))
	form.Set("code_verifier", strings.TrimSpace(input.CodeVerifier))
	form.Set("redirect_uri", strings.TrimSpace(input.RedirectURI))
	form.Set("device_id", strings.TrimSpace(input.DeviceID))
	form.Set("state", strings.TrimSpace(input.State))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, apperrors.Wrap(
			apperrors.KindInternal,
			"vk_oauth_request_failed",
			"failed to create vk oauth request",
			err,
		)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TokenResponse{}, apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_oauth_unreachable",
			"vk oauth service is unavailable",
			err,
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return TokenResponse{}, apperrors.Wrap(
			apperrors.KindInternal,
			"vk_oauth_response_read_failed",
			"failed to read vk oauth response",
			err,
		)
	}

	var payload tokenResponsePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return TokenResponse{}, apperrors.Wrap(
			apperrors.KindInternal,
			"vk_oauth_response_invalid",
			"vk oauth response is invalid",
			err,
		)
	}

	if payload.Error != "" || payload.ErrorDescription != "" {
		return TokenResponse{}, mapOAuthError(payload.Error, payload.ErrorDescription, resp.StatusCode)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return TokenResponse{}, apperrors.New(
			apperrors.KindUnavailable,
			"vk_oauth_exchange_failed",
			fmt.Sprintf("vk oauth exchange failed with status %d", resp.StatusCode),
		)
	}

	userID, err := parseUserID(payload.UserID)
	if err != nil {
		return TokenResponse{}, apperrors.Wrap(
			apperrors.KindInternal,
			"vk_oauth_user_id_invalid",
			"vk oauth user id is invalid",
			err,
		)
	}

	if strings.TrimSpace(payload.AccessToken) == "" || userID == "" {
		return TokenResponse{}, apperrors.New(
			apperrors.KindInternal,
			"vk_oauth_token_missing",
			"vk oauth response does not contain access token",
		)
	}

	return TokenResponse{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresIn:    payload.ExpiresIn,
		UserID:       userID,
		Scope:        payload.Scope,
		State:        payload.State,
	}, nil
}

func parseUserID(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("user_id is empty")
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return strings.TrimSpace(asString), nil
	}

	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return strings.TrimSpace(asNumber.String()), nil
	}

	return "", fmt.Errorf("unsupported user_id format")
}

func mapOAuthError(code, description string, statusCode int) error {
	message := strings.TrimSpace(description)
	if message == "" {
		message = strings.TrimSpace(code)
	}
	if message == "" {
		message = "vk oauth exchange rejected the request"
	}

	switch strings.TrimSpace(code) {
	case "invalid_request", "invalid_scope", "invalid_client":
		return apperrors.New(apperrors.KindValidation, "vk_oauth_invalid_request", message)
	case "access_denied":
		return apperrors.New(apperrors.KindForbidden, "vk_oauth_access_denied", message)
	default:
		if statusCode == http.StatusBadRequest {
			return apperrors.New(apperrors.KindValidation, "vk_oauth_invalid_request", message)
		}

		return apperrors.New(apperrors.KindUnavailable, "vk_oauth_exchange_failed", message)
	}
}

func ParseUserIDString(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("user id is empty")
	}
	if _, err := strconv.ParseInt(trimmed, 10, 64); err != nil {
		return "", err
	}

	return trimmed, nil
}
