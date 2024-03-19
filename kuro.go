package kuro

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	Err4xx               = errors.New("error 4xx")
	Err5xx               = errors.New("error 5xx")
	ErrClientMalfunction = errors.New("error client malfunction")
)

type Response struct {
	Header     http.Header `json:"-"`
	StatusCode int         `json:"-"`
}

func (r *Response) SetHeader(header http.Header) {
	r.Header = header
}

func (r *Response) SetStatus(status int) {
	r.StatusCode = status
}

func (r *Response) String() string {
	return fmt.Sprintf("header: %s, status: %d", r.Header, r.StatusCode)
}

type ResponseI[T any] interface {
	*T
	SetHeader(header http.Header)
	SetStatus(status int)
}

type Option func(params *http.Request)

func WithHeader(key, value string) Option {
	return func(req *http.Request) {
		req.Header.Add(key, value)
	}
}

func Get[ResponseT any, ResponsePT ResponseI[ResponseT]](ctx context.Context, url string, opts ...Option) (ResponsePT, error) {
	return do[ResponseT, struct{}, ResponsePT](ctx, http.MethodGet, url, nil, opts...)
}

func Post[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...Option) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodPost, url, body, opts...)
}

func Put[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...Option) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodPut, url, body, opts...)
}

func Patch[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...Option) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodPatch, url, body, opts...)
}

func Delete[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...Option) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodDelete, url, body, opts...)
}

func do[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, method string, url string, body RequestPT, opts ...Option) (ResponsePT, error) {
	payload, err := json.Marshal(body)

	if err != nil {
		return nil, fmt.Errorf("could not marshal body")
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(payload))

	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(req)
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))

	if err != nil {
		return nil, fmt.Errorf("%w: do request, method: %s", ErrClientMalfunction, method)
	}

	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 500:

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, &Error{
				fmt.Errorf("%w:  could not read error body - %w", Err5xx, ErrClientMalfunction),
				resp.Header,
				resp.StatusCode,
				body,
			}
		}

		return nil, &Error{
			Err5xx,
			resp.Header,
			resp.StatusCode,
			body,
		}

	case resp.StatusCode >= 400:

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, &Error{
				Err4xx,
				resp.Header,
				resp.StatusCode,
				body,
			}
		}

		return nil, &Error{
			Err4xx,
			resp.Header,
			resp.StatusCode,
			body,
		}

	default:
		result := ResponsePT(new(ResponseT))

		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return nil, fmt.Errorf("%w: malformed json body", ErrClientMalfunction)
		}

		result.SetHeader(resp.Header)
		result.SetStatus(resp.StatusCode)

		return result, nil
	}
}

type Error struct {
	error
	Header     http.Header
	StatusCode int
	Body       []byte
}

func (e *Error) Error() string {
	return e.error.Error()
}

func (e *Error) Unwrap() error {
	return e.error
}
