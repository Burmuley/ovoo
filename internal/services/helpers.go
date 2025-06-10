package services

import (
	"context"

	"github.com/Burmuley/ovoo/internal/entities"
	"github.com/Burmuley/ovoo/internal/repositories/factory"
)

/*
deletePrAddrsForUser deletes all Protected addresses belonging to a given user, as well as any associated Alias addresses and their related chains.

It performs the following steps:
1. Retrieves all Protected addresses owned by the specified user.
2. For each Protected address, deletes all its aliases using deleteAliasesForPrAddr.
3. Deletes all the Protected addresses themselves in batch.

Parameters:
- ctx: context for the operation.
- repof: the repository factory to access database operations.
- userId: ID of the user whose Protected addresses and related aliases are to be deleted.

Returns:
- error: describes any error encountered during the process.
*/
func deletePrAddrsForUser(ctx context.Context, repof *factory.RepoFactory, userId entities.Id) error {
	// get all protected addresses
	praddrs, _, err := repof.Address.GetAll(ctx, entities.AddressFilter{
		Filter: entities.Filter{Page: 0, PageSize: 0},
		Types:  []entities.AddressType{entities.ProtectedAddress},
		Owners: []entities.Id{userId},
	})
	if err != nil {
		return err
	}

	// delete aliases for each protected address
	for _, praddr := range praddrs {
		if err := deleteAliasesForPrAddr(ctx, repof, praddr); err != nil {
			return err
		}
	}

	// delete protected addresses
	praddrsIds := make([]entities.Id, 0, len(praddrs))
	for _, addr := range praddrs {
		praddrsIds = append(praddrsIds, addr.ID)
	}

	if err := repof.Address.BatchDeleteById(ctx, praddrsIds); err != nil {
		return err
	}

	return nil
}

/*
deleteAliasesForPrAddr deletes all Alias addresses that forward to a specific Protected address and cleans up their chains.

It performs the following steps:
1. Fetches all Alias addresses forwarding to the given Protected address.
2. Deletes each alias and its chains using deleteAliasIds.

Parameters:
- ctx: context for the operation.
- repof: the repository factory to access database operations.
- praddr: the Protected address entity whose forwarding aliases are to be deleted.

Returns:
- error: describes any error encountered during the process.
*/
func deleteAliasesForPrAddr(ctx context.Context, repof *factory.RepoFactory, praddr entities.Address) error {
	filter := entities.AddressFilter{
		Filter:            entities.Filter{Page: 0, PageSize: 0},
		Types:             []entities.AddressType{entities.AliasAddress},
		Owners:            []entities.Id{praddr.Owner.ID},
		ForwardAddressIds: []entities.Id{praddr.ID},
	}

	aliases, _, err := repof.Address.GetAll(ctx, filter)
	if err != nil {
		return err
	}

	if len(aliases) > 0 {
		aliasIds := make([]entities.Id, 0, len(aliases))
		for _, alias := range aliases {
			aliasIds = append(aliasIds, alias.ID)
		}

		if err := deleteAliasIds(ctx, repof, aliasIds); err != nil {
			return err
		}
	}

	return nil
}

/*
deleteAliasIds deletes a batch of Alias addresses, including their associated chains.

It performs the following steps:
1. Deletes all chains associated with the provided Alias address IDs using deleteChainsForAliasIds.
2. Deletes the Alias addresses themselves in batch.

Parameters:
- ctx: context for the operation.
- repof: the repository factory to access database operations.
- aliasIds: slice of Alias address IDs to delete.

Returns:
- error: describes any error encountered during the process.
*/
func deleteAliasIds(ctx context.Context, repof *factory.RepoFactory, aliasIds []entities.Id) error {
	if err := deleteChainsForAliasIds(ctx, repof, aliasIds); err != nil {
		return err
	}

	if err := repof.Address.BatchDeleteById(ctx, aliasIds); err != nil {
		return err
	}

	return nil
}

/*
deleteChainsForAliasIds deletes all Chain entities associated with a set of Alias address IDs, and also deletes related Reply Alias addresses.

It performs the following steps:
1. Fetches all chain entities where the given Alias address IDs appear in OrigToAddrIds (forward chains) or FromAddrsIds (reply chains).
2. Collects all chain IDs for deletion.
3. For any chain where the FromAddress is of type ReplyAliasAddress, collects those Reply Alias addresses for deletion.
4. Deletes all collected chains in batch.
5. Deletes all found Reply Alias addresses in batch.

Parameters:
- ctx: context for the operation.
- repof: the repository factory to access database operations.
- aliasIds: slice of Alias address IDs to find and clean associated chains.

Returns:
- error: describes any error encountered during the process.
*/
func deleteChainsForAliasIds(ctx context.Context, repof *factory.RepoFactory, aliasIds []entities.Id) error {
	chainsFwd, err := repof.Chain.GetByFilters(ctx, entities.ChainFilter{
		Filter:        entities.Filter{Page: 0, PageSize: 0},
		OrigToAddrIds: aliasIds,
	})
	if err != nil {
		return err
	}

	chainsReply, err := repof.Chain.GetByFilters(ctx, entities.ChainFilter{
		Filter:       entities.Filter{Page: 0, PageSize: 0},
		FromAddrsIds: aliasIds,
	})

	chainIds := make([]entities.Hash, 0, len(chainsFwd))
	chainAddrs := make([]entities.Id, 0)
	chainsLen := max(len(chainsFwd), len(chainsReply))
	for i := range chainsLen {
		if i < len(chainsFwd) {
			chainIds = append(chainIds, chainsFwd[i].Hash)
			if chainsFwd[i].FromAddress.Type == entities.ReplyAliasAddress {
				chainAddrs = append(chainAddrs, chainsFwd[i].FromAddress.ID)
			}
		}
		if i < len(chainsReply) {
			chainIds = append(chainIds, chainsReply[i].Hash)
		}
	}

	if err := repof.Chain.BatchDelete(ctx, chainIds); err != nil {
		return err
	}

	// clean up Reply Alias addresses belong to chains
	if err := repof.Address.BatchDeleteById(ctx, chainAddrs); err != nil {
		return err
	}

	return nil
}
