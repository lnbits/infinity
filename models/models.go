package models

import (
	"time"
)

type User struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	MasterKey string     `gorm:"not null;uniqueIndex" json:"-"`
	Apps      StringList `gorm:"not null" json:"apps"`

	// associations
	Wallets []Wallet `json:"wallets,omitempty"`
}

type Wallet struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	Name          string `gorm:"not null" json:"name"`
	InvoiceKey    string `gorm:"not null" json:"invoicekey"`
	AdminKey      string `gorm:"not null" json:"adminkey"`
	BalanceNotify string `json:"balanceNotify"`

	Balance    int64  `gorm:"->" json:"balance"`
	LNURLDrain string `gorm:"-" json:"drain"`

	// associations
	UserID        string         `gorm:"index;not null" json:"userID"`
	Payments      []Payment      `json:"payments,omitempty"`
	BalanceChecks []BalanceCheck `json:"balanceChecks,omitempty"`
	AppDataItems  []AppDataItem  `json:"appDataItems,omitempty"`
}

type Payment struct {
	CreatedAt time.Time `json:"date"`
	UpdatedAt time.Time `json:"-"`

	CheckingID    string     `gorm:"uniqueIndex;not null" json:"checkingID"`
	Pending       bool       `gorm:"not null" json:"pending"`
	Amount        int64      `gorm:"not null" json:"amount"`
	Fee           int64      `json:"fee"`
	Description   string     `json:"description"`
	Bolt11        string     `json:"bolt11"`
	Preimage      string     `json:"preimage"`
	Hash          string     `gorm:"index:hash_idx;not null" json:"hash"`
	Tag           string     `json:"tag"`
	Extra         JSONObject `json:"extra"`
	ItemKey       string     `json:"itemKey"`
	Webhook       string     `json:"webhook"`
	WebhookStatus int        `json:"webhookStatus"`

	// associations
	WalletID string `gorm:"index;not null" json:"walletID"`
}

type BalanceCheck struct {
	WalletID string `gorm:"primaryKey" json:"walletID"`
	Service  string `gorm:"primaryKey" json:"service"`
	URL      string `json:"-"`
}

type AppDataItem struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"-"`

	App      string `gorm:"primaryKey;index:app_user_items_idx,wallet_item_idx" json:"app"`
	WalletID string `gorm:"primaryKey;index:app_user_items_idx,wallet_item_idx" json:"walletID"`
	Model    string `gorm:"primaryKey;index:app_user_items_idx" json:"model"`
	Key      string `gorm:"primaryKey;index:wallet_item_idx" json:"key"`

	Value JSONObject `gorm:"not null" json:"value"`
}
