package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

const (
	testUserUsecaseUserID   = "550e8400-e29b-41d4-a716-446655440000"
	testUserUsecaseOtherID  = "661e8400-e29b-41d4-a716-446655440001"
	testUserUsecaseEmail    = "user@example.com"
	testUserUsecaseOtherEmail = "other@example.com"
	testUserUsecasePassword = "ValidPass1!"
)

func TestServiceGetCurrentUserReturnsProfile(t *testing.T) {
	t.Parallel()

	user := mustUsecaseUser(t, testUserUsecaseUserID, testUserUsecaseEmail, testUserUsecasePassword)
	if err := user.UpdateDisplayName("Alice", time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("UpdateDisplayName() error = %v", err)
	}

	repo := newFakeProfileRepository()
	repo.registerUser(user)

	service, err := NewService(repo, fixedProfileClock{now: time.Now().UTC()})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	profile, err := service.GetCurrentUser(context.Background(), user.ID().String())
	if err != nil {
		t.Fatalf("GetCurrentUser() error = %v", err)
	}

	if profile.Email != testUserUsecaseEmail {
		t.Fatalf("GetCurrentUser() email = %q, want %q", profile.Email, testUserUsecaseEmail)
	}
	if profile.DisplayName != "Alice" {
		t.Fatalf("GetCurrentUser() display name = %q, want %q", profile.DisplayName, "Alice")
	}
}

func TestServiceUpdateProfileSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 16, 0, 0, 0, time.UTC)
	user := mustUsecaseUser(t, testUserUsecaseUserID, testUserUsecaseEmail, testUserUsecasePassword)

	repo := newFakeProfileRepository()
	repo.registerUser(user)

	service, err := NewService(repo, fixedProfileClock{now: now})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	profile, err := service.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      user.ID().String(),
		DisplayName: "  Alice  ",
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}

	if profile.DisplayName != "Alice" {
		t.Fatalf("UpdateProfile() display name = %q, want %q", profile.DisplayName, "Alice")
	}
	if len(repo.updatedUsers) != 1 {
		t.Fatalf("expected one UpdateProfile() call, got %d", len(repo.updatedUsers))
	}
	if !repo.updatedUsers[0].UpdatedAt().Equal(now) {
		t.Fatalf("updated user timestamp = %v, want %v", repo.updatedUsers[0].UpdatedAt(), now)
	}
}

func TestServiceUpdateProfileInvalidUserID(t *testing.T) {
	t.Parallel()

	service, err := NewService(newFakeProfileRepository(), fixedProfileClock{now: time.Now().UTC()})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      "not-a-uuid",
		DisplayName: "Alice",
	})
	if err == nil {
		t.Fatal("UpdateProfile() expected validation error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindValidation {
		t.Fatalf("UpdateProfile() error kind = %q, want %q", appErr.Kind(), apperrors.KindValidation)
	}
}

func TestServicePromoteUserToAdminSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	target := mustUsecaseUser(t, testUserUsecaseOtherID, testUserUsecaseOtherEmail, testUserUsecasePassword)

	repo := newFakeProfileRepository()
	repo.registerUser(target)

	service, err := NewService(repo, fixedProfileClock{now: now})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	profile, err := service.PromoteUserToAdmin(context.Background(), testUserUsecaseOtherEmail)
	if err != nil {
		t.Fatalf("PromoteUserToAdmin() error = %v", err)
	}
	if profile.Role != string(domain.UserRoleAdmin) {
		t.Fatalf("PromoteUserToAdmin() role = %q, want admin", profile.Role)
	}
	if len(repo.roleUpdatedUsers) != 1 {
		t.Fatalf("expected UpdateUserRole once, got %d", len(repo.roleUpdatedUsers))
	}
}

