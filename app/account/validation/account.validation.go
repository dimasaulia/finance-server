package account_validation

type AccountCreationRequest struct {
	IdUser  int64   `json:"id_user" validate:"required"`
	Name    string  `json:"name" validate:"required"`
	Balance float64 `json:"balance" validate:"required"`
	Type    string  `json:"type" validate:"required"`
}

type AccountCreationResponse struct {
	IdAccount int64   `json:"id_account"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
	Type      string  `json:"type"`
}

type AccountListRequest struct {
	Type   string `json:"type"`
	IdUser string `json:"id_user"`
}
