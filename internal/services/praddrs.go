package services

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Burmuley/ovoo/internal/config"
	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

type PrAddrCreateCmd struct {
	Email    entities.Email
	Metadata struct {
		Comment     *string
		ServiceName *string
	}
}

type PrAddrUpdateCmd struct {
	PrAddrId entities.Id
	Metadata struct {
		Comment     *string
		ServiceName *string
	}
	Active *bool
}

// ProtectedAddrService handles operations related to protected addresses
type ProtectedAddrService struct {
	repof                *factory.RepoFactory
	smtpClient           *SMTPClient
	praddrVerifyTmpl     string
	praddrVerifyHostname string
	logger               *slog.Logger
	verifyEmailWG        sync.WaitGroup
}

// NewProtectedAddrService creates a new ProtectedAddrService
func NewProtectedAddrService(repoFactory *factory.RepoFactory, template string, notifyCfg config.MailNotificationConfig, logger *slog.Logger) (*ProtectedAddrService, error) {
	if repoFactory == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	smtpClient, err := NewSMTPClient(notifyCfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", entities.ErrConfiguration, err)
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &ProtectedAddrService{
		repof: repoFactory, smtpClient: smtpClient,
		praddrVerifyTmpl:     template,
		praddrVerifyHostname: notifyCfg.OvooHostname,
		logger:               logger,
		verifyEmailWG:        sync.WaitGroup{},
	}, nil
}

// Create creates a new protected address
func (prs *ProtectedAddrService) Create(ctx context.Context, cuser entities.User, cmd PrAddrCreateCmd) (entities.Address, error) {
	if !canCreatePrAddr(cuser) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	if err := cmd.Email.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	// check if protected address with the email already exists
	if addrs, err := prs.repof.Address.GetByEmail(ctx, cmd.Email); err == nil {
		for _, addr := range addrs {
			if addr.Type == entities.ProtectedAddress {
				return entities.Address{}, fmt.Errorf("%w: %s", entities.ErrDuplicateEntry, cmd.Email)
			}
		}
	}

	metadata := entities.AddressMetadata{}
	if cmd.Metadata.Comment != nil {
		metadata.Comment = strings.TrimSpace(*cmd.Metadata.Comment)
	}

	if cmd.Metadata.ServiceName != nil {
		metadata.ServiceName = strings.TrimSpace(*cmd.Metadata.ServiceName)
	}
	praddr := entities.Address{
		Type:      entities.ProtectedAddress,
		ID:        entities.NewId(),
		Email:     cmd.Email,
		Metadata:  metadata,
		Owner:     cuser,
		UpdatedBy: cuser,
		Active:    true,
	}

	if err := praddr.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	if err := prs.repof.Address.Create(ctx, praddr); err != nil {
		return entities.Address{}, err
	}

	// send the verification email asynchronously so creation latency isn't tied to
	// the SMTP round trip, and a slow/failing send can't undo an already-persisted address
	prs.verifyEmailWG.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				prs.logger.Error("panic while sending protected address verification email", "praddr_id", praddr.ID, "panic", r)
			}
		}()

		sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()

		if err := prs.SendVerifyEmail(sendCtx, cuser, praddr.ID); err != nil {
			prs.logger.Error("sending protected address verification email", "praddr_id", praddr.ID, "email", string(praddr.Email), "err", err)
		}
	})

	return praddr, nil
}

// ShutdownVerifyEmailWG waits for any in-flight verification emails to finish sending, up to ctx's deadline.
func (prs *ProtectedAddrService) ShutdownVerifyEmailWG(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		prs.verifyEmailWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		prs.logger.Info("all verification emails handlers has been finished")
	case <-ctx.Done():
		prs.logger.Warn("shutdown timed out waiting for in-flight verification emails")
	}
}

// Update updates an existing protected address
func (prs *ProtectedAddrService) Update(ctx context.Context, cuser entities.User, cmd PrAddrUpdateCmd) (entities.Address, error) {
	praddr, err := prs.repof.Address.GetById(ctx, cmd.PrAddrId)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
		}

		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrDatabase, err)
	}

	if !canUpdatePrAddr(cuser, praddr) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	praddr.UpdatedBy = cuser
	if cmd.Metadata.Comment != nil {
		praddr.Metadata.Comment = *cmd.Metadata.Comment
	}
	if cmd.Metadata.ServiceName != nil {
		praddr.Metadata.ServiceName = *cmd.Metadata.ServiceName
	}

	if cmd.Active != nil {
		if canSetActivePrAddr(praddr, cuser) {
			if *cmd.Active && !praddr.Active {
				praddr.Active = *cmd.Active
			} else if !*cmd.Active && praddr.Active {
				if err := deactivateAliasesForPrAddr(ctx, prs.repof, cuser, praddr.ID); err != nil {
					return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrDatabase, err)
				}
				praddr.Active = *cmd.Active
			}
		} else {
			return entities.Address{}, entities.ErrNotAuthorized
		}
	}

	if err := praddr.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	if err := prs.repof.Address.Update(ctx, praddr); err != nil {
		return entities.Address{}, err
	}

	return praddr, nil
}

// GetAll retrieves all protected addresses for a given owner
func (prs *ProtectedAddrService) GetAll(ctx context.Context, cuser entities.User, filter entities.AddressFilter) ([]entities.Address, entities.PaginationMetadata, error) {
	filter.Types = []entities.AddressType{entities.ProtectedAddress}
	// reset Owners filter for non-admins
	if cuser.Type != entities.AdminUser {
		filter.Owners = []entities.Id{cuser.ID}
	} else if slices.Contains(filter.Owners, "all") && cuser.Type == entities.AdminUser {
		filter.Owners = nil
	} else if filter.Owners == nil {
		filter.Owners = []entities.Id{cuser.ID}
	}

	return prs.repof.Address.GetAll(ctx, filter)
}

