package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

type HTTPError interface {
	StatusCode() int
	ToJSON() ([]byte, error)
}

type GRPCError interface {
	StatusCode() string
}

func toSnakeCase(str string) string {
	result := make([]rune, 0, len(str))
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

type BaseError struct {
	Code       string
	Message    string
	Cause      error
	Attributes map[string]any
}

func (b *BaseError) Error() string {
	if b.Cause != nil {
		return fmt.Sprintf("%s: %s", b.Code, b.Cause.Error())
	}

	return b.Code
}

func (b *BaseError) ToJSON() ([]byte, error) {
	output := map[string]any{
		"code":    b.Code,
		"message": b.Message,
	}

	if b.Cause != nil {
		output["cause"] = b.Cause.Error()
	}

	if len(b.Attributes) > 0 {
		output["attributes"] = b.Attributes
	}

	return json.Marshal(output)
}

func (b *BaseError) StatusCode() int {
	return http.StatusUnprocessableEntity
}

type TypeErrInternalServerErr struct {
	*BaseError
}

func (t *TypeErrInternalServerErr) Is(target error) bool {
	_, ok := target.(*TypeErrInternalServerErr)
	return ok
}

func (t *TypeErrInternalServerErr) StatusCode() int {
	return http.StatusInternalServerError
}

var ErrInternalServer = &TypeErrInternalServerErr{}

func NewErrInternalServer(cause error) error {
	return &TypeErrInternalServerErr{
		BaseError: &BaseError{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "Oops, something went wrong, please try again or contact support.",
			Cause:   errors.New(strings.ReplaceAll(cause.Error(), "\"", "")),
		},
	}
}

type TypeErrBearerNotFound struct {
	*BaseError
}

var ErrBearerNotFound = &TypeErrBearerNotFound{}

// Is implements the error interface for type checking
func (t *TypeErrBearerNotFound) Is(target error) bool {
	_, ok := target.(*TypeErrBearerNotFound)
	return ok
}

func (t *TypeErrBearerNotFound) StatusCode() int {
	return http.StatusUnauthorized
}

func NewErrBearerNotFound() error {
	return &TypeErrBearerNotFound{
		BaseError: &BaseError{
			Code:    "INVALID_BEARER_TOKEN", // Was 40100
			Message: "The authorization header does not contain a valid bearer token",
		},
	}
}

type TypeErrNotAuthorized struct {
	*BaseError
}

var ErrNotAuthorized = &TypeErrNotAuthorized{}

// Is implements the error interface for type checking
func (t *TypeErrNotAuthorized) Is(target error) bool {
	_, ok := target.(*TypeErrNotAuthorized)
	return ok
}

func (t *TypeErrNotAuthorized) StatusCode() int {
	return http.StatusUnauthorized
}

// NewErrNotAuthorized creates a new TypeErrNotAuthorized error
func NewErrNotAuthorized(cause error) error {
	return &TypeErrNotAuthorized{
		BaseError: &BaseError{
			Code:    "NOT_AUTHORIZED", // Was 40101
			Message: "Not authorized",
			Cause:   cause,
		},
	}
}

type TypeErrParseRequest struct {
	*BaseError
}

func (t *TypeErrParseRequest) Is(target error) bool {
	_, ok := target.(*TypeErrParseRequest)
	return ok
}

func (t *TypeErrParseRequest) StatusCode() int {
	return http.StatusBadRequest
}

var ErrParseRequest = &TypeErrParseRequest{}

func NewErrParseRequest(cause error) error {
	return &TypeErrParseRequest{
		BaseError: &BaseError{
			Code:    "INVALID_REQUEST_BODY", // Was 40102
			Message: "Oops, something went wrong while decoding the body. Please validate the Content-Type or body message",
			Cause:   cause,
		},
	}
}

type TypeErrRequiredFields struct {
	*BaseError
}

func (t *TypeErrRequiredFields) Is(target error) bool {
	_, ok := target.(*TypeErrRequiredFields)
	return ok
}

func (t *TypeErrRequiredFields) StatusCode() int {
	return http.StatusBadRequest
}

var ErrRequiredFields = &TypeErrRequiredFields{}

func NewErrRequiredFields(cause error) error {
	attr := make(map[string]any)
	var fields validator.ValidationErrors
	if ok := errors.As(cause, &fields); ok {
		for _, field := range fields {
			attr[toSnakeCase(field.Field())] = field.Error()
		}
	}

	return &TypeErrRequiredFields{
		BaseError: &BaseError{
			Code:       "MISSING_REQUIRED_FIELDS", // Was 40103
			Message:    "Missing required fields",
			Attributes: attr,
		},
	}
}

type TypeErrRequiredField struct {
	*BaseError
}

func (t *TypeErrRequiredField) Is(target error) bool {
	_, ok := target.(*TypeErrRequiredField)
	return ok
}

func (t *TypeErrRequiredField) StatusCode() int {
	return http.StatusBadRequest
}

var ErrRequiredField = &TypeErrRequiredField{}

func NewErrRequiredField(field string) error {
	return &TypeErrRequiredField{
		BaseError: &BaseError{
			Code:    "MISSING_REQUIRED_FIELD", // Was 40104
			Message: "Missing required field",
			Attributes: map[string]any{
				"field": field,
			},
		},
	}
}

type TypeErrResourceNotFound struct {
	*BaseError
}

func (t *TypeErrResourceNotFound) Is(target error) bool {
	_, ok := target.(*TypeErrResourceNotFound)
	return ok
}

func (t *TypeErrResourceNotFound) StatusCode() int {
	return http.StatusNotFound
}

var ErrResourceNotFound = &TypeErrResourceNotFound{}

func NewErrResourceNotFound(err error) error {
	return &TypeErrResourceNotFound{
		BaseError: &BaseError{
			Code:    "RESOURCE_NOT_FOUND", // Was 40400
			Message: "Resource not found",
			Cause:   err,
		},
	}
}

type TypeErrRecoverableError struct {
	*BaseError
}

func (t *TypeErrRecoverableError) Is(target error) bool {
	_, ok := target.(*TypeErrRecoverableError)
	return ok
}

func (t *TypeErrRecoverableError) FailureMode() FailureMode {
	return FailureModeRecoverable
}

var ErrRecoverableError = &TypeErrRecoverableError{}

func NewErrRecoverableError(cause error) error {
	return &TypeErrRecoverableError{
		BaseError: &BaseError{
			Code:    "RECOVERABLE_ERROR",
			Message: "Recoverable error",
			Cause:   cause,
		},
	}
}

type TypeErrNonRecoverableError struct {
	*BaseError
}

func (t *TypeErrNonRecoverableError) Is(target error) bool {
	_, ok := target.(*TypeErrNonRecoverableError)
	return ok
}

func (t *TypeErrNonRecoverableError) FailureMode() FailureMode {
	return FailureModeNonRecoverable
}

var ErrNonRecoverableError = &TypeErrNonRecoverableError{}

func NewErrNonRecoverableError(cause error) error {
	return &TypeErrNonRecoverableError{
		BaseError: &BaseError{
			Code:    "NON_RECOVERABLE_ERROR",
			Message: "Non-recoverable error",
			Cause:   cause,
		},
	}
}

type TypeErrConflict struct {
	*BaseError
}

func (e *TypeErrConflict) Is(target error) bool {
	_, ok := target.(*TypeErrConflict)
	return ok
}

func (e *TypeErrConflict) StatusCode() int {
	return http.StatusConflict
}

var ErrConflict = &TypeErrConflict{}

func NewErrConflict(cause error) error {
	return &TypeErrConflict{
		BaseError: &BaseError{
			Code:    "CONFLICT",
			Message: cause.Error(),
			Cause:   cause,
		},
	}
}
