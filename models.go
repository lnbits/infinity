package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

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

type User struct {
	gorm.Model

	MasterKey string
	Apps      StringList

	// associations
	Wallets []Wallet
}

type Wallet struct {
	gorm.Model

	Name       string
	UserID     uint `gorm:"index"`
	InvoiceKey string
	AdminKey   string

	// associations
	Payments      []Payment
	BalanceChecks []BalanceCheck
}

type Payment struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	CheckingID    string `gorm:"uniqueIndex"`
	Pending       bool
	Amount        int64
	Fee           int64
	Memo          string
	Bolt11        string
	Preimage      string
	Hash          string `gorm:"uniqueIndex"`
	Tag           string
	Extra         JSONObject
	Webhook       string
	WebhoosStatus int

	// associations
	WalletID uint `gorm:"index"`
}

type BalanceCheck struct {
	WalletID uint   `gorm:"primaryKey"`
	Service  string `gorm:"primaryKey"`
	URL      string
}

type AppDataItem struct {
	CreatedAt time.Time
	UpdatedAt time.Time

	App    string `gorm:"primaryKey:index:app_user_items_idx"`
	UserID uint   `gorm:"primaryKey:index:app_user_items_idx"`
	Key    string `gorm:"primaryKey"`

	Value JSONObject
}
