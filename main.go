package main

import (
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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

var usage = `usage: harvestdoc [options] ( endpoint | file )

The harvestdoc service pulls downs Harvest concept data and exports it in various
formats current as CSV.

Examples:

  Export a CSV file.

    harvestdoc http://harvest.research.chop.edu/demo/api/ > demo.csv

Options:

`

func main() {
	// Set custom usage message.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}

	var (
		format string
		token  string
	)

	flag.StringVar(&format, "format", "csv", "Export format.")
	flag.StringVar(&token, "token", "", "API token if authorization is required.")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
	}

	var (
		err error
		cs  []*Concept
	)

	if strings.HasPrefix(args[0], "http") {
		c := NewClient(args[0])
		c.Token = token

		cs, err = c.Concepts()

		if err != nil {
			log.Fatal(err)
		}
	} else {
		f, err := os.Open(args[0])

		if err != nil {
			log.Fatal(err)
		}

		if err = json.NewDecoder(f).Decode(&cs); err != nil {
			log.Fatal(err)
		}
	}

	if err := NewCSVEncoder(os.Stdout).Encode(cs); err != nil {
		log.Fatal(err)
	}
}

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

var header = []string{
	"Field",
	"Concept",
	"Category",
	"Description",
}

type CSVEncoder struct {
	csv     *csv.Writer
	started bool
}

func (e *CSVEncoder) Encode(cs []*Concept) error {
	var err error

	if !e.started {
		if err = e.csv.Write(header); err != nil {
			return err
		}

		e.started = true
	}

	for _, c := range cs {
		for _, f := range c.Fields {
			err = e.csv.Write([]string{
				f.Name,
				c.Name,
				c.Category.Name,
				strings.TrimSpace(f.Description),
			})

			if err != nil {
				return err
			}
		}
	}

	e.csv.Flush()

	return nil
}

func NewCSVEncoder(w io.Writer) *CSVEncoder {
	return &CSVEncoder{
		csv: csv.NewWriter(w),
	}
}
