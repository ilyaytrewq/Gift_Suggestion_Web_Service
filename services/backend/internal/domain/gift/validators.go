package gift

import "github.com/google/uuid"

func (id GiftID) IsValid() bool {
	return isValidGiftID(string(id))
}

func isValidGiftID(id string) bool {
	return uuid.Validate(id) == nil
}
