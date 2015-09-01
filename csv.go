package main

import (
	"encoding/csv"
	"io"
	"strings"
)

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
