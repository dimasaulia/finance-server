package transaction_validation

import (
	"database/sql"
	"time"
)

type NewTransactionRequest struct {
	IdUser               int64    `json:"id_user" validate:"required,number"`
	IdAccount            int64    `json:"id_account" validate:"required,number"`
	TransactionType      string   `json:"transaction_type" validate:"required,alphaspace"`
	TransactionGroup     string   `json:"transaction_group" validate:"required,alphaspace"`
	IdAccountDestination *int64   `json:"id_transaction_destination" validate:"omitempty,number"`
	Amount               float64  `json:"amount" validate:"required,number,gte=0"`
	Description          *string  `json:"description" validate:"required"`
	AdminFee             *float64 `json:"admin_fee" validate:"omitempty,number,gte=0"`
}

type UpdateTransactionRequest struct {
	IdTransaction        int64    `json:"id_transaction" validate:"required,number"`
	IdUser               int64    `json:"id_user" validate:"required,number"`
	TransactionType      string   `json:"transaction_type" validate:"required,alphaspace"`
	TransactionGroup     string   `json:"transaction_group" validate:"required,alphaspace"`
	IdAccountDestination *int64   `json:"id_transaction_destination" validate:"omitempty,number"`
	Amount               float64  `json:"amount" validate:"required,number,gte=0"`
	Description          *string  `json:"description" validate:"required"`
	AdminFee             *float64 `json:"admin_fee" validate:"omitempty,number,gte=0"`
}

type TransactionResponse struct {
	TransactionCode string  `json:"transaction_code"`
	Amount          float64 `json:"amount"`
	BalanceBefore   float64 `json:"balance_before"`
	BalanceAfter    float64 `json:"balance_after"`
}

type DeleteTransactionRequest struct {
	IdTransaction int64 `json:"id_transaction" validate:"required,number"`
	IdUser        int64 `json:"id_user" validate:"required,number"`
}

type TransactionData struct {
	IdTransaction           int64         `json:"id_transaction"`
	TransactionCode         string        `json:"transaction_code"`
	TransactionType         string        `json:"transaction_type"`
	Amount                  float64       `json:"amount"`
	BalanceBefore           float64       `json:"balance_before"`
	BalanceAfter            float64       `json:"balance_after"`
	Description             string        `json:"description"`
	CreatedAt               time.Time     `json:"created_at"`
	IsHaveParentTransaction int8          `json:"is_have_parent_transaction"`
	IdRelatedTransaction    sql.NullInt64 `json:"id_related_transaction"`
	IdTransactionGroup      int64         `json:"id_transaction_group"`
	TransactionName         string        `json:"transaction_name"`
	IdAccount               int64         `json:"id_account"`
	AccountName             string        `json:"account_name"`
}

type UserTransactionDetailRequest struct {
	IdUser    *string `json:"id_user" validate:"required"`
	IdAccount *string `json:"id_account" validate:"required"`
}

type NewSubTransactionRequest struct {
	IdUser                   int64    `json:"id_user" validate:"required,number"`
	IdTransaction            int64    `json:"id_transaction" validate:"required,number"`
	TransactionType          string   `json:"transaction_type" validate:"required,alphaspace"`
	TransactionGroup         string   `json:"transaction_group" validate:"required,alphaspace"`
	IdTransactionDestination *int64   `json:"id_transaction_destination" validate:"omitempty,number"`
	Amount                   float64  `json:"amount" validate:"required,number,gte=0"`
	Description              *string  `json:"description" validate:"required"`
	AdminFee                 *float64 `json:"admin_fee" validate:"omitempty,number,gte=0"`
}
