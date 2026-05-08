package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	vkintegrationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/domain"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/apperrors"
)

type Service struct {
	repo            ConnectionRepository
	userReader      UserReader
	tokenProtector  TokenProtector
	importer        InterestImporter
	connectionIDGen ConnectionIDGenerator
	featureEnabled  bool
	requestTimeout  time.Duration
	clock           Clock
}

func NewService(
	repo ConnectionRepository,
	userReader UserReader,
	tokenProtector TokenProtector,
	importer InterestImporter,
	connectionIDGen ConnectionIDGenerator,
	featureEnabled bool,
	requestTimeout time.Duration,
	clock Clock,
) (*Service, error) {
	switch {
	case repo == nil:
		return nil, ErrNilConnectionRepository
	case userReader == nil:
		return nil, ErrNilUserReader
	case tokenProtector == nil:
		return nil, ErrNilTokenProtector
	case importer == nil:
		return nil, ErrNilInterestImporter
	case connectionIDGen == nil:
		return nil, ErrNilConnectionIDGen
	case clock == nil:
		return nil, ErrNilClock
	case requestTimeout <= 0:
		return nil, errors.New("vk request timeout must be greater than zero")
	}

	return &Service{
		repo:            repo,
		userReader:      userReader,
		tokenProtector:  tokenProtector,
		importer:        importer,
		connectionIDGen: connectionIDGen,
		featureEnabled:  featureEnabled,
		requestTimeout:  requestTimeout,
		clock:           clock,
	}, nil
}

func (s *Service) Connect(ctx context.Context, input ConnectInput) (ConnectOutput, error) {
	if err := s.ensureFeatureEnabled(); err != nil {
		return ConnectOutput{}, err
	}

	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return ConnectOutput{}, err
	}

	metadata, err := vkintegrationdomain.NewIntegrationMetadata(input.Profile.ScreenName, input.Profile.ProfileURL)
	if err != nil {
		return ConnectOutput{}, mapConnectionValidationError(err)
	}

	if !input.Consent.Granted || input.Consent.ObtainedAt.IsZero() {
		return ConnectOutput{}, apperrors.New(
			apperrors.KindValidation,
			"invalid_vk_connection_payload",
			"vk connection payload is invalid",
		)
	}

	existing, _, err := s.loadConnection(ctx, userID)
	if err != nil {
		return ConnectOutput{}, err
	}

	if existing != nil && existing.IsConnected() && existing.ProviderUserID() != strings.TrimSpace(input.ProviderUserID) {
		return ConnectOutput{}, apperrors.New(
			apperrors.KindConflict,
			"vk_connection_already_exists",
			"vk connection is already linked to another VK account",
		)
	}

	tokenCiphertext, tokenExpiresAt, scopes, err := s.resolveCredential(existing, input.Credential)
	if err != nil {
		return ConnectOutput{}, err
	}

	now := s.clock.Now()
	if existing == nil {
		connectionID, err := s.connectionIDGen.NewVKConnectionID()
		if err != nil {
			return ConnectOutput{}, err
		}

		connection, err := vkintegrationdomain.NewConnection(
			connectionID,
			userID,
			input.ProviderUserID,
			input.Consent.Version,
			tokenCiphertext,
			tokenExpiresAt,
			scopes,
			metadata,
			now,
			input.Consent.ObtainedAt,
		)
		if err != nil {
			return ConnectOutput{}, mapConnectionValidationError(err)
		}

		if err := s.repo.Save(ctx, &connection); err != nil {
			return ConnectOutput{}, mapRepositoryError(err)
		}

		return ConnectOutput{
			Connection: newConnectionOutput(s.featureEnabled, &connection, nil),
		}, nil
	}

	profileMetadata := metadata
	if profileIsEmpty(input.Profile) {
		profileMetadata = existing.Metadata()
	}

	if err := existing.Reconnect(
		input.ProviderUserID,
		input.Consent.Version,
		tokenCiphertext,
		tokenExpiresAt,
		scopes,
		profileMetadata,
		now,
		input.Consent.ObtainedAt,
	); err != nil {
		return ConnectOutput{}, mapConnectionValidationError(err)
	}

	if err := s.repo.ReplaceImportedInterests(ctx, existing, nil); err != nil {
		return ConnectOutput{}, mapRepositoryError(err)
	}

	return ConnectOutput{
		Connection: newConnectionOutput(s.featureEnabled, existing, nil),
	}, nil
}

