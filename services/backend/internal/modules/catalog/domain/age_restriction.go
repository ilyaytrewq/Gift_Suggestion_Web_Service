package domain

type AgeRestriction int16

const (
	AgeRestrictionNone AgeRestriction = 0
	AgeRestriction12   AgeRestriction = 12
	AgeRestriction16   AgeRestriction = 16
	AgeRestriction18   AgeRestriction = 18
)

func NewAgeRestriction(value int) (AgeRestriction, error) {
	switch value {
	case 0:
		return AgeRestrictionNone, nil
	case 12:
		return AgeRestriction12, nil
	case 16:
		return AgeRestriction16, nil
	case 18:
		return AgeRestriction18, nil
	default:
		return 0, ErrInvalidAgeRestriction
	}
}

func (r AgeRestriction) Int() int {
	return int(r)
}
