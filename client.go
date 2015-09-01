package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

var (
	// The ECDHE cipher suites are preferred for performance and forward
	// secrecy.  See https://community.qualys.com/blogs/securitylabs/2013/06/25/ssl-labs-deploying-forward-secrecy.
	preferredCipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}
)

type Category struct {
	ID     int
	Name   string
	Parent *Category
	Order  float32
}

type Field struct {
	ID            int `json:"pk"`
	Name          string
	Description   string
	AltName       string `json:"alt_name"`
	AltPluralName string `json:"alt_plural_name"`
}

type Concept struct {
	ID          int
	Name        string
	PluralName  string `json:"plural_name"`
	Description string

	Category *Category
	Fields   []*Field

	Published bool
	Queryable bool
	Sortable  bool
	Viewable  bool

	Order float32
}

type Client struct {
	Endpoint string
	Token    string

	http *http.Client
}

func (c *Client) send(path string) (*http.Response, error) {
	uri, err := url.Parse(c.Endpoint)

	if err != nil {
		return nil, err
	}

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	uri.Path = filepath.Join(uri.Path, path)

	req, err := http.NewRequest("GET", uri.String(), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	if c.Token != "" {
		req.Header.Set("Api-Token", c.Token)
	}

	return c.http.Do(req)
}

func (c *Client) Concepts() ([]*Concept, error) {
	resp, err := c.send("concepts/")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("client: %s", resp.Status)
	}

	defer resp.Body.Close()

	var concepts []*Concept

	if err = json.NewDecoder(resp.Body).Decode(&concepts); err != nil {
		return nil, fmt.Errorf("json: %s", err)
	}

	return concepts, nil
}

func NewClient(endpoint string) *Client {
	return &Client{
		Endpoint: endpoint,

		http: &http.Client{
			Timeout: 5 * time.Second,

			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					CipherSuites: preferredCipherSuites,
				},
			},
		},
	}
}