func (s *Service) GetCurrentConnection(ctx context.Context, input GetCurrentConnectionInput) (GetCurrentConnectionOutput, error) {
	if err := s.ensureFeatureEnabled(); err != nil {
		return GetCurrentConnectionOutput{}, err
	}

	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return GetCurrentConnectionOutput{}, err
	}

	connection, interests, err := s.loadConnection(ctx, userID)
	if err != nil {
		return GetCurrentConnectionOutput{}, err
	}

	return GetCurrentConnectionOutput{
		Connection: newConnectionOutput(s.featureEnabled, connection, interests),
	}, nil
}

func (s *Service) Disconnect(ctx context.Context, input DisconnectInput) (DisconnectOutput, error) {
	if err := s.ensureFeatureEnabled(); err != nil {
		return DisconnectOutput{}, err
	}

	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return DisconnectOutput{}, err
	}

	connection, _, err := s.loadConnection(ctx, userID)
	if err != nil {
		return DisconnectOutput{}, err
	}
	if connection == nil {
		return DisconnectOutput{
			Connection: newConnectionOutput(s.featureEnabled, nil, nil),
		}, nil
	}

	connection.Disconnect(s.clock.Now())
	if err := s.repo.ReplaceImportedInterests(ctx, connection, nil); err != nil {
		return DisconnectOutput{}, mapRepositoryError(err)
	}

	return DisconnectOutput{
		Connection: newConnectionOutput(s.featureEnabled, connection, nil),
	}, nil
}

func (s *Service) SyncInterests(ctx context.Context, input SyncInterestsInput) (SyncInterestsOutput, error) {
	if err := s.ensureFeatureEnabled(); err != nil {
		return SyncInterestsOutput{}, err
	}

	userID, err := s.ensureUserExists(ctx, input.UserID)
	if err != nil {
		return SyncInterestsOutput{}, err
	}

	connection, _, err := s.loadConnection(ctx, userID)
	if err != nil {
		return SyncInterestsOutput{}, err
	}
	if connection == nil || !connection.IsConnected() || !connection.HasTokenConfigured() {
		return SyncInterestsOutput{}, apperrors.New(
			apperrors.KindConflict,
			"vk_connection_not_ready",
			"vk connection is not ready for sync",
		)
	}
	if !connection.HasGrantedConsent() {
		return SyncInterestsOutput{}, apperrors.New(
			apperrors.KindConflict,
			"vk_consent_required",
			"vk consent is required before sync",
		)
	}
	if connection.TokenExpired(s.clock.Now()) {
		return SyncInterestsOutput{}, apperrors.New(
			apperrors.KindConflict,
			"vk_token_expired",
			"vk token metadata is expired",
		)
	}
	if !s.tokenProtector.Configured() {
		return SyncInterestsOutput{}, apperrors.New(
			apperrors.KindUnavailable,
			"vk_token_storage_unavailable",
			"vk token storage is unavailable",
		)
	}

	token, err := s.tokenProtector.Open(*connection.TokenCiphertext())
	if err != nil {
		return SyncInterestsOutput{}, mapTokenProtectionError(err)
	}

	callCtx := ctx
	cancel := func() {}
	if s.requestTimeout > 0 {
		callCtx, cancel = context.WithTimeout(ctx, s.requestTimeout)
	}
	defer cancel()

	result, err := s.importer.ImportInterests(callCtx, ImportInterestsRequest{
		ProviderUserID: connection.ProviderUserID(),
		AccessToken:    token,
		Scopes:         connection.Scopes(),
	})
	if err != nil {
		return SyncInterestsOutput{}, s.failSync(ctx, connection, mapImportFailure(err))
	}

	importedAt := s.clock.Now()
	interests, err := buildImportedInterests(result, importedAt)
	if err != nil {
		return SyncInterestsOutput{}, s.failSync(ctx, connection, apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_interest_import_invalid",
			"vk interest import result is invalid",
			err,
		))
	}

	connection.MarkSyncSucceeded(importedAt)
	if err := s.repo.ReplaceImportedInterests(ctx, connection, interests); err != nil {
		return SyncInterestsOutput{}, mapRepositoryError(err)
	}

	return SyncInterestsOutput{
		Connection: newConnectionOutput(s.featureEnabled, connection, interests),
	}, nil
}