// GetById retrieves a protected address by its ID
func (prs *ProtectedAddrService) GetById(ctx context.Context, cuser entities.User, id entities.Id) (entities.Address, error) {
	if err := id.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	praddr, err := prs.repof.Address.GetById(ctx, id)
	if err != nil {
		return entities.Address{}, err
	}

	if !canGetPrAddr(cuser, praddr) {
		return entities.Address{}, entities.ErrNotAuthorized
	}

	return praddr, nil
}

func (prs *ProtectedAddrService) GetByEmail(ctx context.Context, cuser entities.User, email entities.Email) (entities.Address, error) {
	if err := email.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	praddrs, err := prs.repof.Address.GetByEmail(ctx, entities.Email(email))
	if err != nil {
		return entities.Address{}, err
	}

	for _, praddr := range praddrs {
		if praddr.Type == entities.ProtectedAddress {
			if !canGetPrAddr(cuser, praddr) {
				return entities.Address{}, entities.ErrNotAuthorized
			}
			return praddr, nil
		}
	}

	return entities.Address{}, fmt.Errorf("%w: %s", entities.ErrNotFound, email)
}

func (prs *ProtectedAddrService) DeleteById(ctx context.Context, cuser entities.User, id entities.Id) error {
	if err := id.Validate(); err != nil {
		return fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	praddr, err := prs.repof.Address.GetById(ctx, id)
	if err != nil {
		return err
	}

	if !canDeletePrAddr(cuser, praddr) {
		return entities.ErrNotAuthorized
	}

	// first - retrieve all aliases and delete them
	if err := deleteAliasesForPrAddr(ctx, prs.repof, cuser, praddr); err != nil {
		return err
	}

	// delete protected address after all referencing entities (aliases) has been deleted successfully
	if err := prs.repof.Address.DeleteById(ctx, cuser, id); err != nil {
		return err
	}

	return nil
}

func (prs *ProtectedAddrService) SendVerifyEmail(ctx context.Context, cuser entities.User, prAddrId entities.Id) error {
	if err := prAddrId.Validate(); err != nil {
		return fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	// get protected address by ID
	prAddr, err := prs.repof.Address.GetById(ctx, prAddrId)
	if err != nil {
		return err
	}

	token, tokenStr, err := entities.NewAddressVerifyToken(time.Now().Add(time.Hour*12), prAddrId)
	if err != nil {
		return err
	}

	// create new token keeping old ones also active
	// in case previous emails did not reach the destination or received after delay
	if err := prs.repof.AddressVerify.Create(ctx, token); err != nil {
		return err
	}

	// form a verification link that points to the WebUI handler
	// token is placed in the URL fragment so it is never sent to any server
	verifyLink := fmt.Sprintf(
		"https://%s/verify/%s#%s",
		prs.praddrVerifyHostname,
		prAddrId.String(),
		tokenStr,
	)

	// render email template
	tmpl, err := template.New("verify").Parse(prs.praddrVerifyTmpl)
	if err != nil {
		return err
	}

	msg := new(bytes.Buffer)
	if err := tmpl.Execute(msg, struct{ VerifyLink string }{VerifyLink: verifyLink}); err != nil {
		return err
	}

	// send verification email to the Protected Address
	if err := prs.smtpClient.SendMessage(string(prAddr.Email), msg.Bytes()); err != nil {
		return err
	}

	return nil
}

func (prs *ProtectedAddrService) ValidateVerifyToken(ctx context.Context, cuser entities.User, addrId entities.Id, tokenInput string) (entities.Address, error) {
	if err := addrId.Validate(); err != nil {
		return entities.Address{}, fmt.Errorf("%w: %w", entities.ErrValidation, err)
	}

	tokenId, rawToken, err := entities.DecodeAddressVerifyToken(tokenInput)
	if err != nil {
		return entities.Address{}, err
	}

	token, err := prs.repof.AddressVerify.GetById(ctx, tokenId)
	if err != nil {
		return entities.Address{}, err
	}

	// verify token validity
	if err := token.Validate(rawToken); err != nil {
		return entities.Address{}, err
	}

	// verify if token expired
	if token.IsExpired() {
		return entities.Address{}, fmt.Errorf("%w: verification token has expired", entities.ErrValidation)
	}

	// get protected address by ID
	addr, err := prs.repof.Address.GetById(ctx, addrId)
	if err != nil {
		return entities.Address{}, err
	}

	// if address is already verified - return clean
	if addr.Verified {
		return addr, nil
	}

	// verify token belongs to the address
	if subtle.ConstantTimeCompare([]byte(token.AddrId), []byte(addrId)) != 1 {
		return entities.Address{}, fmt.Errorf("%w: token does not belong to the address", entities.ErrValidation)
	}

	// mark address as verified
	addr.Verified = true
	addr.VerifiedAt = time.Now().UTC()

	if err := prs.repof.Address.Update(ctx, addr); err != nil {
		return entities.Address{}, err
	}

	// delete verification token after successful address verification
	if err := prs.repof.AddressVerify.Delete(ctx, cuser, token.ID); err != nil {
		return entities.Address{}, err
	}

	return addr, nil
}