func TestServicePromoteUserToAdminIdempotentWhenAlreadyAdmin(t *testing.T) {
	t.Parallel()

	u := mustUsecaseUser(t, testUserUsecaseOtherID, testUserUsecaseOtherEmail, testUserUsecasePassword)
	if err := u.PromoteToAdmin(time.Date(2026, 4, 19, 9, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("PromoteToAdmin(domain) error = %v", err)
	}

	repo := newFakeProfileRepository()
	repo.registerUser(u)

	service, err := NewService(repo, fixedProfileClock{now: time.Now().UTC()})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	profile, err := service.PromoteUserToAdmin(context.Background(), testUserUsecaseOtherEmail)
	if err != nil {
		t.Fatalf("PromoteUserToAdmin() error = %v", err)
	}
	if profile.Role != string(domain.UserRoleAdmin) {
		t.Fatalf("role = %q, want admin", profile.Role)
	}
	if len(repo.roleUpdatedUsers) != 0 {
		t.Fatal("expected no UpdateUserRole when already admin")
	}
}

func TestServicePromoteUserToAdminNotFound(t *testing.T) {
	t.Parallel()

	service, err := NewService(newFakeProfileRepository(), fixedProfileClock{now: time.Now().UTC()})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.PromoteUserToAdmin(context.Background(), "nobody@example.com")
	if err == nil {
		t.Fatal("PromoteUserToAdmin() expected not found")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindNotFound {
		t.Fatalf("PromoteUserToAdmin() error kind = %q, want %q", appErr.Kind(), apperrors.KindNotFound)
	}
}

func TestServicePromoteUserToAdminInvalidEmail(t *testing.T) {
	t.Parallel()

	service, err := NewService(newFakeProfileRepository(), fixedProfileClock{now: time.Now().UTC()})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	_, err = service.PromoteUserToAdmin(context.Background(), "not-email")
	if err == nil {
		t.Fatal("PromoteUserToAdmin() expected validation error")
	}

	appErr := apperrors.From(err)
	if appErr.Kind() != apperrors.KindValidation {
		t.Fatalf("PromoteUserToAdmin() error kind = %q, want %q", appErr.Kind(), apperrors.KindValidation)
	}
}

type fakeProfileRepository struct {
	usersByID         map[string]*domain.User
	usersByEmail      map[string]*domain.User
	updatedUsers      []*domain.User
	roleUpdatedUsers  []*domain.User
}

func newFakeProfileRepository() *fakeProfileRepository {
	return &fakeProfileRepository{
		usersByID:    make(map[string]*domain.User),
		usersByEmail: make(map[string]*domain.User),
	}
}

func (r *fakeProfileRepository) Save(context.Context, *domain.User) error {
	return nil
}

func (r *fakeProfileRepository) GetByID(_ context.Context, id domain.UserID) (*domain.User, error) {
	return r.usersByID[id.String()], nil
}

func (r *fakeProfileRepository) registerUser(u *domain.User) {
	r.usersByID[u.ID().String()] = u
	r.usersByEmail[u.Email().String()] = u
}

func (r *fakeProfileRepository) GetByEmail(_ context.Context, em domain.Email) (*domain.User, error) {
	return r.usersByEmail[em.String()], nil
}

func (r *fakeProfileRepository) UpdateProfile(_ context.Context, user *domain.User) error {
	r.updatedUsers = append(r.updatedUsers, user)
	return nil
}

func (r *fakeProfileRepository) UpdateUserRole(_ context.Context, user *domain.User) error {
	r.roleUpdatedUsers = append(r.roleUpdatedUsers, user)
	return nil
}

func (r *fakeProfileRepository) MarkLastLogin(context.Context, domain.UserID, time.Time) error {
	return nil
}

type fixedProfileClock struct {
	now time.Time
}

func (c fixedProfileClock) Now() time.Time {
	return c.now.UTC()
}

func mustUsecaseUser(t *testing.T, id, email, password string) *domain.User {
	t.Helper()

	user, err := domain.NewUser(id, email, password, string(domain.UserRoleUser))
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	return &user
}
