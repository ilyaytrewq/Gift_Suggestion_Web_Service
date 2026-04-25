package vk

import (
	"context"
	"time"

	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

type Client struct {
	enabled        bool
	requestTimeout time.Duration
}

func NewClient(cfg config.VKConfig) *Client {
	return &Client{
		enabled:        cfg.Enabled,
		requestTimeout: cfg.RequestTimeout,
	}
}

func (c *Client) ImportInterests(context.Context, vkintegrationusecase.ImportInterestsRequest) (vkintegrationusecase.ImportInterestsResult, error) {
	if !c.enabled {
		return vkintegrationusecase.ImportInterestsResult{}, vkintegrationusecase.ErrInterestImportUnavailable
	}

	return vkintegrationusecase.ImportInterestsResult{}, vkintegrationusecase.ErrInterestImportNotImplemented
}
