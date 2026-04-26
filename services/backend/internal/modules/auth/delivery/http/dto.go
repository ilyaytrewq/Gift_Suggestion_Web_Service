package http

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type passwordResetRequest struct {
	Email string `json:"email"`
}

type emailVerificationConfirmRequest struct {
	Token string `json:"token"`
}

type passwordResetConfirmRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}
