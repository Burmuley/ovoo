package milter

import (
	"context"
	"fmt"

	"github.com/d--j/go-milter/mailfilter"
)

func AddressRewriter(ovooCli OvooClient) func(ctx context.Context, trx mailfilter.Trx) (mailfilter.Decision, error) {
	return func(ctx context.Context, trx mailfilter.Trx) (mailfilter.Decision, error) {
		for _, rcpt := range trx.RcptTos() {
			chain, err := ovooCli.CreateChain(ctx, trx.MailFrom().Addr, rcpt.Addr)
			if err != nil {
				return mailfilter.Reject, fmt.Errorf("error creating chain in Ovoo API: %w", err)
			}
			fmt.Printf(
				"changed addresses: from_orig=%s, from_changed=%s, to_orig=%s, to_changed=%s\n",
				trx.MailFrom().Addr, chain.FromAddr, rcpt.Addr, chain.ToAddr,
			)
			trx.DelRcptTo(rcpt.Addr)
			trx.AddRcptTo(chain.ToAddr, rcpt.Args)
			trx.ChangeMailFrom(chain.FromAddr, trx.MailFrom().Args)

		}

		return mailfilter.Accept, nil
	}
}
