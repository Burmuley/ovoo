package ports

import (
	"github.com/Burmuley/domovoi/internal/core/domain"
	"github.com/Burmuley/domovoi/internal/model"
	"io"
)

type EmailProcessor interface {
	ParseMessage(r io.Reader) error
	GetFromAddress() (string, error)
	GetToAddress() (string, error)
	SetFromAddress(addr string) error
	SetToAddress(addr string) error
	SetReplyToAddress(addr string) error
}

type AliasRepo interface {
	Create(alias model.AliasEmail) error
	Delete(alias string) error
	List() []model.AliasEmail
	Get(alias string) model.AliasEmail

	MatchAlias(alias string) (model.AliasEmail, error)
	MatchProtectedEmail(protected string) (model.AliasEmail, error)
}

type RAliasRepoReader interface {
	Create(alias domain.ReplyAlias) error
	Delete(alias domain.ReplyAlias) error
	MatchProtectedEmail(protected string) (domain.ReplyAlias, error)
	MatchProtectedEmailAlias(protected, alias string) (domain.ReplyAlias, error)
}
