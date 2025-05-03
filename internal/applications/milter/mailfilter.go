package milter

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/d--j/go-milter/mailfilter"
)

func AddressRewriter(ovooCli OvooClient) func(ctx context.Context, trx mailfilter.Trx) (mailfilter.Decision, error) {
	return func(ctx context.Context, trx mailfilter.Trx) (mailfilter.Decision, error) {
		curFrom, err := getHeaderAddr("from", trx)
		if err != nil {
			return mailfilter.Reject, err
		}

		for _, rcpt := range trx.RcptTos() {
			if !strings.Contains(rcpt.Addr, ovooCli.domain) {
				continue
			}

			chain, err := ovooCli.CreateChain(ctx, trx.MailFrom().Addr, rcpt.Addr)
			if err != nil {
				return mailfilter.Reject, fmt.Errorf("error creating chain: %w", err)
			}

			var nto mail.Address
			var nfrom mail.Address

			if chain.OrigToAddress.Type == "reply_alias" {
				nfrom = mail.Address{
					Name:    "",
					Address: chain.FromEmail,
				}
				nto = mail.Address{
					Name:    "",
					Address: chain.ToEmail,
				}
			} else {
				nfrom = mail.Address{
					Name:    curFrom.Name,
					Address: chain.FromEmail,
				}
				nto = mail.Address{
					Name:    "Ovoo Hidden Mail",
					Address: rcpt.Addr,
				}
			}

			trx.DelRcptTo(rcpt.Addr)
			trx.AddRcptTo(chain.ToEmail, rcpt.Args)
			trx.ChangeMailFrom(chain.FromEmail, trx.MailFrom().Args)
			trx.Headers().Set("from", nfrom.String())
			trx.Headers().Set("to", nto.String())
		}

		return mailfilter.Accept, nil
	}
}

func getHeaderAddr(header string, trx mailfilter.Trx) (*mail.Address, error) {
	hdr, err := trx.Headers().Text(header)
	if err != nil {
		return nil, fmt.Errorf("error getting %s header: %w", header, err)
	}
	addr, err := mail.ParseAddress(hdr)
	if err != nil {
		return nil, fmt.Errorf("error parsing header %s address: %w", header, err)
	}

	return addr, nil
}
