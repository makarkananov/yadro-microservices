package domain

type Role string

const (
	UNDEFINED Role = "undefined"
	USER      Role = "user"
	ADMIN     Role = "admin"
)

// User represents a user of the system.
type User struct {
	Username string
	Password string
	Role     Role
}

// NewUser creates a new user with the given username, password, and role.
func NewUser(username, password string, role Role) *User {
	return &User{
		Username: username,
		Password: password,
		Role:     role,
	}
}
