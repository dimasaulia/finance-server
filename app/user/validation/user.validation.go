package user

type UserRegistrationRequest struct {
	Username   string `json:"username"`
	Fullname   string `json:"fullname"`
	Password   string `json:"password"`
	Email      string `json:"email"`
	Provider   string `json:"provider"`
	ProviderId string `json:"provider_id"`
}

type ManualRegistrationRequest struct {
	Username string `json:"username" validate:"required"`
	Fullname string `json:"fullname" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Provider string `json:"provider" validate:"required"`
}

type GoogleRegistrationRequest struct {
	Username   string `json:"username" validate:"required"`
	Fullname   string `json:"fullname" validate:"required"`
	Password   string `json:"password"`
	Email      string `json:"email" validate:"required"`
	Provider   string `json:"provider" validate:"required"`
	ProviderId string `json:"provider_id" validate:"required"`
}

type UserLoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password"`
	Provider        string `json:"provider" validate:"required"`
	ProviderId      string `json:"provider_id"`
}

type ManualLoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password" validate:"required"`
	Provider        string `json:"provider" validate:"required"`
	ProviderId      string `json:"provider_id"`
}

type GoogleLoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" validate:"required"`
	Password        string `json:"password"`
	Provider        string `json:"provider" validate:"required"`
	ProviderId      string `json:"provider_id"`
}

type UserResponse struct {
	Username string  `json:"username"`
	Fullname string  `json:"fullname"`
	Email    string  `json:"email"`
	Role     string  `json:"role"`
	Token    *string `json:"token"`
}

type GoogleRedirectResponse struct {
	RedirectUrl string
	State       string
}

type OAuth2Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	IdToken     string `json:"id_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type GoogleUserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
}
