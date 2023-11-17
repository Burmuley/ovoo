package services

import (
	"errors"
	"fmt"
	"github.com/Burmuley/domovoi/internal/core/ports"
	"io"
	"net/mail"
)

var (
	ErrEmptyMessage  = errors.New("no message available to parse")
	ErrFieldNotFound = errors.New("requested field not found in the message header")
	ErrEmptyHeader   = errors.New("header is nil")
)

type EmailProcessor struct {
	aliasRepo ports.AliasRepo
	emailBody *mail.Message
}

func NewEmailProcessor(aliasRepo ports.AliasRepo) *EmailProcessor {
	return &EmailProcessor{
		aliasRepo: aliasRepo,
	}
}

func (e *EmailProcessor) GetFromAddress() (string, error) {
	if e.emailBody == nil {
		return "", ErrEmptyMessage
	}

	return getFieldFromHeader(e.emailBody.Header, "From")
}

func (e *EmailProcessor) GetToAddress() (string, error) {
	if e.emailBody == nil {
		return "", ErrEmptyMessage
	}

	return getFieldFromHeader(e.emailBody.Header, "To")
}

func (e *EmailProcessor) SetFromAddress(addr string) error {
	//TODO implement me
	panic("implement me")
}

func (e *EmailProcessor) SetToAddress(addr string) error {
	//TODO implement me
	panic("implement me")
}

func (e *EmailProcessor) SetReplyToAddress(addr string) error {
	//TODO implement me
	panic("implement me")
}

func (e *EmailProcessor) ParseMessage(r io.Reader) error {
	msg, err := mail.ReadMessage(r)

	if err != nil {
		return err
	}

	e.emailBody = msg
	return nil
}

func getFieldFromHeader(h mail.Header, f string) (string, error) {
	if h == nil {
		return "", ErrEmptyHeader
	}

	val := h.Get(f)
	if len(val) < 1 {
		return "", fmt.Errorf("%w: %s", ErrFieldNotFound, f)
	}

	return val, nil
}
