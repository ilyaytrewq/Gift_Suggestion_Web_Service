package usecase

import (
	"context"
	"testing"
	"time"

	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	recommendationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/recommendation/domain"
	trackingdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/tracking/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const testTrackingUserID = "550e8400-e29b-41d4-a716-446655445000"

func TestServiceTrackEventSuccessCardView(t *testing.T) {
	t.Parallel()

	gift := mustTrackingGift(t, "550e8400-e29b-41d4-a716-446655445010")
	giftID := gift.ID().String()
	surface := "catalog"
	position := 1
	service := mustTrackingService(t, trackingServiceDeps{
		repo:       &fakeTrackingRepository{},
		userReader: fakeTrackingUserReader{user: mustTrackingUser(t)},
		giftReader: fakeTrackingGiftReader{gift: &gift},
	})

	output, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID: testTrackingUserID,
		Type:   "card_view",
		GiftID: &giftID,
		Metadata: EventMetadataInput{
			Surface:  &surface,
			Position: &position,
		},
	})
	if err != nil {
		t.Fatalf("TrackEvent() error = %v", err)
	}

	if output.Event.Type != "card_view" {
		t.Fatalf("TrackEvent() type = %q, want %q", output.Event.Type, "card_view")
	}
	if output.Event.GiftID == nil || *output.Event.GiftID != gift.ID().String() {
		t.Fatalf("TrackEvent() gift_id = %v, want %q", output.Event.GiftID, gift.ID().String())
	}
	if output.Event.Duplicate {
		t.Fatal("TrackEvent() duplicate = true, want false")
	}
}

func TestServiceTrackEventSuccessWishlistAdd(t *testing.T) {
	t.Parallel()

	gift := mustTrackingGift(t, "550e8400-e29b-41d4-a716-446655445011")
	wishlist := mustTrackingWishlist(t, "550e8400-e29b-41d4-a716-446655445012", testTrackingUserID)
	giftID := gift.ID().String()
	wishlistID := wishlist.ID().String()
	service := mustTrackingService(t, trackingServiceDeps{
		repo:           &fakeTrackingRepository{},
		userReader:     fakeTrackingUserReader{user: mustTrackingUser(t)},
		giftReader:     fakeTrackingGiftReader{gift: &gift},
		wishlistReader: fakeTrackingWishlistReader{wishlist: &wishlist},
	})

	output, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID:     testTrackingUserID,
		Type:       "wishlist_add",
		GiftID:     &giftID,
		WishlistID: &wishlistID,
	})
	if err != nil {
		t.Fatalf("TrackEvent() error = %v", err)
	}

	if output.Event.WishlistID == nil || *output.Event.WishlistID != wishlist.ID().String() {
		t.Fatalf("TrackEvent() wishlist_id = %v, want %q", output.Event.WishlistID, wishlist.ID().String())
	}
}

func TestServiceTrackEventRejectsInvalidEventType(t *testing.T) {
	t.Parallel()

	service := mustTrackingService(t, trackingServiceDeps{
		repo:       &fakeTrackingRepository{},
		userReader: fakeTrackingUserReader{user: mustTrackingUser(t)},
	})

	_, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID: testTrackingUserID,
		Type:   "unknown",
	})
	if err == nil {
		t.Fatal("TrackEvent() expected validation error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "invalid_event_type" {
		t.Fatalf("TrackEvent() code = %q, want %q", appErr.Code(), "invalid_event_type")
	}
}

func TestServiceTrackEventRejectsMissingGiftForCardView(t *testing.T) {
	t.Parallel()

	service := mustTrackingService(t, trackingServiceDeps{
		repo:       &fakeTrackingRepository{},
		userReader: fakeTrackingUserReader{user: mustTrackingUser(t)},
	})

	_, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID: testTrackingUserID,
		Type:   "card_view",
	})
	if err == nil {
		t.Fatal("TrackEvent() expected validation error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "invalid_event_payload" {
		t.Fatalf("TrackEvent() code = %q, want %q", appErr.Code(), "invalid_event_payload")
	}
}

