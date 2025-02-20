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
	repoFactory *factory.RepoFactory
	domain      string
}

// NewChainsService creates a new instance of ChainsUsecase
// It takes a RepoFabric as a parameter and returns a pointer to ChainsUsecase and an error
func NewChainsService(domain string, repoFactory *factory.RepoFactory) (*ChainsService, error) {
	if repoFactory == nil {
		return nil, fmt.Errorf("%w: repository fabric should be defined", entities.ErrConfiguration)
	}

	if len(domain) == 0 {
		return nil, fmt.Errorf("%w: domain should be defined", entities.ErrConfiguration)
	}

	return &ChainsService{domain: domain, repoFactory: repoFactory}, nil
}

func (cuc *ChainsService) GetByHash(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	if err := hash.Validate(); err != nil {
		return entities.Chain{}, fmt.Errorf("getting chain by hash: parsing hash: %w", err)
	}

	chain, err := cuc.repoFactory.Chain.GetByHash(ctx, hash)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("getting chain by hash: %w", err)
	}

	// TODO: add logic to check if addresses in the chain are active (alias or praddr)
	return chain, nil
}

func (cuc *ChainsService) DeleteByHash(ctx context.Context, hash entities.Hash) (entities.Chain, error) {
	if err := hash.Validate(); err != nil {
		return entities.Chain{}, fmt.Errorf("getting chain by hash: parsing hash: %w", err)
	}

	chain, err := cuc.repoFactory.Chain.Delete(ctx, hash)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("deleting chain by hash: %w", err)
	}

	return chain, nil
}

func (cuc *ChainsService) Create(ctx context.Context, faddr, taddr string) (entities.Chain, error) {
	// 1. Calculate Hash(FromAddress, ToAddress) (HashForward)
	fwdHash := entities.NewHash(faddr, taddr)

	// 2. Check if there's a record already exists in the DB, if found - return exiting record
	chain, err := cuc.repoFactory.Chain.GetByHash(ctx, fwdHash)
	if err == nil {
		return chain, nil
	}

	// 3. Check if ToAddress already exists in the Addresses table
	destAddr, err := cuc.repoFactory.Address.GetByEmail(ctx, entities.Email(taddr))
	if err != nil {
		return entities.Chain{}, fmt.Errorf("creating chain: getting destination address: %w", err)
	}

	// 3.1 If found and AddressType == ReplyAlias - return error
	// (reply chain can not be created without initial email)
	// 3.2 If found and AddressType == ProtectedAddress - return error
	// (Ovoo don't accept email for outer domains)
	if destAddr.Type != entities.AliasAddress {
		return entities.Chain{}, fmt.Errorf("%w: creating chain: destination address is not of type 'Alias'", entities.ErrValidation)
	}

	srcAddr, err := checkCreateSrcAddr(ctx, *cuc.repoFactory, faddr)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("creating chain: creating source address: %w", err)
	}

	// 4. Generate ReplyAlias(FromAddress, ToAddress)
	// (creates Address record with ForwardAddress set to original external sender)
	replyAliasEmail, _, err := entities.GenReplyAliasEmail(
		entities.Email(faddr),
		entities.Email(taddr),
		cuc.domain,
	)
	if err != nil {
		return entities.Chain{}, fmt.Errorf("creating chain: generating new reply alias: %w", err)
	}

	replyAlias := entities.Address{
		Type:           entities.ReplyAliasAddress,
		ID:             entities.NewId(),
		Email:          replyAliasEmail,
		ForwardAddress: &srcAddr,
		Owner:          getFirstUser(ctx, *cuc.repoFactory),
		Metadata:       entities.AddressMetadata{},
	}

	if err := cuc.repoFactory.Address.Create(ctx, replyAlias); err != nil {
		return entities.Chain{}, fmt.Errorf("creating chain: storing new reply alias: %w", err)
	}

	// 5. Calculate Hash(AliasAddress.Email, FromAddress) (HashReply)
	replyHash := entities.NewHash(string(destAddr.Email), string(srcAddr.Email))

	// 6. Create forward chain
	//   - Hash: HashForward
	//   - FromAddress: ReplyAlias
	//   - ToAddress: ProtectedAddress
	fwdChain := entities.Chain{
		Hash:        fwdHash,
		FromAddress: replyAlias,
		ToAddress:   *destAddr.ForwardAddress,
		CreatedAt:   time.Now().UTC(),
	}

	// 6.1 If chain creation failed = return error
	if _, err := cuc.repoFactory.Chain.Create(ctx, fwdChain); err != nil {
		return entities.Chain{}, fmt.Errorf("creating forward chain: %w", err)
	}

	// 7. Create reply chain:
	//   - Hash: HashReply
	//   - FromAddress: AliasAddress
	//   - ToAddress: ReplyAlias.ForwardAddress
	replyChain := entities.Chain{
		Hash:        replyHash,
		FromAddress: destAddr,
		ToAddress:   srcAddr,
		CreatedAt:   time.Now().UTC(),
	}

	if _, err := cuc.repoFactory.Chain.Create(ctx, replyChain); err != nil {
		cuc.repoFactory.Chain.Delete(ctx, fwdHash)
		return entities.Chain{}, fmt.Errorf("creating reply chain: %w", err)
	}

	return fwdChain, nil

}

func checkCreateSrcAddr(ctx context.Context, repoFactory factory.RepoFactory, faddr string) (entities.Address, error) {
	var srcAddr entities.Address
	var err error
	srcAddr, err = repoFactory.Address.GetByEmail(ctx, entities.Email(faddr))
	if err != nil && errors.Is(err, entities.ErrNotFound) {
		{
			var err error
			srcAddr = entities.Address{
				Type:  entities.ExternalAddress,
				ID:    entities.NewId(),
				Email: entities.Email(faddr),
				Owner: getFirstUser(ctx, repoFactory),
			}
			err = repoFactory.Address.Create(ctx, srcAddr)
			if err != nil {
				return entities.Address{}, err
			}
		}
	} else if err != nil {
		return entities.Address{}, err
	}

	return srcAddr, nil
}
