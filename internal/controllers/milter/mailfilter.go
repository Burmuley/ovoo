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
				return mailfilter.Reject, fmt.Errorf("error creating chain: %w", err)
			}
			trx.DelRcptTo(rcpt.Addr)
			trx.AddRcptTo(chain.ToAddr, rcpt.Args)
			trx.ChangeMailFrom(chain.FromAddr, trx.MailFrom().Args)
		}

		return mailfilter.Accept, nil
	}
}
