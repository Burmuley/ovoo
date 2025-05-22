package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

// ChainsService represents a use case for managing chains
type ChainsService struct {
	repof  *factory.RepoFactory
	domain string
}

// NewChainsService creates a new instance of ChainsUsecase
// It takes a RepoFabric as a parameter and returns a pointer to ChainsUsecase and an error
func NewChainsService(domain string, repof *factory.RepoFactory) (*ChainsService, error) {
	if repof == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	if len(domain) == 0 {
		return nil, fmt.Errorf("%w: domain should be defined", entities.ErrConfiguration)
	}

	return &ChainsService{domain: domain, repof: repof}, nil
}

func (cs *ChainsService) GetByHash(ctx context.Context, cuser entities.User, hash entities.Hash) (entities.Chain, error) {
	if !canGetChain(cuser) {
		return entities.Chain{}, fmt.Errorf("getting chain: %w", entities.ErrNotAuthorized)
	}

	if err := hash.Validate(); err != nil {
		return entities.Chain{}, fmt.Errorf("getting chain by hash: parsing hash: %w", err)
	}

	chain, err := cs.repof.Chain.GetByHash(ctx, hash)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("getting chain by hash: %w", err)
	}

	// TODO: add logic to check if addresses in the chain are active (alias or praddr)
	return chain, nil
}

func (cs *ChainsService) DeleteByHash(ctx context.Context, cuser entities.User, hash entities.Hash) (entities.Chain, error) {
	if !canDeleteChain(cuser) {
		return entities.Chain{}, fmt.Errorf("getting chain: %w", entities.ErrNotAuthorized)
	}

	if err := hash.Validate(); err != nil {
		return entities.Chain{}, fmt.Errorf("getting chain by hash: parsing hash: %w", err)
	}

	chain, err := cs.repof.Chain.Delete(ctx, hash)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("deleting chain by hash: %w", err)
	}

	return chain, nil
}

func (cs *ChainsService) Create(ctx context.Context, cuser entities.User, fromEmail, toEmail string, owner entities.User) (entities.Chain, error) {
	if !canCreateChain(cuser) {
		return entities.Chain{}, fmt.Errorf("getting chain: %w", entities.ErrNotAuthorized)
	}

	// calculate hash and return corresponding chain if found in the DB
	hash := entities.NewHash(fromEmail, toEmail)
	if chain, err := cs.repof.Chain.GetByHash(ctx, hash); err == nil {
		return chain, nil
	}

	// create new chains if no hash present in the DB
	// heck if toEmail already exists in the DB and is of type AliasAddress
	// (reply chain can not be created without initial email)
	// (Ovoo don't accept email for outer domains)
	addrs, err := cs.repof.Address.GetByEmail(ctx, entities.Email(toEmail))
	if err != nil {
		return entities.Chain{}, fmt.Errorf("creating chain: getting destination alias: %w", err)
	}

	var alias *entities.Address
	for _, addr := range addrs {
		if addr.Type == entities.AliasAddress {
			alias = &addr
		}
	}

	if alias == nil {
		return entities.Chain{}, fmt.Errorf("%w: creating chain: destination alias not found", entities.ErrValidation)
	}

	src, err := checkCreateSrcAddr(ctx, cs.repof, fromEmail, owner)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("creating chain: creating source address: %w", err)
	}

	// Generate ReplyAlias(FromAddress, ToAddress)
	// (creates Address record with ForwardAddress set to original external sender)
	ralias, err := genReplyAlias(ctx, cs.repof, fromEmail, toEmail, cs.domain, &src, owner)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("creating chain: %w", err)
	}

	// forward chain
	fchain := entities.Chain{
		Hash:            hash,
		FromAddress:     ralias,
		ToAddress:       *alias.ForwardAddress,
		OrigFromAddress: src,
		OrigToAddress:   *alias,
		CreatedAt:       time.Now().UTC(),
	}

	// reply chain
	rhash := entities.NewHash(string(alias.ForwardAddress.Email), string(ralias.Email))
	rchain := entities.Chain{
		Hash:            rhash,
		FromAddress:     *alias,
		ToAddress:       src,
		OrigFromAddress: *alias.ForwardAddress,
		OrigToAddress:   ralias,
		CreatedAt:       time.Now().UTC(),
	}

	// create chains
	if err := cs.repof.Chain.BatchCreate(ctx, []entities.Chain{fchain, rchain}); err != nil {
		return entities.Chain{}, fmt.Errorf("creating chains: %w", err)
	}

	return fchain, nil
}

func genReplyAlias(ctx context.Context, repof *factory.RepoFactory, fromEmail, toEmail, domain string, fwdAddr *entities.Address, owner entities.User) (entities.Address, error) {
	raliasEmail, _, err := entities.GenReplyAliasEmail(entities.Email(fromEmail), entities.Email(toEmail), domain)
	if err != nil {
		return entities.Address{}, fmt.Errorf("generating new reply alias: %w", err)
	}

	ralias := entities.Address{
		Type:           entities.ReplyAliasAddress,
		ID:             entities.NewId(),
		Email:          raliasEmail,
		ForwardAddress: fwdAddr,
		Owner:          owner,
	}

	if err := repof.Address.Create(ctx, ralias); err != nil {
		return entities.Address{}, fmt.Errorf("storing new reply alias: %w", err)
	}

	return ralias, nil
}

func checkCreateSrcAddr(ctx context.Context, repof *factory.RepoFactory, faddr string, owner entities.User) (entities.Address, error) {
	srcAddrs, err := repof.Address.GetByEmail(ctx, entities.Email(faddr))
	if err != nil && !errors.Is(err, entities.ErrNotFound) {
		return entities.Address{}, err
	}

	// check if requested external address is in the results
	if srcAddrs != nil {
		for _, addr := range srcAddrs {
			if addr.Type == entities.ExternalAddress {
				return addr, nil
			}
		}
	}

	// fallback creating new external address
	srcAddr := entities.Address{
		Type:  entities.ExternalAddress,
		ID:    entities.NewId(),
		Email: entities.Email(faddr),
		Owner: owner,
	}

	if err := repof.Address.Create(ctx, srcAddr); err != nil {
		return entities.Address{}, err
	}

	return srcAddr, nil

	// return entities.Address{}, fmt.Errorf("%w: source address found in DB but is not external address type", entities.ErrValidation)
}
