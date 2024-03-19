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

// Response is the markup of a return json body return
type Response struct {
	Header     http.Header `json:"-"`
	StatusCode int         `json:"-"`
}

// SetHeader method for a response
func (r *Response) SetHeader(header http.Header) {
	r.Header = header
}

// SetStatus for a response
func (r *Response) SetStatus(status int) {
	r.StatusCode = status
}

// String default output of Response
func (r *Response) String() string {
	return fmt.Sprintf("header: %s, status: %d", r.Header, r.StatusCode)
}

// ResponseI constraint for a Response
type ResponseI[T any] interface {
	*T
	SetHeader(header http.Header)
	SetStatus(status int)
}

// Option for a request
type OptionFunc func(req *http.Request)

// WithHeader attach header into the request
func WithHeader(key, value string) OptionFunc {
	return func(req *http.Request) {
		req.Header.Add(key, value)
	}
}

// WithCookie attach cookie into the request
func WithCookie(c *http.Cookie) OptionFunc {
	return func(req *http.Request) {
		req.AddCookie(c)
	}
}

// Get request to url with a context
func Get[ResponseT any, ResponsePT ResponseI[ResponseT]](ctx context.Context, url string, opts ...OptionFunc) (ResponsePT, error) {
	return do[ResponseT, struct{}, ResponsePT](ctx, http.MethodGet, url, nil, opts...)
}

// Post request to url with a context
func Post[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...OptionFunc) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodPost, url, body, opts...)
}

// Put request to url with a context
func Put[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...OptionFunc) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodPut, url, body, opts...)
}

// Put request to url with a context
func Patch[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...OptionFunc) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodPatch, url, body, opts...)
}

// Delete request to url with a context
func Delete[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, url string, body RequestPT, opts ...OptionFunc) (ResponsePT, error) {
	return do[ResponseT, RequestT, ResponsePT](ctx, http.MethodDelete, url, body, opts...)
}

func do[ResponseT any, RequestT any, ResponsePT ResponseI[ResponseT], RequestPT *RequestT](ctx context.Context, method string, url string, body RequestPT, opts ...OptionFunc) (ResponsePT, error) {
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

// Error that response a request
type Error struct {
	error
	Header     http.Header
	StatusCode int
	Body       []byte
}

// Error message
func (e *Error) Error() string {
	return e.error.Error()
}

// Unwrap error
func (e *Error) Unwrap() error {
	return e.error
}
