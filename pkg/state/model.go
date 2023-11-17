package state

import (
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Model) BeforeCreate(db *gorm.DB) error {
	m.ID = ulid.Make().String()
	return nil
}

type ProtectedAddress struct {
	Model
	Email  string `gorm:"uniqueIndex"`
	Active bool
}

type Alias struct {
	Model
	ProtectedAddressID string
	Comment            string
	ServiceName        string
	Email              string           `gorm:"uniqueIndex"`
	ProtectedAddress   ProtectedAddress `gorm:"foreignKey:ProtectedAddressID"`
	Active             bool
}

type ReplyAlias struct {
	Model
	Email    string `gorm:"email"`
	AliasID  string `gorm:"alias_id"`
	SenderID string `gorm:"sender_email_id"`
	Alias    Alias  `gorm:"foreignKey:AliasID"`
	Sender   Sender `gorm:"foreignKey:SenderID"`
	Active   bool
}

type Sender struct {
	Model
	Email string `gorm:"email"`
}
