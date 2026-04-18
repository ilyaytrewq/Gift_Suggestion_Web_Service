package idgen

import (
	"github.com/google/uuid"

	authdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/domain"
	catalogdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalog/domain"
	catalogimportdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/catalogimport/domain"
	userdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/user/domain"
	wishlistdomain "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/wishlist/domain"
)

type UUIDGenerator struct{}

func (UUIDGenerator) NewUserID() (userdomain.UserID, error) {
	return userdomain.NewUserID(uuid.NewString())
}

func (UUIDGenerator) NewSessionID() (authdomain.SessionID, error) {
	return authdomain.NewSessionID(uuid.NewString())
}

func (UUIDGenerator) NewPasswordResetTokenID() (authdomain.PasswordResetTokenID, error) {
	return authdomain.NewPasswordResetTokenID(uuid.NewString())
}

func (UUIDGenerator) NewGiftID() (catalogdomain.GiftID, error) {
	return catalogdomain.NewGiftID(uuid.NewString())
}

func (UUIDGenerator) NewImportJobID() (catalogimportdomain.ImportJobID, error) {
	return catalogimportdomain.NewImportJobID(uuid.NewString())
}

func (UUIDGenerator) NewImportErrorID() (catalogimportdomain.ImportErrorID, error) {
	return catalogimportdomain.NewImportErrorID(uuid.NewString())
}

func (UUIDGenerator) NewWishlistID() (wishlistdomain.WishlistID, error) {
	return wishlistdomain.NewWishlistID(uuid.NewString())
}

func (UUIDGenerator) NewWishlistItemID() (wishlistdomain.WishlistItemID, error) {
	return wishlistdomain.NewWishlistItemID(uuid.NewString())
}
