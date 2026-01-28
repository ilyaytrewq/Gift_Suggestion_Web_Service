package shared

import "slices"

type CategoryID string
type TagID string

func UniqTags(tags []TagID) []TagID {
	if len(tags) == 0 {
		return nil
	}
	uniq := append([]TagID(nil), tags...)
	slices.Sort(uniq)
	return slices.Compact(uniq)
}

type AgeLimit int8

const (
	AgeNone AgeLimit = 0
	Age12   AgeLimit = 12
	Age16   AgeLimit = 16
	Age18   AgeLimit = 18
)

func (limit AgeLimit) IsValid() bool {
	switch limit {
	case AgeNone, Age12, Age16, Age18:
		return true
	default:
		return false
	}
}

type Money int64

func (m Money) IsNonNegative() bool {
	return m >= 0
}

type Occasion string

const (
	HappyBirthday Occasion = "happy_birthday"
	NewYear       Occasion = "new_year"
	///....
)

type Relation string

const (
	Mother      Relation = "mother"
	Father      Relation = "father"
	Brother     Relation = "brother"
	Sister      Relation = "sister"
	GrandMother Relation = "grand_mother"
	GrandFather Relation = "grand_father"
)