func TestServiceTrackEventRejectsForeignRecommendationRequest(t *testing.T) {
	t.Parallel()

	request := mustTrackingRecommendationRequest(t, "550e8400-e29b-41d4-a716-446655445013", "550e8400-e29b-41d4-a716-446655445099")
	requestID := request.ID().String()
	service := mustTrackingService(t, trackingServiceDeps{
		repo:                 &fakeTrackingRepository{},
		userReader:           fakeTrackingUserReader{user: mustTrackingUser(t)},
		recommendationReader: fakeTrackingRecommendationReader{request: &request},
	})

	_, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID:                  testTrackingUserID,
		Type:                    "recommendation_request",
		RecommendationRequestID: &requestID,
	})
	if err == nil {
		t.Fatal("TrackEvent() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "recommendation_request_not_found" {
		t.Fatalf("TrackEvent() code = %q, want %q", appErr.Code(), "recommendation_request_not_found")
	}
}

func TestServiceTrackEventRejectsUnknownGift(t *testing.T) {
	t.Parallel()

	giftID := "550e8400-e29b-41d4-a716-446655445014"
	service := mustTrackingService(t, trackingServiceDeps{
		repo:       &fakeTrackingRepository{},
		userReader: fakeTrackingUserReader{user: mustTrackingUser(t)},
		giftReader: fakeTrackingGiftReader{},
	})

	_, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID: testTrackingUserID,
		Type:   "card_view",
		GiftID: &giftID,
	})
	if err == nil {
		t.Fatal("TrackEvent() expected not found error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "gift_not_found" {
		t.Fatalf("TrackEvent() code = %q, want %q", appErr.Code(), "gift_not_found")
	}
}

func TestServiceTrackEventRejectsTooLongClientEventID(t *testing.T) {
	t.Parallel()

	gift := mustTrackingGift(t, "550e8400-e29b-41d4-a716-446655445015")
	giftID := gift.ID().String()
	tooLong := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn"
	service := mustTrackingService(t, trackingServiceDeps{
		repo:       &fakeTrackingRepository{},
		userReader: fakeTrackingUserReader{user: mustTrackingUser(t)},
		giftReader: fakeTrackingGiftReader{gift: &gift},
	})

	_, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID:        testTrackingUserID,
		Type:          "card_view",
		GiftID:        &giftID,
		ClientEventID: &tooLong,
	})
	if err == nil {
		t.Fatal("TrackEvent() expected validation error")
	}

	appErr := apperrors.From(err)
	if appErr.Code() != "invalid_client_event_id" {
		t.Fatalf("TrackEvent() code = %q, want %q", appErr.Code(), "invalid_client_event_id")
	}
}

func TestServiceTrackEventReturnsDuplicateOnSameClientEventID(t *testing.T) {
	t.Parallel()

	gift := mustTrackingGift(t, "550e8400-e29b-41d4-a716-446655445016")
	giftID := gift.ID().String()
	eventID := "client-event-1"
	service := mustTrackingService(t, trackingServiceDeps{
		repo:       &fakeTrackingRepository{duplicate: true},
		userReader: fakeTrackingUserReader{user: mustTrackingUser(t)},
		giftReader: fakeTrackingGiftReader{gift: &gift},
	})

	output, err := service.TrackEvent(context.Background(), TrackEventInput{
		UserID:        testTrackingUserID,
		Type:          "card_view",
		GiftID:        &giftID,
		ClientEventID: &eventID,
	})
	if err != nil {
		t.Fatalf("TrackEvent() error = %v", err)
	}

	if !output.Event.Duplicate {
		t.Fatal("TrackEvent() duplicate = false, want true")
	}
}

type trackingServiceDeps struct {
	repo                 *fakeTrackingRepository
	userReader           fakeTrackingUserReader
	giftReader           fakeTrackingGiftReader
	wishlistReader       fakeTrackingWishlistReader
	recommendationReader fakeTrackingRecommendationReader
}

