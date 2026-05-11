package vk

import (
	"context"
	"errors"
	"strings"

	vkid "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/infra/vkid"
	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

// RoutingImporter uses groups.get when the token has groups scope; otherwise VK ID profile fallback.
type RoutingImporter struct {
	groups  *Client
	profile *vkid.ProfileImporter
}

func NewRoutingImporter(cfg config.VKConfig) *RoutingImporter {
	return &RoutingImporter{
		groups:  NewClient(cfg),
		profile: vkid.NewProfileImporter(cfg),
	}
}

func (r *RoutingImporter) ImportInterests(
	ctx context.Context,
	input vkintegrationusecase.ImportInterestsRequest,
) (vkintegrationusecase.ImportInterestsResult, error) {
	if hasGroupsScope(input.Scopes) {
		result, err := r.groups.ImportInterests(ctx, input)
		if err == nil {
			return result, nil
		}
		if shouldFallbackFromGroups(err) {
			return r.profile.ImportInterests(ctx, input)
		}

		return result, err
	}

	return r.profile.ImportInterests(ctx, input)
}

func hasGroupsScope(scopes []string) bool {
	for _, scope := range scopes {
		if strings.EqualFold(strings.TrimSpace(scope), "groups") {
			return true
		}
	}

	return false
}

func shouldFallbackFromGroups(err error) bool {
	var apiErr *APIError
	return errors.Is(err, vkintegrationusecase.ErrVKGroupsScopeRequired) ||
		(errors.As(err, &apiErr) && apiErr.Code == 1051)
}
