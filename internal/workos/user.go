package workos

// User contains user information from WorkOS.
type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}
