package milter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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

type OvooError struct {
	Id  float32 `json:"id"`
	Msg string  `json:"msg"`
}

type OvooClient struct {
	client *http.Client
	server string
	token  string
	domain string
}

func NewOvooClient(server string, authToken string, tlsSkipVerify bool, domain string) (OvooClient, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
		},
	}

	if domain == "" {
		return OvooClient{}, errors.New("missing value for configuration option 'domain'")
	}
	return OvooClient{client: client, server: server, token: authToken, domain: domain}, nil
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

	req, err := http.NewRequestWithContext(ctx, method, queryUrl.String(), body)
	if err != nil {
		return nil, err
	}

	for header, value := range headers {
		req.Header.Add(header, value)
	}

	for param, value := range queryParams {
		req.URL.Query().Add(param, value)
	}

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

func (o OvooClient) parseError(resp *http.Response) error {
	ovooError := OvooError{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &ovooError); err != nil {
		return err
	}

	return fmt.Errorf("ovoo api error: id=%d message=%s", int(ovooError.Id), ovooError.Msg)
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

	if resp.StatusCode != http.StatusCreated {
		return nil, o.parseError(resp)
	}

	return o.parseChainData(resp)
}
