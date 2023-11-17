package model

type ProtectedEmail struct {
	ID    string `json:"-" gorm:"id"`
	Email string `json:"email" gorm:"email"`
}

type AliasEmail struct {
	ID               string `json:"id" gorm:"id"`
	Email            string `json:"alias" gorm:"email"`
	ProtectedEmailID string `json:"protected_email_id" gorm:"protected_email_id"`
	Comment          string `json:"comment" gorm:"comment"`
	ServiceName      string `json:"service_name" gorm:"service_name"`
}

type ReplyAliasEmail struct {
	ID           string `json:"-" gorm:"id"`
	Email        string `json:"email" gorm:"email"`
	AliasEmailID string `json:"alias_email_id" gorm:"alias_email_id"`
}

type SenderEmail struct {
	ID    string `json:"-" gorm:"id"`
	Email string `json:"email" gorm:"email"`
}
