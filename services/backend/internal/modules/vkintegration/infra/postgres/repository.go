package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	vkintegrationdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/domain"
	vkintegrationusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/vkintegration/usecase"
)

type Repository struct {
	db *sql.DB
}

type metadataSnapshot struct {
	ScreenName *string `json:"screen_name,omitempty"`
	ProfileURL *string `json:"profile_url,omitempty"`
}

type rowScanner interface {
	Scan(dest ...any) error
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByUserID(ctx context.Context, userID userdomain.UserID) (*vkintegrationdomain.Connection, error) {
	const query = `
		SELECT
			id,
			user_id,
			vk_user_id,
			connection_state,
			consent_state,
			consent_version,
			token_ciphertext,
			token_expires_at,
			scopes,
			integration_metadata,
			last_sync_state,
			last_sync_error_code,
			last_synced_at,
			connected_at,
			disconnected_at,
			consent_granted_at,
			consent_revoked_at,
			created_at,
			updated_at
		FROM vk_connections
		WHERE user_id = $1
	`

	connection, err := scanConnection(r.db.QueryRowContext(ctx, query, userID.String()))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &connection, nil
}

func (r *Repository) Save(ctx context.Context, connection *vkintegrationdomain.Connection) error {
	return upsertConnection(ctx, r.db, connection)
}

func (r *Repository) ListImportedInterests(
	ctx context.Context,
	connectionID vkintegrationdomain.ConnectionID,
) (interests []vkintegrationdomain.ImportedInterest, err error) {
	const query = `
		SELECT raw_value, normalized_value, source_label, position, imported_at
		FROM vk_imported_interests
		WHERE connection_id = $1
		ORDER BY position ASC, normalized_value ASC
	`

	rows, err := r.db.QueryContext(ctx, query, connectionID.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	interests = make([]vkintegrationdomain.ImportedInterest, 0)
	for rows.Next() {
		var (
			rawValue        string
			normalizedValue string
			sourceLabel     string
			position        int
			importedAt      sql.NullTime
		)

		if err := rows.Scan(&rawValue, &normalizedValue, &sourceLabel, &position, &importedAt); err != nil {
			return nil, err
		}

		interest, err := vkintegrationdomain.RestoreImportedInterest(
			rawValue,
			normalizedValue,
			sourceLabel,
			position,
			importedAt.Time,
		)
		if err != nil {
			return nil, err
		}

		interests = append(interests, interest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return interests, nil
}

func (r *Repository) ReplaceImportedInterests(
	ctx context.Context,
	connection *vkintegrationdomain.Connection,
	interests []vkintegrationdomain.ImportedInterest,
) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	if err = upsertConnection(ctx, tx, connection); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM vk_imported_interests WHERE connection_id = $1`, connection.ID().String()); err != nil {
		return err
	}

	const insertQuery = `
		INSERT INTO vk_imported_interests (
			connection_id,
			raw_value,
			normalized_value,
			source_label,
			position,
			imported_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, interest := range interests {
		if _, err = tx.ExecContext(
			ctx,
			insertQuery,
			connection.ID().String(),
			interest.RawValue(),
			interest.NormalizedValue(),
			interest.SourceLabel(),
			interest.Position(),
			interest.ImportedAt(),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func upsertConnection(ctx context.Context, exec execer, connection *vkintegrationdomain.Connection) error {
	metadataJSON, err := marshalMetadata(connection.Metadata())
	if err != nil {
		return err
	}

	scopesJSON, err := json.Marshal(connection.Scopes())
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO vk_connections (
			id,
			user_id,
			vk_user_id,
			connection_state,
			consent_state,
			consent_version,
			token_ciphertext,
			token_expires_at,
			scopes,
			integration_metadata,
			last_sync_state,
			last_sync_error_code,
			last_synced_at,
			connected_at,
			disconnected_at,
			consent_granted_at,
			consent_revoked_at,
			created_at,
			updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19
		)
		ON CONFLICT (user_id) DO UPDATE SET
			vk_user_id = EXCLUDED.vk_user_id,
			connection_state = EXCLUDED.connection_state,
			consent_state = EXCLUDED.consent_state,
			consent_version = EXCLUDED.consent_version,
			token_ciphertext = EXCLUDED.token_ciphertext,
			token_expires_at = EXCLUDED.token_expires_at,
			scopes = EXCLUDED.scopes,
			integration_metadata = EXCLUDED.integration_metadata,
			last_sync_state = EXCLUDED.last_sync_state,
			last_sync_error_code = EXCLUDED.last_sync_error_code,
			last_synced_at = EXCLUDED.last_synced_at,
			connected_at = EXCLUDED.connected_at,
			disconnected_at = EXCLUDED.disconnected_at,
			consent_granted_at = EXCLUDED.consent_granted_at,
			consent_revoked_at = EXCLUDED.consent_revoked_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err = exec.ExecContext(
		ctx,
		query,
		connection.ID().String(),
		connection.UserID().String(),
		connection.ProviderUserID(),
		string(connection.ConnectionState()),
		string(connection.ConsentState()),
		connection.ConsentVersion(),
		nullString(connection.TokenCiphertext()),
		nullTime(connection.TokenExpiresAt()),
		scopesJSON,
		metadataJSON,
		string(connection.LastSyncState()),
		nullString(connection.LastSyncErrorCode()),
		nullTime(connection.LastSyncedAt()),
		nullTime(connection.ConnectedAt()),
		nullTime(connection.DisconnectedAt()),
		nullTime(connection.ConsentGrantedAt()),
		nullTime(connection.ConsentRevokedAt()),
		connection.CreatedAt(),
		connection.UpdatedAt(),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return vkintegrationusecase.ErrVKUserAlreadyConnected
		}

		return err
	}

	return nil
}

func scanConnection(scanner rowScanner) (vkintegrationdomain.Connection, error) {
	var (
		id                string
		userID            string
		vkUserID          string
		connectionState   string
		consentState      string
		consentVersion    string
		tokenCiphertext   sql.NullString
		tokenExpiresAt    sql.NullTime
		scopesRaw         []byte
		metadataRaw       []byte
		lastSyncState     string
		lastSyncErrorCode sql.NullString
		lastSyncedAt      sql.NullTime
		connectedAt       sql.NullTime
		disconnectedAt    sql.NullTime
		consentGrantedAt  sql.NullTime
		consentRevokedAt  sql.NullTime
		createdAt         sql.NullTime
		updatedAt         sql.NullTime
	)

	if err := scanner.Scan(
		&id,
		&userID,
		&vkUserID,
		&connectionState,
		&consentState,
		&consentVersion,
		&tokenCiphertext,
		&tokenExpiresAt,
		&scopesRaw,
		&metadataRaw,
		&lastSyncState,
		&lastSyncErrorCode,
		&lastSyncedAt,
		&connectedAt,
		&disconnectedAt,
		&consentGrantedAt,
		&consentRevokedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return vkintegrationdomain.Connection{}, err
	}

	metadata, err := unmarshalMetadata(metadataRaw)
	if err != nil {
		return vkintegrationdomain.Connection{}, err
	}
	scopes, err := unmarshalScopes(scopesRaw)
	if err != nil {
		return vkintegrationdomain.Connection{}, err
	}

	return vkintegrationdomain.RestoreConnection(
		id,
		userID,
		vkUserID,
		connectionState,
		consentState,
		consentVersion,
		stringPtrFromNull(tokenCiphertext),
		timePtrFromNull(tokenExpiresAt),
		scopes,
		metadata,
		lastSyncState,
		stringPtrFromNull(lastSyncErrorCode),
		timePtrFromNull(lastSyncedAt),
		timePtrFromNull(connectedAt),
		timePtrFromNull(disconnectedAt),
		timePtrFromNull(consentGrantedAt),
		timePtrFromNull(consentRevokedAt),
		createdAt.Time,
		updatedAt.Time,
	)
}

func marshalMetadata(metadata vkintegrationdomain.IntegrationMetadata) ([]byte, error) {
	return json.Marshal(metadataSnapshot{
		ScreenName: metadata.ScreenName(),
		ProfileURL: metadata.ProfileURL(),
	})
}

func unmarshalMetadata(payload []byte) (vkintegrationdomain.IntegrationMetadata, error) {
	if len(payload) == 0 {
		return vkintegrationdomain.NewIntegrationMetadata(nil, nil)
	}

	var snapshot metadataSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return vkintegrationdomain.IntegrationMetadata{}, err
	}

	return vkintegrationdomain.NewIntegrationMetadata(snapshot.ScreenName, snapshot.ProfileURL)
}

func unmarshalScopes(payload []byte) ([]string, error) {
	if len(payload) == 0 {
		return nil, nil
	}

	var scopes []string
	if err := json.Unmarshal(payload, &scopes); err != nil {
		return nil, err
	}

	return scopes, nil
}

func nullString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return value.UTC()
}

func stringPtrFromNull(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	current := value.String
	return &current
}

func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	current := value.Time.UTC()
	return &current
}
