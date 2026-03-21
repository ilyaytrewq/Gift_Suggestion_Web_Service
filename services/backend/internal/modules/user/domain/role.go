package domain

type Role string

const (
	UserRoleAdmin Role = "admin"
	UserRoleUser  Role = "user"
)

func newRole(role string) (Role, error) {
	if isBlank(role) {
		return "", ErrRoleEmpty
	}

	switch role {
	case string(UserRoleAdmin):
		return UserRoleAdmin, nil
	case string(UserRoleUser):
		return UserRoleUser, nil
	default:
		return "", ErrInvalidRole
	}
}
