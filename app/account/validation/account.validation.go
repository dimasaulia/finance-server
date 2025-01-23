package account_validation

type AccountCreationRequest struct {
	Name    string `json:"name" validate:"required"`
	Balance string `json:"balance" validate:"required"`
	Type    string `json:"type" validate:"required"`
}

type AccountCreationResponse struct {
	IdAccount string `json:"id_account"`
	Name      string `json:"name"`
	Balance   string `json:"balance"`
	Type      string `json:"type"`
}
