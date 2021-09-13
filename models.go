package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type User struct {
	Model

	MasterKey string     `gorm:"not null;uniqueIndex" json:"-"`
	Apps      StringList `gorm:"not null" json:"apps"`

	// associations
	Wallets []Wallet `json:"wallets"`
}

type Wallet struct {
	Model

	Name       string `gorm:"not null" json:"name"`
	InvoiceKey string `gorm:"not null" json:"invoicekey"`
	AdminKey   string `gorm:"not null" json:"adminkey"`

	// associations
	UserID        uint           `gorm:"index;not null" json:"userID"`
	Payments      []Payment      `json:"payments"`
	BalanceChecks []BalanceCheck `json:"balanceChecks"`
}

type Payment struct {
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	CheckingID    string     `gorm:"uniqueIndex;not null" json:"checkingID"`
	Pending       bool       `gorm:"not null" json:"pending"`
	Amount        int64      `gorm:"not null" json:"amount"`
	Fee           int64      `json:"fee"`
	Memo          string     `json:"memo"`
	Bolt11        string     `json:"bolt11"`
	Preimage      string     `json:"preimage"`
	Hash          string     `gorm:"index:hash_idx;not null" json:"hash"`
	Tag           string     `json:"tag"`
	Extra         JSONObject `json:"extra"`
	Webhook       string     `json:"webhook"`
	WebhookStatus int        `json:"webhookStatus"`

	// associations
	WalletID uint `gorm:"index;not null" json:"walletID"`
}

type BalanceCheck struct {
	WalletID uint   `gorm:"primaryKey" json:"walletID"`
	Service  string `gorm:"primaryKey" json:"service"`
	URL      string `json:"-"`
}

type AppDataItem struct {
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	App    string `gorm:"primaryKey;index:app_user_items_idx" json:"app"`
	UserID uint   `gorm:"primaryKey;index:app_user_items_idx" json:"userID"`
	Key    string `gorm:"primaryKey" json:"key"`

	Value JSONObject `gorm:"not null" json:"value"`
}

type StringList []string

func (sl StringList) Scan(src interface{}) error {
	if jstr, ok := src.(string); ok {
		return json.Unmarshal([]byte(jstr), sl)
	} else {
		return errors.New("value is not a string")
	}
}

func (sl StringList) Value() (driver.Value, error) {
	if j, err := json.Marshal(sl); err == nil {
		return string(j), nil
	} else {
		return nil, err
	}
}

type JSONObject map[string]interface{}

func (jo JSONObject) Scan(src interface{}) error {
	if jstr, ok := src.(string); ok {
		return json.Unmarshal([]byte(jstr), jo)
	} else {
		return errors.New("value is not a string")
	}
}

func (jo JSONObject) Value() (driver.Value, error) {
	if j, err := json.Marshal(jo); err == nil {
		return string(j), nil
	} else {
		return nil, err
	}
}
