package milter

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/d--j/go-milter/mailfilter"
	"github.com/d--j/go-milter/mailfilter/addr"
)

func AddressRewriter(ovooCli OvooClient) func(ctx context.Context, trx mailfilter.Trx) (mailfilter.Decision, error) {
	return func(ctx context.Context, trx mailfilter.Trx) (mailfilter.Decision, error) {
		curFrom, err := getHeaderAddr("from", trx)
		if err != nil {
			return mailfilter.Reject, err
		}

		// check recipients matching any of our configured domains
		var matchingRcpts []*addr.RcptTo
		for _, rcpt := range trx.RcptTos() {
			for _, domain := range ovooCli.domains {
				if strings.Contains(rcpt.Addr, domain) {
					matchingRcpts = append(matchingRcpts, rcpt)
					break
				}
			}
		}

		if len(matchingRcpts) == 0 {
			return mailfilter.Accept, nil
		}

		// only allow single matching recipient per message
		if len(matchingRcpts) > 1 {
			return mailfilter.CustomErrorResponse(522, "5.5.3 Too many recipients"), fmt.Errorf("too many recipients")
		}

		rcpt := matchingRcpts[0]
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
		// if reply-to header is set - reset it to masquaraded "from" value
		if replyto, err := trx.Headers().Text("reply-to"); err == nil && len(replyto) != 0 {
			trx.Headers().Set("reply-to", nfrom.String())
		}

		// delete DKIM headers belong to different domain
		trx.Headers().Set("dkim-signature", "")
		trx.Headers().Set("x-google-dkim-signature", "") // google specific signature

		// delete ARC sealing headers
		trx.Headers().Set("ARC-Seal", "")
		trx.Headers().Set("ARC-Message-Signature", "")
		trx.Headers().Set("ARC-Authentication-Results", "")

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
