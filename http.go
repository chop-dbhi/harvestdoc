package main

import (
	"encoding/json"
	"net/http"
)

const StatusUnprocessableEntity = 422

type harvestAPI struct {
	URL   string
	Token string
}

func httpServe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var err error

	api := new(harvestAPI)

	defer r.Body.Close()

	if err = json.NewDecoder(r.Body).Decode(api); err != nil {
		w.WriteHeader(StatusUnprocessableEntity)
		w.Write([]byte(err.Error()))
		return
	}

	c := NewClient(api.URL)
	c.Token = api.Token

	cs, err := c.Concepts()

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("content-type", "text/csv")

	if err = NewCSVEncoder(w).Encode(cs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
