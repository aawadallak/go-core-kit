package rest

import (
	"encoding/json"
	"io"
	"net/http"
)

// ResponseOption defines a function type for configuring response options
type ResponseOption func(*configResponse)

// configResponse holds the configuration for an HTTP response
type configResponse struct {
	body   []byte
	header map[string]string
}

// WithBodyByte sets the response body as a byte slice
func WithBodyByte(body []byte) ResponseOption {
	return func(r *configResponse) {
		r.body = body
	}
}

// WithBodyReader sets the response body from an io.Reader
func WithBodyReader(body io.Reader) ResponseOption {
	return func(c *configResponse) {
		b, err := io.ReadAll(body)
		if err != nil {
			c.body = []byte(`{"error": "failed to read body"}`)
			return
		}
		c.body = b
	}
}

// WithBody marshals any value to JSON and sets it as the response body
func WithBody(body any) ResponseOption {
	return func(c *configResponse) {
		b, err := json.Marshal(body)
		if err != nil {
			c.body = []byte(`{"error": "failed to marshal body"}`)
			return
		}
		c.body = b
	}
}

// WithHeader adds a header key-value pair to the response
func WithHeader(key, value string) ResponseOption {
	return func(c *configResponse) {
		if c.header == nil {
			c.header = make(map[string]string)
		}
		c.header[key] = value
	}
}

// newConfigResponse creates a new response configuration with the given options
func newConfigResponse(opts ...ResponseOption) *configResponse {
	c := &configResponse{
		header: make(map[string]string),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// New creates and sends an HTTP response with the specified status and options
func New(w http.ResponseWriter, status int, opts ...ResponseOption) error {
	cfg := newConfigResponse(opts...)

	w.Header().Set("Content-Type", "application/json")
	for k, v := range cfg.header {
		w.Header().Set(k, v)
	}

	w.WriteHeader(status)
	if len(cfg.body) == 0 {
		return nil
	}

	_, err := w.Write(cfg.body)
	return err
}

// Success response helpers
func NewStatusOK(w http.ResponseWriter, opts ...ResponseOption) error {
	return New(w, http.StatusOK, opts...)
}

func NewStatusAccepted(w http.ResponseWriter, opts ...ResponseOption) error {
	return New(w, http.StatusAccepted, opts...)
}

func NewStatusCreated(w http.ResponseWriter, opts ...ResponseOption) error {
	return New(w, http.StatusCreated, opts...)
}

func NewStatusNoContent(w http.ResponseWriter) error {
	return New(w, http.StatusNoContent)
}

// ErrorResponseOption defines a function type for configuring error responses
type ErrorResponseOption func(*errorConfig)

// errorConfig holds the configuration for an error response
type errorConfig struct {
	useErrorField bool
}

// WithErrorField includes the error message under an "error" key in the JSON response
func WithErrorField() ErrorResponseOption {
	return func(cfg *errorConfig) {
		cfg.useErrorField = true
	}
}

// newErrorResponse creates and sends an error response
func newErrorResponse(w http.ResponseWriter, status int, cause error, opts ...ErrorResponseOption) error {
	if cause == nil {
		cause = http.ErrBodyNotAllowed
	}

	cfg := &errorConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var rOpt []ResponseOption
	if cfg.useErrorField {
		rOpt = []ResponseOption{WithBody(struct {
			Error string `json:"error"`
		}{Error: cause.Error()})}
	} else {
		rOpt = []ResponseOption{WithBodyByte([]byte(cause.Error()))}
	}

	return New(w, status, rOpt...)
}

// Error response helpers
func NewStatusBadRequest(w http.ResponseWriter, cause error, opts ...ErrorResponseOption) error {
	return newErrorResponse(w, http.StatusBadRequest, cause, opts...)
}

func NewStatusUnauthorized(w http.ResponseWriter, cause error, opts ...ErrorResponseOption) error {
	return newErrorResponse(w, http.StatusUnauthorized, cause, opts...)
}

func NewStatusNotFound(w http.ResponseWriter, cause error, opts ...ErrorResponseOption) error {
	return newErrorResponse(w, http.StatusNotFound, cause, opts...)
}

func NewStatusConflict(w http.ResponseWriter, cause error, opts ...ErrorResponseOption) error {
	return newErrorResponse(w, http.StatusConflict, cause, opts...)
}

func NewStatusUnprocessableEntity(w http.ResponseWriter, cause error, opts ...ErrorResponseOption) error {
	return newErrorResponse(w, http.StatusUnprocessableEntity, cause, opts...)
}

func NewInternalServerError(w http.ResponseWriter, cause error, opts ...ErrorResponseOption) error {
	return newErrorResponse(w, http.StatusInternalServerError, cause, opts...)
}
