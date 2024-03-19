package kuro_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duythinht/kuro"
)

type Body struct {
	kuro.Response
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func TestGetOk(t *testing.T) {
	type Body struct {
		kuro.Response
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	body, err := kuro.Get[Body](context.Background(), "https://dummyjson.com/products/1")

	if err != nil {
		t.Fail()
	}

	switch {
	case body.ID != 1:
		t.Fail()
	case body.Title != "iPhone 9":
		t.Fail()
	}
}

func TestGetFail(t *testing.T) {
	type Body struct {
		kuro.Response
		ID    int    `json:"id"`
		Title string `json:"title"`
	}

	_, err := kuro.Get[Body](context.Background(), "https://dummyjson.com/products/1111")

	if err == nil {
		t.Fail()
	}

	var e4xx *kuro.Error

	switch {
	case errors.As(err, &e4xx):
		if e4xx.StatusCode != 404 {
			t.Fail()
		}
	default:
		t.Fail()
	}
}

func TestPost(t *testing.T) {

	type Response struct {
		kuro.Response
		ID      int
		Message string
		Method  string
	}

	type Request struct {
		Message string
		Method  string
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &Request{}

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			w.WriteHeader(400)
		}

		json.NewEncoder(w).Encode(&Response{
			ID:      1,
			Message: req.Message,
			Method:  r.Method,
		})
	}))

	defer srv.Close()

	message := "hello"

	resp, err := kuro.Post[Response](
		context.Background(),
		srv.URL,
		&Request{
			Message: message,
		},
	)

	if err != nil {
		t.Fail()
	}

	if resp.Method != http.MethodPost {
		t.Fail()
	}

	if resp.Message != message {
		t.Fail()
	}
}
