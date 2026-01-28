package shared

func (id CategoryID) IsValid() bool {
	return isValidCategoryID(string(id))
}

func isValidCategoryID(id string) bool {
	return id != ""
}

func (id TagID) IsValid() bool {
	return isValidTagID(string(id))
}

func isValidTagID(id string) bool {
	return id != ""
}