func (s *Service) ensureFeatureEnabled() error {
	if s.featureEnabled {
		return nil
	}

	return apperrors.New(
		apperrors.KindUnavailable,
		"vk_integration_disabled",
		"vk integration is disabled",
	)
}

func (s *Service) ensureUserExists(ctx context.Context, rawUserID string) (userdomain.UserID, error) {
	userID, err := userdomain.NewUserID(rawUserID)
	if err != nil {
		return userdomain.UserID{}, apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_user_id",
			"user id is invalid",
			err,
		)
	}

	user, err := s.userReader.GetByID(ctx, userID)
	if err != nil {
		return userdomain.UserID{}, err
	}
	if user == nil {
		return userdomain.UserID{}, apperrors.New(
			apperrors.KindNotFound,
			"user_not_found",
			"user not found",
		)
	}

	return userID, nil
}

func (s *Service) loadConnection(
	ctx context.Context,
	userID userdomain.UserID,
) (*vkintegrationdomain.Connection, []vkintegrationdomain.ImportedInterest, error) {
	connection, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if connection == nil {
		return nil, nil, nil
	}

	interests, err := s.repo.ListImportedInterests(ctx, connection.ID())
	if err != nil {
		return nil, nil, err
	}

	return connection, interests, nil
}

func (s *Service) resolveCredential(
	existing *vkintegrationdomain.Connection,
	input CredentialInput,
) (*string, *time.Time, []string, error) {
	tokenInput := trimStringPtr(input.AccessToken)
	if tokenInput != nil {
		if !s.tokenProtector.Configured() {
			return nil, nil, nil, apperrors.New(
				apperrors.KindUnavailable,
				"vk_token_storage_not_configured",
				"vk token storage is not configured",
			)
		}

		sealed, err := s.tokenProtector.Seal(*tokenInput)
		if err != nil {
			return nil, nil, nil, mapTokenProtectionError(err)
		}

		expiresAt := cloneTimePtr(input.ExpiresAt)
		return &sealed, expiresAt, append([]string(nil), input.Scopes...), nil
	}

	if existing == nil {
		return nil, cloneTimePtr(input.ExpiresAt), append([]string(nil), input.Scopes...), nil
	}

	if input.ExpiresAt != nil || len(input.Scopes) > 0 {
		return existing.TokenCiphertext(), cloneTimePtr(input.ExpiresAt), append([]string(nil), input.Scopes...), nil
	}

	return existing.TokenCiphertext(), existing.TokenExpiresAt(), existing.Scopes(), nil
}

func (s *Service) failSync(ctx context.Context, connection *vkintegrationdomain.Connection, syncErr error) error {
	appErr := apperrors.From(syncErr)
	if err := connection.MarkSyncFailed(appErr.Code(), s.clock.Now()); err != nil {
		return err
	}
	if err := s.repo.Save(ctx, connection); err != nil {
		return mapRepositoryError(err)
	}

	return syncErr
}

