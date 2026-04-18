package http

import "encoding/json"

type nullableStringField struct {
	Set   bool
	Value *string
}

func (f *nullableStringField) UnmarshalJSON(data []byte) error {
	f.Set = true
	if string(data) == "null" {
		f.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	f.Value = &value

	return nil
}

type updateProfileRequest struct {
	Profile struct {
		DisplayName nullableStringField `json:"display_name"`
	} `json:"profile"`
}
