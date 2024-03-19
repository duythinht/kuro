package kuro_test

import (
	"context"
	"errors"
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