func buildImportedInterests(
	result ImportInterestsResult,
	importedAt time.Time,
) ([]vkintegrationdomain.ImportedInterest, error) {
	if len(result.Interests) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{}, len(result.Interests))
	interests := make([]vkintegrationdomain.ImportedInterest, 0, len(result.Interests))
	for index, record := range result.Interests {
		position := record.Position
		if position <= 0 {
			position = index + 1
		}

		interest, err := vkintegrationdomain.NewImportedInterest(record.Name, record.SourceLabel, position, importedAt)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[interest.NormalizedValue()]; ok {
			continue
		}

		seen[interest.NormalizedValue()] = struct{}{}
		interests = append(interests, interest)
	}

	return interests, nil
}

func mapConnectionValidationError(err error) error {
	switch {
	case errors.Is(err, vkintegrationdomain.ErrProviderUserIDEmpty),
		errors.Is(err, vkintegrationdomain.ErrProviderUserIDTooLong),
		errors.Is(err, vkintegrationdomain.ErrConsentVersionEmpty),
		errors.Is(err, vkintegrationdomain.ErrConsentVersionTooLong),
		errors.Is(err, vkintegrationdomain.ErrScreenNameTooLong),
		errors.Is(err, vkintegrationdomain.ErrProfileURLTooLong),
		errors.Is(err, vkintegrationdomain.ErrInvalidProfileURL),
		errors.Is(err, vkintegrationdomain.ErrInvalidScope),
		errors.Is(err, vkintegrationdomain.ErrTokenCiphertextTooLong),
		errors.Is(err, vkintegrationdomain.ErrConnectedAtRequired),
		errors.Is(err, vkintegrationdomain.ErrConsentGrantedAtRequired):
		return apperrors.Wrap(
			apperrors.KindValidation,
			"invalid_vk_connection_payload",
			"vk connection payload is invalid",
			err,
		)
	default:
		return err
	}
}

func mapTokenProtectionError(err error) error {
	switch {
	case errors.Is(err, ErrTokenProtectionUnavailable):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_token_storage_unavailable",
			"vk token storage is unavailable",
			err,
		)
	case errors.Is(err, ErrTokenCiphertextCorrupted):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_token_storage_unavailable",
			"vk token storage is unavailable",
			err,
		)
	default:
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_token_storage_unavailable",
			"vk token storage is unavailable",
			err,
		)
	}
}

func mapImportFailure(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_interest_import_timeout",
			"vk interest import timed out",
			err,
		)
	case errors.Is(err, ErrVKTokenInvalid):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_token_invalid",
			"vk token is invalid or expired, please reconnect vk account",
			err,
		)
	case errors.Is(err, ErrVKRateLimited):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_rate_limited",
			"vk api rate limit exceeded, please retry later",
			err,
		)
	case errors.Is(err, ErrVKGroupsAccessDenied):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_groups_access_denied",
			"vk groups access is denied by privacy settings",
			err,
		)
	case errors.Is(err, ErrInterestImportNotImplemented), errors.Is(err, ErrInterestImportUnavailable):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_interest_import_unavailable",
			"vk interest import is unavailable",
			err,
		)
	case errors.Is(err, ErrInvalidInterestImportResult):
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_interest_import_invalid",
			"vk interest import result is invalid",
			err,
		)
	default:
		return apperrors.Wrap(
			apperrors.KindUnavailable,
			"vk_interest_import_unavailable",
			"vk interest import is unavailable",
			err,
		)
	}
}

func mapRepositoryError(err error) error {
	switch {
	case errors.Is(err, ErrVKUserAlreadyConnected):
		return apperrors.Wrap(
			apperrors.KindConflict,
			"vk_connection_already_exists",
			"vk connection is already linked to another VK account",
			err,
		)
	default:
		return err
	}
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	cloned := value.UTC()
	return &cloned
}

func profileIsEmpty(profile ProfileInput) bool {
	return trimStringPtr(profile.ScreenName) == nil && trimStringPtr(profile.ProfileURL) == nil
}