func mustTrackingService(t *testing.T, deps trackingServiceDeps) *Service {
	t.Helper()

	repo := deps.repo
	if repo == nil {
		repo = &fakeTrackingRepository{}
	}

	service, err := NewService(
		repo,
		deps.userReader,
		deps.giftReader,
		deps.wishlistReader,
		deps.recommendationReader,
		fakeTrackingEventIDGenerator{id: "550e8400-e29b-41d4-a716-446655445001"},
		fakeTrackingClock{now: time.Date(2026, 4, 19, 14, 0, 0, 0, time.UTC)},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}

type fakeTrackingRepository struct {
	duplicate bool
	stored    *trackingdomain.Event
}

func (r *fakeTrackingRepository) CreateEvent(_ context.Context, event *trackingdomain.Event) (trackingdomain.Event, bool, error) {
	if !r.duplicate {
		cloned := *event
		r.stored = &cloned
		return cloned, true, nil
	}

	if r.stored == nil {
		cloned := *event
		r.stored = &cloned
	}

	return *r.stored, false, nil
}

type fakeTrackingUserReader struct {
	user *userdomain.User
}

func (r fakeTrackingUserReader) GetByID(context.Context, userdomain.UserID) (*userdomain.User, error) {
	return r.user, nil
}

type fakeTrackingGiftReader struct {
	gift *catalogdomain.Gift
}

func (r fakeTrackingGiftReader) GetGift(context.Context, catalogdomain.GiftID) (*catalogdomain.Gift, error) {
	return r.gift, nil
}

type fakeTrackingWishlistReader struct {
	wishlist *wishlistdomain.Wishlist
}

func (r fakeTrackingWishlistReader) GetWishlistByID(context.Context, wishlistdomain.WishlistID) (*wishlistdomain.Wishlist, error) {
	return r.wishlist, nil
}

type fakeTrackingRecommendationReader struct {
	request *recommendationdomain.RecommendationRequest
}

func (r fakeTrackingRecommendationReader) GetRequest(context.Context, recommendationdomain.RequestID) (*recommendationdomain.RecommendationRequest, error) {
	return r.request, nil
}

type fakeTrackingEventIDGenerator struct {
	id string
}

func (g fakeTrackingEventIDGenerator) NewTrackingEventID() (trackingdomain.EventID, error) {
	return trackingdomain.NewEventID(g.id)
}

type fakeTrackingClock struct {
	now time.Time
}

func (c fakeTrackingClock) Now() time.Time {
	return c.now
}

func mustTrackingUser(t *testing.T) *userdomain.User {
	t.Helper()

	user, err := userdomain.RestoreUser(
		testTrackingUserID,
		"tracking@example.com",
		"$2a$10$4s1z0N9H8vK6L2S2Qz4B6OYX3j6Nn3kLBXjg5eQof8RP4eAbOeX6C",
		"user",
		"Tracker",
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC),
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("RestoreUser() error = %v", err)
	}

	return &user
}

func mustTrackingGift(t *testing.T, id string) catalogdomain.Gift {
	t.Helper()

	categoryID := "550e8400-e29b-41d4-a716-446655445100"
	categoryName := "Books"
	gift, err := catalogdomain.RestoreGift(
		id,
		&categoryID,
		&categoryName,
		"Tracking Gift",
		"Gift for tracking",
		"10.00",
		"https://example.com/"+id,
		nil,
		nil,
		time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreGift() error = %v", err)
	}

	return gift
}

func mustTrackingWishlist(t *testing.T, id, userID string) wishlistdomain.Wishlist {
	t.Helper()

	wishlist, err := wishlistdomain.RestoreWishlist(
		id,
		userID,
		"My Wishlist",
		time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreWishlist() error = %v", err)
	}

	return wishlist
}

func mustTrackingRecommendationRequest(t *testing.T, id, userID string) recommendationdomain.RecommendationRequest {
	t.Helper()

	questionnaire, err := recommendationdomain.NewQuestionnaire("", "", nil, nil, "100.00", nil, nil, 5, true)
	if err != nil {
		t.Fatalf("NewQuestionnaire() error = %v", err)
	}

	finishedAt := time.Date(2026, 4, 19, 13, 0, 2, 0, time.UTC)

	request, err := recommendationdomain.RestoreRecommendationRequest(
		id,
		&userID,
		questionnaire,
		"completed",
		"ml",
		10,
		10,
		5,
		0,
		nil,
		nil,
		nil,
		time.Date(2026, 4, 19, 13, 0, 0, 0, time.UTC),
		&finishedAt,
		time.Date(2026, 4, 19, 13, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 19, 13, 0, 2, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("RestoreRecommendationRequest() error = %v", err)
	}

	return request
}
