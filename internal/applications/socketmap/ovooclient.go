package socketmap

import (
	"crypto/tls"
	"net/http"
	"time"
)

type OvooErrorBody struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

type OvooError struct {
	Errors []OvooErrorBody `json:"errors"`
}

type OvooClient struct {
	client *http.Client
	server string
	token  string
}

func NewClient(server string, authToken string, tlsSkipVerify bool, timeout time.Duration) (OvooClient, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify},
		},
		Timeout: timeout,
	}

	return OvooClient{
		client: client,
		server: server,
		token:  authToken,
	}, nil
}
