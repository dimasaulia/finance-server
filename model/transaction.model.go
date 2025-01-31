package model

import (
	"database/sql"
	"errors"
	v "finance/app/transaction/validation"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TB_TRANSACTION string          = "TRANSACTION"
	Debit          TransactionType = "DEBIT"
	Credit         TransactionType = "CREDIT"
)

type Transaction struct {
	IdTransaction   int64           `gorm:"column:id_transaction;primaryKey;autoIncement"`
	TransactionType TransactionType `gorm:"column:transaction_type"`
	TransactionCode string          `gorm:"column:transaction_code"`
	Amount          float64         `gorm:"column:amount;type:decimal(15,2)"`
	BalanceBefore   float64         `gorm:"column:balance_before;type:decimal(15,2)"`
	BalanceAfter    float64         `gorm:"column:balance_after;type:decimal(15,2)"`
	CreatedAt       time.Time       `gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime"`
	Description     sql.NullString  `gorm:"column:description"`

	// Foreign Key
	IdTransactionGroup   int64            `gorm:"column:id_transaction_group;foreignKey:id_transaction_group;references:id_transaction_group"`
	TransactionGroup     TransactionGroup `gorm:"foreignKey:id_transaction_group;references:id_transaction_group"`
	IdUser               int64            `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	User                 User             `gorm:"column:id_user;foreignKey:id_user;references:id_user"`
	IdAccount            int64            `gorm:"column:id_account;foreignKey:id_account;references:id_account"`
	Account              Account          `gorm:"foreignKey:id_account;references:id_account"`
	IdRelatedTransaction sql.NullInt64    `gorm:"column:id_related_transaction;foreignKey:id_related_transaction;references:id_transaction"`
	RelatedTransaction   []Transaction    `gorm:"foreignKey:id_related_transaction;references:id_transaction"`
	SubTransaction       []SubTransaction `gorm:"column:id_transaction;foreignKey:id_transaction;references:id_transaction"`
}

func (t Transaction) ValidateTransactionType() error {
	switch TransactionType(strings.ToUpper(string(t.TransactionType))) {
	case Debit, Credit:
		return nil
	default:
		return errors.New("transaction type not allow")
	}
}

func (t *Transaction) CrateNewTransaction(tx *gorm.DB) error {
	err := t.TransactionGroup.AutoCreateTransactionGroup(tx)
	if err != nil {
		return err
	}

	// Ambil Data Counter
	var tgCounter TransactionCounter
	err = tx.Model(&TransactionCounter{}).Select("*").Where("id_transaction_group", t.TransactionGroup.IdTransactionGroup).First(&tgCounter).Error
	if err != nil {
		return fmt.Errorf("failed when get counter, %v", err.Error())
	}
	var now = time.Now()
	var loopCounterFinder bool = true
	var transactionInitial string
	if t.TransactionType == Debit {
		transactionInitial = "D"
	} else {
		transactionInitial = "C"
	}
	var transactionCode string = fmt.Sprintf("%s%s%03d%02d%02d%v", transactionInitial, tgCounter.Descirption, tgCounter.Counter, now.Day(), now.Month(), now.Year())
	for loopCounterFinder {
		var existingTransaction int64
		err := tx.Model(&Transaction{}).Select("*").Where("transaction_code", transactionCode).Count(&existingTransaction).Error

		if err != nil {
			loopCounterFinder = false
			return err
		}

		if existingTransaction == 0 {
			loopCounterFinder = false
		} else {
			tgCounter.Counter += 1
			transactionCode = fmt.Sprintf("%s%s%03d%02d%02d%v", transactionInitial, tgCounter.Descirption, tgCounter.Counter, now.Day(), now.Month(), now.Year())
		}
	}

	// Cek Account dan ambil amount
	var userAccount Account
	qAccount := tx.Model(&Account{}).Select("*").Where("id_account", t.IdAccount).Where("id_user", t.IdUser).First(&userAccount)
	if qAccount.Error != nil {
		return fmt.Errorf("failed when get account data. %v", qAccount.Error)
	}

	// DEBIT => SALDO BERKURANG; CREDIT => SALDO BERTAMBAH
	if t.TransactionType == Debit && userAccount.Balance < t.Amount {
		return errors.New("insufficient account balance for the requested debit transaction")
	}
	t.BalanceBefore = userAccount.Balance

	// Kurangi atau tambah amount, serta lakukan update amount pada tabel account
	if t.TransactionType == Debit {
		userAccount.Balance = userAccount.Balance - t.Amount
	}

	if t.TransactionType == Credit {
		userAccount.Balance = userAccount.Balance + t.Amount
	}

	t.IdTransactionGroup = t.TransactionGroup.IdTransactionGroup
	t.IdAccount = userAccount.IdAccount
	t.BalanceAfter = userAccount.Balance
	t.TransactionCode = transactionCode

	// Create transaction
	err = tx.Create(&t).Error
	if err != nil {
		return err
	}

	// Update Account, khususnya ammount
	err = tx.Save(&userAccount).Where("id_account", userAccount.IdAccount).Where("id_user", userAccount.IdUser).Error
	if err != nil {
		return err
	}

	// Update Counter
	tgCounter.Counter = tgCounter.Counter + 1
	err = tx.Save(&tgCounter).Where("id_transaction_counter", tgCounter.IdTransactionCounter).Where("id_transaction_group", t.TransactionGroup.IdTransactionGroup).Error
	if err != nil {
		return err
	}

	return nil
}

func (t *Transaction) NewTransactionResponse() *v.TransactionResponse {
	return &v.TransactionResponse{
		TransactionCode: t.TransactionCode,
		Amount:          t.Amount,
		BalanceBefore:   t.BalanceBefore,
		BalanceAfter:    t.BalanceAfter,
	}
}

func (t *Transaction) UpdateExistingTransaction(tx *gorm.DB, newIdAccountDestination *int64, adminFee sql.NullFloat64) (*[]v.TransactionResponse, error) {
	var resp []v.TransactionResponse
	var sourceAmount float64
	// Cek Transaction
	var userTransaction Transaction
	qTransaction := tx.Model(&Transaction{}).Select("*").Where("id_transaction", t.IdTransaction).Where("id_user", t.IdUser).First(&userTransaction)
	if qTransaction.Error != nil {
		return nil, fmt.Errorf("error when try to find your transaction. %v", qTransaction.Error)
	}
	// Cek Account
	var userAccount Account
	qAccount := tx.Model(&Account{}).Select("*").Where("id_account", userTransaction.IdAccount).Where("id_user", userTransaction.IdUser).First(&userAccount)
	if qAccount.Error != nil {
		return nil, fmt.Errorf("error when try to find your account. %v", qAccount.Error)
	}
	// Cek Apakah Account memiliki parent transaction
	if userTransaction.IdRelatedTransaction.Valid {
		return nil, fmt.Errorf("cannot delete this transaction because it has a parent transaction. To delete this transaction, you must first delete the parent transaction")
	}
	// Cek Apakah Terdapat transaksi yang lebih baru, jika ada maka transaksi ini tidak bisa di edit, dan user harus menghapus transaksi terbarunya
	var newerUserTransactionCount int64
	qNewerTransaction := tx.Model(&Transaction{}).Where("id_account", userAccount.IdAccount).Where("id_user", t.IdUser).Where("created_at > ?", userTransaction.CreatedAt).Count(&newerUserTransactionCount)
	if qNewerTransaction.Error != nil {
		return nil, qNewerTransaction.Error
	}
	if newerUserTransactionCount > 0 {
		return nil, fmt.Errorf("this transaction cannot be modified because a newer transaction exists. please delete the latest transaction before making changes")
	}

	err := t.TransactionGroup.AutoCreateTransactionGroup(tx)
	if err != nil {
		return nil, err
	}

	// Reset Nilai ammount pada account ke nilai awal
	// Debit Yang Awalnya mengurangi nilai, jika dalam proses reset maka proses debit akan menambah nilai amount
	// Credit Yang Awalnya menambah nilai, jika dalam proses reset maka proses credit akan mengurangi nilai amount
	if userTransaction.TransactionType == Debit {
		userAccount.Balance = userAccount.Balance + userTransaction.Amount
	}

	if userTransaction.TransactionType == Credit {
		userAccount.Balance = userAccount.Balance - userTransaction.Amount
	}

	sourceAmount = t.Amount
	if adminFee.Valid {
		sourceAmount = sourceAmount + adminFee.Float64
	}

	// Ubah nilai ammount ke nilai baru
	// DEBIT => SALDO BERKURANG; CREDIT => SALDO BERTAMBAH
	if t.TransactionType == Debit && userAccount.Balance < sourceAmount {
		return nil, errors.New("insufficient account balance for the requested debit transaction")
	}

	userTransaction.BalanceBefore = userAccount.Balance
	if t.TransactionType == Debit {
		userAccount.Balance = userAccount.Balance - sourceAmount
	}

	if t.TransactionType == Credit {
		userAccount.Balance = userAccount.Balance + sourceAmount
	}

	userTransaction.IdTransactionGroup = t.TransactionGroup.IdTransactionGroup
	userTransaction.TransactionType = t.TransactionType
	userTransaction.BalanceAfter = userAccount.Balance
	userTransaction.Amount = sourceAmount
	userTransaction.Description = t.Description

	// Update transaction record
	err = tx.Save(&userAccount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to save account update. %v", err)
	}

	err = tx.Save(&userTransaction).Error
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction update. %v", err)
	}

	resp = append(resp, *userTransaction.NewTransactionResponse())

	// Find Other Related Transaction
	var otherRelatedTransaction []Transaction
	qOtherTransaction := tx.Model(&Transaction{}).Select("*").Where("id_related_transaction", userTransaction.IdTransaction).Where("id_user", userTransaction.IdUser).Scan(&otherRelatedTransaction)
	if qOtherTransaction.Error != nil {
		return nil, fmt.Errorf("failed to get other related transaction. %v", qOtherTransaction.Error.Error())
	}

	if qOtherTransaction.RowsAffected > 0 && newIdAccountDestination == nil {
		return nil, errors.New("the transaction you are trying to update is linked to other transactions. you must specify a new destination account for this transaction")
	}

	// TODO: Refactor Related Transaction into seprate function
	for _, v := range otherRelatedTransaction {
		var relatedUserAccount, relatedUserAccount2 Account // relatedUserAccount2 digunakan ketika akun tujuan berbeda
		qRelatedAccount := tx.Model(&Account{}).Select("*").Where("id_account", v.IdAccount).Where("id_user", v.IdUser).First(&relatedUserAccount)
		if qRelatedAccount.Error != nil {
			return nil, fmt.Errorf("error when try to find your related account. %v", qRelatedAccount.Error)
		}

		if v.IdAccount != *newIdAccountDestination {
			qRelatedAccount2 := tx.Model(&Account{}).Select("*").Where("id_account", &newIdAccountDestination).Where("id_user", v.IdUser).First(&relatedUserAccount2)
			if qRelatedAccount2.Error != nil {
				return nil, fmt.Errorf("error when try to find your related account. %v", qRelatedAccount2.Error)
			}
		}

		// Jika user mencoba untuk melakukan reverse jenis transaksi
		// Jika user melakukan edit dan mengubah debit menjadi credit pada transaksi utama
		// Dan jika terdapat related transaction yang mengikat pada akun pertama
		// Maka batalkan operasi
		if t.TransactionType == Credit && v.IdRelatedTransaction.Valid {
			return nil, fmt.Errorf("transaction type cannot be changed because this transaction has linked sub-transactions")
		}

		// Cek Apakah Terdapat transaksi yang lebih baru, jika ada maka transaksi ini tidak bisa di edit, dan user harus menghapus transaksi terbarunya
		var newerRelatedUserTransactionCount int64
		qNewerRelatedTransaction := tx.Model(&Transaction{}).Where("id_account", relatedUserAccount.IdAccount).Where("id_user", relatedUserAccount.IdUser).Where("created_at > ?", v.CreatedAt).Count(&newerRelatedUserTransactionCount)
		if qNewerRelatedTransaction.Error != nil {
			return nil, qNewerRelatedTransaction.Error
		}
		if newerRelatedUserTransactionCount >= 1 {
			return nil, fmt.Errorf("this transaction cannot be modified because a related transaction has a newer transaction. please delete the latest transaction before making changes")
		}

		log.Infof("Related Account ID: %v\n", relatedUserAccount.IdAccount)
		log.Infof("Related Account 2 Name: %v\n", relatedUserAccount2.Name)

		// Reset Nilai akun utama
		if v.TransactionType == Debit {
			relatedUserAccount.Balance = relatedUserAccount.Balance + v.Amount
		}

		if v.TransactionType == Credit {
			relatedUserAccount.Balance = relatedUserAccount.Balance - v.Amount
		}

		// Ubah nilai ammount ke nilai baru
		// Jika permintaan adalah melakukan credit dari transaksi utama
		// Yang berarti adalah proses debit pada related transaction,
		// lakukan validasi terlebih dahulu apakah related account memiliki jumlah yang ingin di pindahkan
		if (t.TransactionType == Credit) && (t.Amount > relatedUserAccount.Balance) {
			return nil, errors.New("insufficient account balance for the requested related credit on related transaction")
		}

		// Mengatur balance before ketika akun sama atau berbeda
		if v.IdAccount == *newIdAccountDestination {
			v.BalanceBefore = relatedUserAccount.Balance
		} else {
			v.BalanceBefore = relatedUserAccount2.Balance
		}
		// Related transaction merupakan kebalikan transaksi utamanya
		// Jika transaksi utamana nya adalah debit, maka transaksi di related transaction merupakan credit
		// Sehingga jika transaksi utama adalah debit, maka related account akan mengalami peningkatan nilai
		if newIdAccountDestination != nil && t.TransactionType == Debit && v.IdAccount == *newIdAccountDestination {
			relatedUserAccount.Balance = relatedUserAccount.Balance + t.Amount
			v.TransactionType = Credit
		}
		if newIdAccountDestination != nil && t.TransactionType == Credit && v.IdAccount == *newIdAccountDestination {
			relatedUserAccount.Balance = relatedUserAccount.Balance - t.Amount
			v.TransactionType = Debit
		}

		// Mengatur balance akun baru menyesuaikan jumlah transaksi
		if newIdAccountDestination != nil && t.TransactionType == Debit && v.IdAccount != *newIdAccountDestination {
			relatedUserAccount2.Balance = relatedUserAccount2.Balance + t.Amount
			v.TransactionType = Credit
		}

		if newIdAccountDestination != nil && t.TransactionType == Credit && v.IdAccount != *newIdAccountDestination {
			relatedUserAccount2.Balance = relatedUserAccount2.Balance - t.Amount
			v.TransactionType = Debit
		}

		// Mengatur balance akun baru menyesuaikan jumlah transaksi
		if newIdAccountDestination != nil && v.IdAccount == *newIdAccountDestination {
			v.BalanceAfter = relatedUserAccount.Balance
		} else {
			v.BalanceAfter = relatedUserAccount2.Balance
		}

		v.IdTransactionGroup = t.TransactionGroup.IdTransactionGroup
		v.Amount = t.Amount
		v.Description = t.Description

		// Jika User malkukan proses edit dan mengganti akun tujuan,
		// Maka id account yang sudah ada akan di pindah ke id account yang baru
		// Dan akan ada transaksi baru
		if newIdAccountDestination != nil && *newIdAccountDestination != v.IdAccount {
			v.IdAccount = *newIdAccountDestination
			err := tx.Save(&relatedUserAccount2).Error
			if err != nil {
				return nil, fmt.Errorf("failed to save related account update. %v", err)
			}
		}

		// Update transaction record
		err = tx.Model(&Transaction{}).Where("id_transaction", v.IdTransaction).Where("id_user", v.IdUser).Updates(v).Error
		if err != nil {
			return nil, fmt.Errorf("failed to save related transaction update. %v", err)
		}

		err = tx.Save(&relatedUserAccount).Error
		if err != nil {
			return nil, fmt.Errorf("failed to save related account update. %v", err)
		}
		resp = append(resp, *v.NewTransactionResponse())
	}
	return &resp, nil
}

func (t *Transaction) DeleteTransaction(tx *gorm.DB) error {
	// Cek Transaction
	var userTransaction Transaction
	qTransaction := tx.Model(&Transaction{}).Select("*").Where("id_transaction", t.IdTransaction).Where("id_user", t.IdUser).First(&userTransaction)
	if qTransaction.Error != nil {
		return fmt.Errorf("error when try to find your transaction. %v", qTransaction.Error)
	}
	// Cek Account
	var userAccount Account
	qAccount := tx.Model(&Account{}).Select("*").Where("id_account", userTransaction.IdAccount).Where("id_user", userTransaction.IdUser).First(&userAccount)
	if qAccount.Error != nil {
		return fmt.Errorf("error when try to find your account. %v", qAccount.Error)
	}
	// Cek Apakah Account memiliki parent transaction
	if userTransaction.IdRelatedTransaction.Valid {
		return fmt.Errorf("cannot delete this transaction because it has a parent transaction. To delete this transaction, you must first delete the parent transaction")
	}
	// Cek Apakah Terdapat transaksi yang lebih baru, jika ada maka transaksi ini tidak bisa di edit, dan user harus menghapus transaksi terbarunya
	var newerUserTransactionCount int64
	qNewerTransaction := tx.Model(&Transaction{}).Where("id_account", userAccount.IdAccount).Where("id_user", t.IdUser).Where("created_at > ?", userTransaction.CreatedAt).Count(&newerUserTransactionCount)
	if qNewerTransaction.Error != nil {
		return qNewerTransaction.Error
	}
	if newerUserTransactionCount > 0 {
		return fmt.Errorf("this transaction cannot be deleted because a newer transaction exists. please delete the latest transaction before making changes")
	}

	// Reset Nilai ammount pada account ke nilai awal
	// Debit Yang Awalnya mengurangi nilai, jika dalam proses reset maka proses debit akan menambah nilai amount
	// Credit Yang Awalnya menambah nilai, jika dalam proses reset maka proses credit akan mengurangi nilai amount
	if userTransaction.TransactionType == Debit {
		userAccount.Balance = userAccount.Balance + userTransaction.Amount
	}

	if userTransaction.TransactionType == Credit {
		userAccount.Balance = userAccount.Balance - userTransaction.Amount
	}

	// Find Other Related Transaction
	var otherRelatedTransaction []Transaction
	qOtherTransaction := tx.Model(&Transaction{}).Select("*").Where("id_related_transaction", userTransaction.IdTransaction).Where("id_user", userTransaction.IdUser).Scan(&otherRelatedTransaction)
	if qOtherTransaction.Error != nil {
		return fmt.Errorf("failed to get other related transaction. %v", qOtherTransaction.Error.Error())
	}

	for _, v := range otherRelatedTransaction {
		var relatedUserAccount Account
		qRelatedAccount := tx.Model(&Account{}).Select("*").Where("id_account", v.IdAccount).Where("id_user", v.IdUser).First(&relatedUserAccount)
		if qRelatedAccount.Error != nil {
			return fmt.Errorf("error when try to find your related account. %v", qRelatedAccount.Error)
		}

		// Cek Apakah Terdapat transaksi yang lebih baru, jika ada maka transaksi ini tidak bisa di hapus, dan user harus menghapus transaksi terbarunya
		var newerRelatedUserTransactionCount int64
		qNewerRelatedTransaction := tx.Model(&Transaction{}).Where("id_account", relatedUserAccount.IdAccount).Where("id_user", relatedUserAccount.IdUser).Where("created_at > ?", v.CreatedAt).Count(&newerRelatedUserTransactionCount)
		if qNewerRelatedTransaction.Error != nil {
			return qNewerRelatedTransaction.Error
		}
		if newerRelatedUserTransactionCount >= 1 {
			return fmt.Errorf("this transaction cannot be deleted because a related transaction has a newer transaction. please delete the latest transaction before making changes")
		}

		// Reset Nilai
		if v.TransactionType == Debit {
			relatedUserAccount.Balance = relatedUserAccount.Balance + v.Amount
		}

		if v.TransactionType == Credit {
			relatedUserAccount.Balance = relatedUserAccount.Balance - v.Amount
		}

		err := tx.Where("id_transaction", v.IdTransaction).Where("id_user", v.IdUser).Delete(&Transaction{}).Error
		if err != nil {
			return fmt.Errorf("failed to delete related transaction. %v", err)
		}

		err = tx.Save(&relatedUserAccount).Error
		if err != nil {
			return fmt.Errorf("failed to save related account changes. %v", err)
		}
	}

	// Delete transaction record and update user account
	err := tx.Where("id_transaction", userTransaction.IdTransaction).Where("id_user", userTransaction.IdUser).Delete(&Transaction{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete transaction. %v", err)
	}

	err = tx.Save(&userAccount).Error
	if err != nil {
		return fmt.Errorf("failed to save account update. %v", err)
	}

	return nil
}
