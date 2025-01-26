package account_service

import (
	av "finance/app/account/validation"
	r "finance/utility/response"
)

type IAccountService interface {
	CreateAccount(req av.AccountCreationRequest) (av.AccountCreationResponse, error)
	UserAccountList(filter *r.StandarGetRequest, data *av.AccountListRequest) ([]av.AccountCreationResponse, error)
	DeleteAccountList(id_account string, id_user string) (int64, error)
	UpdateAccount(av.AccountUpdateRequest) (av.AccountCreationResponse, error)
}
