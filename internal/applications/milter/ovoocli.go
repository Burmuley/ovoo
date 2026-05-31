package milter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultEmailDisplayName = "Ovoo Hidden Mail"
	domainCacheTTL          = 5 * time.Minute
)

// in-memory cache for domains value
var domainCache sync.Map

type cachedDomainsInfo struct {
	domains   []string
	expiresAt time.Time
}

type OvooGetDomainsResponse struct {
	Domains []OvooDomainData `json:"domains"`
}

type OvooDomainData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type OvooChainAddressData struct {
	Email string `json:"email"`
	Type  string `json:"type"`
}

type OvooChainData struct {
	Hash            string               `json:"hash"`
	FromEmail       string               `json:"from_email"`
	ToEmail         string               `json:"to_email"`
	OrigFromAddress OvooChainAddressData `json:"orig_from_address"`
	OrigToAddress   OvooChainAddressData `json:"orig_to_address"`
}

type OvooChainCreateRequestBody struct {
	FromEmail string `json:"from_email"`
	ToEmail   string `json:"to_email"`
}

type OvooErrorBody struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type OvooError struct {
	Errors []OvooErrorBody `json:"errors"`
}

type OvooClient struct {
	client      *http.Client
	server      string
	token       string
	displayName string
}

func NewClient(server string, authToken string, tlsSkipVerify bool, timeout time.Duration, displayName string) (OvooClient, error) {
	displayName = strings.TrimSpace(displayName)
	if len(displayName) == 0 {
		displayName = defaultEmailDisplayName
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
		},
		Timeout: timeout,
	}

	return OvooClient{
		client:      client,
		server:      server,
		token:       authToken,
		displayName: displayName,
	}, nil
}

func (o OvooClient) createRequest(ctx context.Context, server, path, method string, body io.Reader, headers map[string]string, queryParams map[string]string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	queryUrl, err := serverURL.Parse(path)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, method, queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	for header, value := range headers {
		req.Header.Add(header, value)
	}

	qr := req.URL.Query()
	for param, value := range queryParams {
		qr.Add(param, value)
	}
	req.URL.RawQuery = qr.Encode()

	return req, nil
}

func (o OvooClient) parseChainData(resp *http.Response) (*OvooChainData, error) {
	data := OvooChainData{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (o OvooClient) parseDomainData(resp *http.Response) ([]string, error) {
	data := OvooGetDomainsResponse{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	domains := make([]string, 0, len(data.Domains))
	for _, domain := range data.Domains {
		domains = append(domains, domain.Name)
	}

	return domains, nil
}

func (o OvooClient) parseError(resp *http.Response) error {
	ovooError := OvooError{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &ovooError); err != nil {
		return err
	}

	cliErrs := make([]OvooErrorBody, 0, len(ovooError.Errors))
	for _, cliErr := range ovooError.Errors {
		cliErrs = append(cliErrs, OvooErrorBody{cliErr.Status, cliErr.Detail})
	}

	return fmt.Errorf("ovoo api errors: %v", cliErrs)
}

func (o OvooClient) CreateChain(ctx context.Context, fromEmail, toEmail string) (*OvooChainData, error) {
	body := OvooChainCreateRequestBody{
		FromEmail: fromEmail,
		ToEmail:   toEmail,
	}
	bodyBytes, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", o.token),
	}
	req, err := o.createRequest(
		ctx,
		o.server,
		"/private/api/v1/chains",
		http.MethodPost,
		bytes.NewReader(bodyBytes),
		headers,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		return nil, o.parseError(resp)
	}

	return o.parseChainData(resp)
}

func (o OvooClient) GetDomains(ctx context.Context) ([]string, error) {
	if val, ok := domainCache.Load("domains"); ok {
		if entry := val.(cachedDomainsInfo); time.Now().Before(entry.expiresAt) {
			return entry.domains, nil
		}
	}

	domains, err := o.getDomainsNetwork(ctx)
	if err != nil {
		return nil, err
	}

	domainCache.Store("domains", cachedDomainsInfo{
		domains:   domains,
		expiresAt: time.Now().Add(domainCacheTTL),
	})

	return domains, nil
}

func (o OvooClient) getDomainsNetwork(ctx context.Context) ([]string, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", o.token),
	}
	req, err := o.createRequest(
		ctx,
		o.server,
		"/api/v1/domains",
		http.MethodGet,
		nil,
		headers,
		map[string]string{
			"active":   "true",
			"verified": "true",
			"global":   "true",
		},
	)
	if err != nil {
		return nil, err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, o.parseError(resp)
	}

	return o.parseDomainData(resp)
}
