package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	authhttp "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/delivery/http"
	wishlistusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/usecase"
	httpapi "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/transport/httpapi"
)

var ErrNilWishlistService = errors.New("wishlist service is nil")

type service interface {
	CreateWishlist(ctx context.Context, input wishlistusecase.CreateWishlistInput) (wishlistusecase.CreateWishlistOutput, error)
	ListWishlists(ctx context.Context, input wishlistusecase.ListWishlistsInput) (wishlistusecase.ListWishlistsOutput, error)
	GetWishlist(ctx context.Context, input wishlistusecase.GetWishlistInput) (wishlistusecase.GetWishlistOutput, error)
	AddWishlistItem(ctx context.Context, input wishlistusecase.AddWishlistItemInput) (wishlistusecase.AddWishlistItemOutput, error)
	RemoveWishlistItem(ctx context.Context, input wishlistusecase.RemoveWishlistItemInput) (wishlistusecase.RemoveWishlistItemOutput, error)
	DeleteWishlist(ctx context.Context, input wishlistusecase.DeleteWishlistInput) (wishlistusecase.DeleteWishlistOutput, error)
}

type Handler struct {
	service        service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service service, authMiddleware gin.HandlerFunc) (*Handler, error) {
	if service == nil {
		return nil, ErrNilWishlistService
	}

	return &Handler{
		service:        service,
		authMiddleware: authMiddleware,
	}, nil
}

func (h *Handler) Register(root gin.IRouter) {
	wishlist := root.Group("/wishlist")
	wishlist.Use(h.authMiddleware)
	wishlist.GET("", h.getCurrentWishlist)
	wishlist.POST("/items", h.addCurrentWishlistItem)
	wishlist.DELETE("/items/:gift_id", h.removeCurrentWishlistItem)
	wishlist.DELETE("", h.deleteCurrentWishlist)

	wishlists := root.Group("/wishlists")
	wishlists.Use(h.authMiddleware)
	wishlists.POST("", h.createWishlist)
	wishlists.GET("", h.listWishlists)
	wishlists.GET("/:wishlist_id", h.getWishlist)
	wishlists.POST("/:wishlist_id/items", h.addWishlistItem)
	wishlists.DELETE("/:wishlist_id/items/:gift_id", h.removeWishlistItem)
	wishlists.DELETE("/:wishlist_id", h.deleteWishlist)
}

func (h *Handler) getCurrentWishlist(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.GetWishlist(c.Request.Context(), wishlistusecase.GetWishlistInput{
		UserID: actor.UserID,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) addCurrentWishlistItem(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request addWishlistItemRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.AddWishlistItem(c.Request.Context(), wishlistusecase.AddWishlistItemInput{
		UserID: actor.UserID,
		GiftID: request.GiftID,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	status := http.StatusCreated
	if output.AlreadyInWishlist {
		status = http.StatusOK
	}

	httpapi.Success(c, status, output)
}

func (h *Handler) removeCurrentWishlistItem(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.RemoveWishlistItem(c.Request.Context(), wishlistusecase.RemoveWishlistItemInput{
		UserID: actor.UserID,
		GiftID: c.Param("gift_id"),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) deleteCurrentWishlist(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.DeleteWishlist(c.Request.Context(), wishlistusecase.DeleteWishlistInput{
		UserID: actor.UserID,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) createWishlist(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request createWishlistRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.CreateWishlist(c.Request.Context(), wishlistusecase.CreateWishlistInput{
		UserID: actor.UserID,
		Name:   request.Name,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusCreated, output)
}

func (h *Handler) listWishlists(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	input, err := parseListWishlistsInput(c.Request.URL.Query())
	if err != nil {
		httpapi.Fail(c, err)
		return
	}
	input.UserID = actor.UserID

	output, err := h.service.ListWishlists(c.Request.Context(), input)
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) getWishlist(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.GetWishlist(c.Request.Context(), wishlistusecase.GetWishlistInput{
		UserID:     actor.UserID,
		WishlistID: c.Param("wishlist_id"),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) addWishlistItem(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	var request addWishlistItemRequest
	if err := httpapi.DecodeJSON(c, &request); err != nil {
		httpapi.Fail(c, err)
		return
	}

	output, err := h.service.AddWishlistItem(c.Request.Context(), wishlistusecase.AddWishlistItemInput{
		UserID:     actor.UserID,
		WishlistID: c.Param("wishlist_id"),
		GiftID:     request.GiftID,
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusCreated, output)
}

func (h *Handler) removeWishlistItem(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.RemoveWishlistItem(c.Request.Context(), wishlistusecase.RemoveWishlistItemInput{
		UserID:     actor.UserID,
		WishlistID: c.Param("wishlist_id"),
		GiftID:     c.Param("gift_id"),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}

func (h *Handler) deleteWishlist(c *gin.Context) {
	actor, ok := authhttp.ActorFromContext(c)
	if !ok {
		httpapi.Fail(c, authhttp.UnauthorizedError())
		return
	}

	output, err := h.service.DeleteWishlist(c.Request.Context(), wishlistusecase.DeleteWishlistInput{
		UserID:     actor.UserID,
		WishlistID: c.Param("wishlist_id"),
	})
	if err != nil {
		httpapi.Fail(c, err)
		return
	}

	httpapi.Success(c, http.StatusOK, output)
}
