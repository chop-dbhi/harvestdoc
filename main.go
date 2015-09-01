package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var usage = `usage: harvestdoc [options] ( http | endpoint | file )

The harvestdoc service pulls downs Harvest concept data and exports it in various
formats current as CSV.

Examples:

  Export a CSV file.

    harvestdoc http://harvest.research.chop.edu/demo/api/ > demo.csv

  Run the service.

    harvestdoc http > /dev/null 2>&1 &

    curl -X POST \
        -H "Accept: text/csv" \
        -H "Content-Type: application/json" \
        http://localhost:5000 -d '{
            "url": "http://harvest.research.chop.edu/demo/api/"
        }' > demo.csv


Options:

`

func main() {
	// Set custom usage message.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}

	var (
		host   string
		port   int
		format string
		token  string
	)

	flag.StringVar(&host, "host", "", "Host of the server.")
	flag.IntVar(&port, "port", 5000, "Port of the server.")
	flag.StringVar(&format, "format", "csv", "Export format (CLI only).")
	flag.StringVar(&token, "token", "", "API token if authorization is required (CLI only).")

	flag.Parse()

	args := flag.Args()

	// Process as CLI command.
	if len(args) == 0 {
		flag.Usage()
	}

	// Run HTTP server..
	if args[0] == "http" {
		if token != "" {
			fmt.Fprintln(os.Stderr, "warn: The -token option only applies to the CLI.")
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		fmt.Fprintf(os.Stderr, "* Listening on %s\n", addr)

		http.HandleFunc("/", httpServe)
		log.Fatal(http.ListenAndServe(addr, nil))
		return
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
