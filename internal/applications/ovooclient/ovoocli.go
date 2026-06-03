package ovooclient

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
	domainCacheTTL = 5 * time.Minute
)

// in-memory cache for domains value
var domainCache sync.Map

type cachedActiveDomainsInfo struct {
	domains   []string
	expiresAt time.Time
}

type cachedDomainNameInfo struct {
	domain_name string
	expiresAt   time.Time
}

type GetDomainsResponse struct {
	Domains []DomainData `json:"domains"`
}

type DomainData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type ChainAddressData struct {
	Email string `json:"email"`
	Type  string `json:"type"`
}

type ChainData struct {
	Hash            string           `json:"hash"`
	FromEmail       string           `json:"from_email"`
	ToEmail         string           `json:"to_email"`
	OrigFromAddress ChainAddressData `json:"orig_from_address"`
	OrigToAddress   ChainAddressData `json:"orig_to_address"`
}

type ChainCreateRequestBody struct {
	FromEmail string `json:"from_email"`
	ToEmail   string `json:"to_email"`
}

type ErrorBody struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type Error struct {
	Errors []ErrorBody `json:"errors"`
}

type Client struct {
	client *http.Client
	server string
	token  string
}

func NewClient(server string, authToken string, tlsSkipVerify bool, timeout time.Duration) (Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
		},
		Timeout: timeout,
	}

	return Client{
		client: client,
		server: server,
		token:  authToken,
	}, nil
}

func (o Client) createRequest(ctx context.Context, server, path, method string, body io.Reader, headers map[string]string, queryParams map[string]string) (*http.Request, error) {
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

func (o Client) parseChainData(resp *http.Response) (*ChainData, error) {
	data := ChainData{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (o Client) parseDomainData(resp *http.Response) ([]string, error) {
	data := GetDomainsResponse{}
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

func (o Client) parseError(resp *http.Response) error {
	ovooError := Error{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &ovooError); err != nil {
		return err
	}

	cliErrs := make([]ErrorBody, 0, len(ovooError.Errors))
	for _, cliErr := range ovooError.Errors {
		cliErrs = append(cliErrs, ErrorBody{cliErr.Status, cliErr.Detail})
	}

	return fmt.Errorf("ovoo api errors: %v", cliErrs)
}

func (o Client) CreateChain(ctx context.Context, fromEmail, toEmail string) (*ChainData, error) {
	body := ChainCreateRequestBody{
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

func (o Client) GetDomains(ctx context.Context) ([]string, error) {
	if val, ok := domainCache.Load("domains"); ok {
		if entry := val.(cachedActiveDomainsInfo); time.Now().Before(entry.expiresAt) {
			return entry.domains, nil
		}
	}

	domains, err := o.getDomainsNetwork(ctx, "")
	if err != nil {
		return nil, err
	}

	domainCache.Store("domains", cachedActiveDomainsInfo{
		domains:   domains,
		expiresAt: time.Now().Add(domainCacheTTL),
	})

	return domains, nil
}

func (o Client) GetDomainByName(ctx context.Context, domain_name string) bool {
	domain_name = strings.TrimSpace(domain_name)
	if len(domain_name) == 0 {
		return false
	}

	if val, ok := domainCache.Load("domain_name:" + domain_name); ok {
		if entry := val.(cachedDomainNameInfo); time.Now().Before(entry.expiresAt) {
			return true
		}
	}

	domain, err := o.getDomainsNetwork(ctx, domain_name)
	if err != nil {
		fmt.Printf("getDomainsNetwork err: %s\n", err.Error())
		return false
	}

	if len(domain) == 0 {
		return false
	}

	domainCache.Store("domain_name:"+domain_name, cachedDomainNameInfo{
		domain_name: domain_name,
		expiresAt:   time.Now().Add(domainCacheTTL),
	})

	return true
}

func (o Client) getDomainsNetwork(ctx context.Context, domain_name string) ([]string, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", o.token),
	}
	query_params := map[string]string{
		"active":   "true",
		"verified": "true",
	}

	if len(domain_name) > 0 {
		query_params["domain_name"] = domain_name
	}

	req, err := o.createRequest(
		ctx,
		o.server,
		"/api/v1/domains",
		http.MethodGet,
		nil,
		headers,
		query_params,
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
