package gorm

import (
	"time"

	"gorm.io/gorm"
)

// Model represents the base model structure for database entities
type Model struct {
	ID        string         `gorm:"column:id;primaryKey"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"columnt:deleted_at"`
}

// User represents a user in the system
type User struct {
	Model
	FirstName      string    `gorm:"column:first_name"`
	LastName       string    `gorm:"column:last_name"`
	Login          string    `gorm:"column:login;uniqueIndex"`
	Type           int       `gorm:"column:type"`
	PwdHash        string    `gorm:"column:pwd_hash"`
	FailedAttempts int       `gorm:"column:failed_attempts"`
	LockoutUntil   time.Time `gorm:"column:lockout_until"`
}

// TableName specifies the table name for User
func (u User) TableName() string {
	return "users"
}

// AddressMetadata contains additional information about an address
type AddressMetadata struct {
	Comment     string `json:"comment"`
	ServiceName string `json:"service_name"`
}

// Address represents an email address in the system
type Address struct {
	Model
	Type             int             `gorm:"column:type"`
	Email            string          `gorm:"column:email"`
	ForwardAddressID string          `gorm:"column:forward_address_id"`
	ForwardAddress   *Address        `gorm:"foreignKey:ForwardAddressID"`
	OwnerID          string          `gorm:"column:owner_id"`
	Owner            User            `gorm:"foreignKey:OwnerID"`
	Metadata         AddressMetadata `gorm:"serializer:json;index"`
}

// TableName specifies the table name for Address
func (a Address) TableName() string {
	return "addresses"
}

// Chain represents a chain of addresses
type Chain struct {
	Hash              string         `gorm:"column:hash;primaryKey"`
	FromAddressID     string         `gorm:"column:from_address_id"`
	FromAddress       Address        `gorm:"foreignKey:FromAddressID"`
	ToAddressID       string         `gorm:"column:to_address_id"`
	ToAddress         Address        `gorm:"foreignKey:ToAddressID"`
	OrigFromAddressID string         `gorm:"column:orig_from_address_id"`
	OrigFromAddress   Address        `gorm:"foreignKey:OrigFromAddressID"`
	OrigToAddressID   string         `gorm:"column:orig_to_address_id"`
	OrigToAddress     Address        `gorm:"foreignKey:OrigToAddressID"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"columnt:deleted_at"`
}

// TableName specifies the table name for Chain
func (c Chain) TableName() string {
	return "chains"
}

// ApiToken represents an API token for authentication
type ApiToken struct {
	Model
	Name        string    `gorm:"column:name"`
	TokenHash   string    `gorm:"column:token_hash"`
	Salt        string    `gorm:"column:salt"`
	Description string    `gorm:"column:description"`
	OwnerID     string    `gorm:"column:owner_id"`
	Owner       User      `gorm:"foreignKey:OwnerID"`
	Expiration  time.Time `gorm:"column:expiration"`
}

// TableName specifies the table name for ApiToken
func (t ApiToken) TableName() string {
	return "tokens"
}
