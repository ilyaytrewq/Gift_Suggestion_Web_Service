package vkid

import (
	"context"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
)

type UsecaseAdapter struct {
	Client *Client
}

func (a UsecaseAdapter) ExchangeAuthorizationCode(
	ctx context.Context,
	input vkintegrationusecase.OAuthExchangeRequest,
) (vkintegrationusecase.OAuthTokenResult, error) {
	if a.Client == nil {
		return vkintegrationusecase.OAuthTokenResult{}, ErrNilClient
	}

	response, err := a.Client.ExchangeAuthorizationCode(ctx, ExchangeInput{
		ClientID:     input.ClientID,
		Code:         input.Code,
		CodeVerifier: input.CodeVerifier,
		RedirectURI:  input.RedirectURI,
		DeviceID:     input.DeviceID,
		State:        input.State,
	})
	if err != nil {
		return vkintegrationusecase.OAuthTokenResult{}, err
	}

	return vkintegrationusecase.OAuthTokenResult{
		AccessToken: response.AccessToken,
		ExpiresIn:   response.ExpiresIn,
		UserID:      response.UserID,
		Scope:       response.Scope,
	}, nil
}
