package account_validation

type AccountCreationRequest struct {
	IdUser  int64   `json:"id_user" validate:"required"`
	Name    string  `json:"name" validate:"required"`
	Balance float64 `json:"balance" validate:"min=0"`
	Type    string  `json:"type" validate:"required"`
}

type AccountCreationResponse struct {
	IdAccount int64   `json:"id_account"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
	Type      string  `json:"type"`
}

type AccountListRequest struct {
	Type      string `json:"type"`
	IdAccount string `json:"id_account"`
	IdUser    string `json:"id_user"`
}

type AccountUpdateRequest struct {
	IdUser    int64   `json:"id_user" validate:"required"`
	IdAccount int64   `json:"id_account" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Balance   float64 `json:"balance" validate:"min=0"`
	Type      string  `json:"type" validate:"required"`
}
