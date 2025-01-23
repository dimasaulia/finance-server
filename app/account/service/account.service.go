package account_service

import av "finance/app/account/validation"

type IAccountService interface {
	CreateAccount(req av.AccountCreationRequest) (av.AccountCreationResponse, error)
}
