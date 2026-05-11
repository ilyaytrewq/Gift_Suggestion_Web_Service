package vk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

const (
	vkAPIVersion    = "5.199"
	pageCount       = 1000
	maxResponseBody = 10 * 1024 * 1024
)

type Client struct {
	enabled    bool
	httpClient *http.Client
	baseURL    string
}

func NewClient(cfg config.VKConfig) *Client {
	return &Client{
		enabled:    cfg.Enabled,
		httpClient: &http.Client{Timeout: cfg.RequestTimeout},
		baseURL:    "https://api.vk.com/method",
	}
}

type vkGroupsResponse struct {
	Response *vkGroupsPage `json:"response,omitempty"`
	Error    *vkAPIError   `json:"error,omitempty"`
}

type vkGroupsPage struct {
	Count int       `json:"count"`
	Items []vkGroup `json:"items"`
}

type vkGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Activity    string `json:"activity"`
}

type vkAPIError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func (c *Client) ImportInterests(ctx context.Context, input vkintegrationusecase.ImportInterestsRequest) (vkintegrationusecase.ImportInterestsResult, error) {
	if !c.enabled {
		return vkintegrationusecase.ImportInterestsResult{}, vkintegrationusecase.ErrInterestImportUnavailable
	}

	var groups []vkGroup
	offset := 0

	for {
		page, total, err := c.fetchPage(ctx, input.AccessToken, input.ProviderUserID, offset)
		if err != nil {
			return vkintegrationusecase.ImportInterestsResult{}, err
		}

		groups = append(groups, page...)
		offset += len(page)

		if len(page) == 0 || offset >= total {
			break
		}
	}

	records := make([]vkintegrationusecase.ImportedInterestRecord, 0, len(groups))
	for i, g := range groups {
		if g.Name == "" {
			continue
		}
		records = append(records, vkintegrationusecase.ImportedInterestRecord{
			Name:        g.Name,
			SourceLabel: "vk_group",
			Position:    i + 1,
		})
	}

	return vkintegrationusecase.ImportInterestsResult{Interests: records}, nil
}

func (c *Client) fetchPage(ctx context.Context, accessToken, providerUserID string, offset int) ([]vkGroup, int, error) {
	form := url.Values{}
	form.Set("extended", "1")
	form.Set("fields", "description,activity")
	form.Set("count", strconv.Itoa(pageCount))
	form.Set("offset", strconv.Itoa(offset))
	form.Set("v", vkAPIVersion)
	form.Set("access_token", accessToken)
	if userID := strings.TrimSpace(providerUserID); userID != "" {
		form.Set("user_id", userID)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/groups.get",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("vk: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("vk: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("vk: unexpected HTTP status %d", resp.StatusCode)
	}

	rawBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, 0, fmt.Errorf("vk: read response: %w", err)
	}

	var parsed vkGroupsResponse
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return nil, 0, fmt.Errorf("vk: parse response: %w", err)
	}

	if parsed.Error != nil {
		return nil, 0, mapAPIError(parsed.Error)
	}

	if parsed.Response == nil {
		return nil, 0, fmt.Errorf("vk: missing response field")
	}

	return parsed.Response.Items, parsed.Response.Count, nil
}

func mapAPIError(e *vkAPIError) error {
	switch e.ErrorCode {
	case 5, 1116, 1117:
		return vkintegrationusecase.ErrVKTokenInvalid
	case 6, 29:
		return vkintegrationusecase.ErrVKRateLimited
	case 15, 260:
		return vkintegrationusecase.ErrVKGroupsAccessDenied
	case 1051:
		return vkintegrationusecase.ErrVKGroupsScopeRequired
	default:
		return &APIError{Code: e.ErrorCode, Message: e.ErrorMsg}
	}
}

// APIError is returned for unmapped VK API error codes.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("vk api error %d: %s", e.Code, e.Message)
}
