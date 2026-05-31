package milter

import (
	"io"
	"time"

	gomail "github.com/emersion/go-message/mail"

	"github.com/d--j/go-milter/mailfilter"
	"github.com/d--j/go-milter/mailfilter/addr"
	milterheader "github.com/d--j/go-milter/mailfilter/header"
)

// setCall records a single call to mockHeader.Set.
type setCall struct{ key, value string }

// mockHeader is a minimal in-memory implementation of header.Header for testing.
type mockHeader struct {
	textValues map[string]string
	textErrors map[string]error
	setCalls   []setCall
}

func newMockHeader() *mockHeader {
	return &mockHeader{
		textValues: make(map[string]string),
		textErrors: make(map[string]error),
	}
}

func (m *mockHeader) Text(key string) (string, error) {
	if err, ok := m.textErrors[key]; ok {
		return "", err
	}
	return m.textValues[key], nil
}

func (m *mockHeader) Set(key, value string) {
	m.setCalls = append(m.setCalls, setCall{key, value})
}

func (m *mockHeader) hasSet(key, value string) bool {
	for _, c := range m.setCalls {
		if c.key == key && c.value == value {
			return true
		}
	}
	return false
}

// Stub implementations for unused header.Header methods.
func (m *mockHeader) Add(key, value string)                             {}
func (m *mockHeader) Value(key string) string                           { return "" }
func (m *mockHeader) UnfoldedValue(key string) string                   { return "" }
func (m *mockHeader) AddressList(key string) ([]*gomail.Address, error) { return nil, nil }
func (m *mockHeader) SetText(key, value string)                         {}
func (m *mockHeader) SetAddressList(key string, _ []*gomail.Address)    {}
func (m *mockHeader) Subject() (string, error)                          { return "", nil }
func (m *mockHeader) SetSubject(value string)                           {}
func (m *mockHeader) Date() (time.Time, error)                          { return time.Time{}, nil }
func (m *mockHeader) SetDate(value time.Time)                           {}
func (m *mockHeader) Reader() io.Reader                                 { return nil }
func (m *mockHeader) Fields() milterheader.Fields                       { return nil }

// changeMailFromCall records a single call to mockTrx.ChangeMailFrom.
type changeMailFromCall struct{ from, args string }

// addRcptToCall records a single call to mockTrx.AddRcptTo.
type addRcptToCall struct{ rcptTo, args string }

// mockTrx is a minimal implementation of mailfilter.Trx for testing.
type mockTrx struct {
	mailFrom *addr.MailFrom
	rcptTos  []*addr.RcptTo
	headers  *mockHeader

	changeMailFromCalls []changeMailFromCall
	delRcptToCalls      []string
	addRcptToCalls      []addRcptToCall
}

// newMockTrx builds a mockTrx with fromHeader set in the "from" header field and
// envelopeFrom as the SMTP envelope sender address.
func newMockTrx(fromHeader, envelopeFrom string, rcpts ...*addr.RcptTo) *mockTrx {
	hdr := newMockHeader()
	hdr.textValues["from"] = fromHeader
	mf := addr.NewMailFrom(envelopeFrom, "", "", "", "")
	return &mockTrx{
		mailFrom: &mf,
		rcptTos:  rcpts,
		headers:  hdr,
	}
}

func (m *mockTrx) MailFrom() *addr.MailFrom     { return m.mailFrom }
func (m *mockTrx) RcptTos() []*addr.RcptTo      { return m.rcptTos }
func (m *mockTrx) Headers() milterheader.Header { return m.headers }

func (m *mockTrx) ChangeMailFrom(from, args string) {
	m.changeMailFromCalls = append(m.changeMailFromCalls, changeMailFromCall{from, args})
}

func (m *mockTrx) DelRcptTo(a string) {
	m.delRcptToCalls = append(m.delRcptToCalls, a)
}

func (m *mockTrx) AddRcptTo(a, args string) {
	m.addRcptToCalls = append(m.addRcptToCalls, addRcptToCall{a, args})
}

// Stub implementations for unused mailfilter.Trx methods.
func (m *mockTrx) MTA() *mailfilter.MTA         { return nil }
func (m *mockTrx) Connect() *mailfilter.Connect { return nil }
func (m *mockTrx) Helo() *mailfilter.Helo       { return nil }
func (m *mockTrx) HasRcptTo(rcptTo string) bool { return false }
func (m *mockTrx) HeadersEnforceOrder()         {}
func (m *mockTrx) Body() io.ReadSeeker          { return nil }
func (m *mockTrx) ReplaceBody(r io.Reader)      {}
func (m *mockTrx) QueueId() string              { return "" }
func (m *mockTrx) Data() io.Reader              { return nil }
