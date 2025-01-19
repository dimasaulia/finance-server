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

type UserResponse struct {
	Username string `json:"username"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
