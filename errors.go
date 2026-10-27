// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Category classifies an [Error]. Every error the client returns belongs to
// exactly one category, and each category has a sentinel error (such as
// [ErrNotFound]) that matches it with [errors.Is].
type Category int

// The error categories.
const (
	// CategoryAuthentication: the API key is missing or invalid (401).
	CategoryAuthentication Category = iota + 1
	// CategoryPermission: the key's plan does not allow the request (403),
	// for example PRO_REQUIRED.
	CategoryPermission
	// CategoryNotFound: the location, run or report does not exist (404, 410).
	CategoryNotFound
	// CategoryInvalidRequest: the server rejected the request's parameters (400).
	CategoryInvalidRequest
	// CategoryOutsideCoverage: the coordinates are outside the forecast domain.
	CategoryOutsideCoverage
	// CategoryRateLimited: a rate limit or quota was exceeded (429).
	CategoryRateLimited
	// CategoryUnavailable: the server could not answer (5xx).
	CategoryUnavailable
	// CategoryConnection: no response arrived (DNS, TLS, reset, timeout).
	CategoryConnection
	// CategoryUsage: the SDK refused the call before sending it, for example
	// an invalid probability term.
	CategoryUsage
)

var categoryNames = map[Category]string{
	CategoryAuthentication:  "authentication",
	CategoryPermission:      "permission",
	CategoryNotFound:        "not_found",
	CategoryInvalidRequest:  "invalid_request",
	CategoryOutsideCoverage: "outside_coverage",
	CategoryRateLimited:     "rate_limited",
	CategoryUnavailable:     "unavailable",
	CategoryConnection:      "connection",
	CategoryUsage:           "usage",
}

// String returns the category's snake_case name, such as "not_found".
func (c Category) String() string {
	if s, ok := categoryNames[c]; ok {
		return s
	}
	return fmt.Sprintf("Category(%d)", int(c))
}

// Sentinel errors, one per [Category]. Test for a category with errors.Is,
// and use errors.As with *[Error] for the details:
//
//	if errors.Is(err, ensembleweather.ErrRateLimited) { ... }
var (
	ErrAuthentication  = errors.New("ensembleweather: authentication failed")
	ErrPermission      = errors.New("ensembleweather: permission denied")
	ErrNotFound        = errors.New("ensembleweather: not found")
	ErrInvalidRequest  = errors.New("ensembleweather: invalid request")
	ErrOutsideCoverage = errors.New("ensembleweather: outside coverage")
	ErrRateLimited     = errors.New("ensembleweather: rate limited")
	ErrUnavailable     = errors.New("ensembleweather: unavailable")
	ErrConnection      = errors.New("ensembleweather: connection failed")
	ErrUsage           = errors.New("ensembleweather: invalid usage")
)

var sentinels = map[Category]error{
	CategoryAuthentication:  ErrAuthentication,
	CategoryPermission:      ErrPermission,
	CategoryNotFound:        ErrNotFound,
	CategoryInvalidRequest:  ErrInvalidRequest,
	CategoryOutsideCoverage: ErrOutsideCoverage,
	CategoryRateLimited:     ErrRateLimited,
	CategoryUnavailable:     ErrUnavailable,
	CategoryConnection:      ErrConnection,
	CategoryUsage:           ErrUsage,
}

// ErrorCode is the API's machine-readable error code, the "error" field of an
// error response. Codes the SDK does not know are kept verbatim.
type ErrorCode string

// The error codes the SDK maps to a category. The API may add codes; an
// unknown code keeps its string and is categorised by HTTP status.
const (
	CodeInvalidAPIKey        ErrorCode = "INVALID_API_KEY" //nolint:gosec // G101: an error code, not a credential
	CodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	CodeProRequired          ErrorCode = "PRO_REQUIRED"
	CodeForbidden            ErrorCode = "FORBIDDEN"
	CodeLocationNotFound     ErrorCode = "LOCATION_NOT_FOUND"
	CodeModelRunNotFound     ErrorCode = "MODEL_RUN_NOT_FOUND"
	CodeNotFound             ErrorCode = "NOT_FOUND"
	CodeGone                 ErrorCode = "GONE"
	CodeBadRequest           ErrorCode = "BAD_REQUEST"
	CodeInvalidCoordinates   ErrorCode = "INVALID_COORDINATES"
	CodeInvalidElevation     ErrorCode = "INVALID_ELEVATION"
	CodeInvalidMaxHours      ErrorCode = "INVALID_MAX_HOURS"
	CodeElevationRequired    ErrorCode = "ELEVATION_REQUIRED"
	CodeAmbiguous            ErrorCode = "AMBIGUOUS"
	CodeFeelsLikeUnavailable ErrorCode = "FEELS_LIKE_UNAVAILABLE"
	CodeOutsideCoverage      ErrorCode = "OUTSIDE_COVERAGE"
	CodeTooManyRequests      ErrorCode = "TOO_MANY_REQUESTS"
	CodeForecastUnavailable  ErrorCode = "FORECAST_UNAVAILABLE"
	CodeTimeout              ErrorCode = "TIMEOUT"
	CodeInternalError        ErrorCode = "INTERNAL_ERROR"
	CodeDatabaseError        ErrorCode = "DATABASE_ERROR"
)

var codeCategories = map[ErrorCode]Category{
	CodeInvalidAPIKey:        CategoryAuthentication,
	CodeUnauthorized:         CategoryAuthentication,
	CodeProRequired:          CategoryPermission,
	CodeForbidden:            CategoryPermission,
	CodeLocationNotFound:     CategoryNotFound,
	CodeModelRunNotFound:     CategoryNotFound,
	CodeNotFound:             CategoryNotFound,
	CodeGone:                 CategoryNotFound,
	CodeBadRequest:           CategoryInvalidRequest,
	CodeInvalidCoordinates:   CategoryInvalidRequest,
	CodeInvalidElevation:     CategoryInvalidRequest,
	CodeInvalidMaxHours:      CategoryInvalidRequest,
	CodeElevationRequired:    CategoryInvalidRequest,
	CodeAmbiguous:            CategoryInvalidRequest,
	CodeFeelsLikeUnavailable: CategoryInvalidRequest,
	CodeOutsideCoverage:      CategoryOutsideCoverage,
	CodeTooManyRequests:      CategoryRateLimited,
	CodeForecastUnavailable:  CategoryUnavailable,
	CodeTimeout:              CategoryUnavailable,
	CodeInternalError:        CategoryUnavailable,
	CodeDatabaseError:        CategoryUnavailable,
}

// categorize maps an error response to its category: by code first, then by
// HTTP status when the code is unknown or absent.
func categorize(status int, code ErrorCode) Category {
	if c, ok := codeCategories[code]; ok {
		return c
	}
	switch {
	case status == 401:
		return CategoryAuthentication
	case status == 403:
		return CategoryPermission
	case status == 404, status == 410:
		return CategoryNotFound
	case status == 429:
		return CategoryRateLimited
	case status >= 500:
		return CategoryUnavailable
	default:
		return CategoryInvalidRequest
	}
}

// Error is the error type every client call returns. It carries the HTTP
// status, the API's error code and message, and the request ID when the server
// sent one. It never contains the API key.
//
// Match a category with errors.Is and a sentinel (errors.Is(err, ErrNotFound)),
// and read the details with errors.As:
//
//	var apiErr *ensembleweather.Error
//	if errors.As(err, &apiErr) && apiErr.Code == ensembleweather.CodeOutsideCoverage { ... }
type Error struct {
	// Category is the error's category.
	Category Category
	// Operation is the API operation that failed, such as "getLocationForecast".
	Operation string
	// Status is the HTTP status, or 0 when no response arrived.
	Status int
	// Code is the API's error code, verbatim; empty when the body was not a
	// JSON error (or for connection and usage errors).
	Code ErrorCode
	// Message is the human-readable explanation.
	Message string
	// RequestID is the response's X-Request-Id, when present.
	RequestID string
	// RetryAfter is the server's Retry-After on a 429 or 503, or 0.
	RetryAfter time.Duration
	// RateLimit is the response's rate-limit state, when it carried one.
	RateLimit *RateLimit
	// Err is the underlying error for connection errors, or nil.
	Err error
}

// Error implements the error interface.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString("ensembleweather: ")
	if e.Operation != "" {
		b.WriteString(e.Operation)
		b.WriteString(": ")
	}
	switch {
	case e.Category == CategoryUsage:
		b.WriteString("invalid usage: ")
		b.WriteString(e.Message)
	case e.Status == 0:
		b.WriteString(e.Category.String())
		b.WriteString(" error")
		if e.Message != "" {
			b.WriteString(": ")
			b.WriteString(e.Message)
		}
	default:
		fmt.Fprintf(&b, "%d", e.Status)
		if e.Code != "" {
			b.WriteString(" ")
			b.WriteString(string(e.Code))
		}
		if e.Message != "" {
			b.WriteString(": ")
			b.WriteString(e.Message)
		}
	}
	if e.RequestID != "" {
		b.WriteString(" (request ID ")
		b.WriteString(e.RequestID)
		b.WriteString(")")
	}
	return b.String()
}

// Is reports whether target is the sentinel error of e's category.
func (e *Error) Is(target error) bool {
	s, ok := sentinels[e.Category]
	return ok && target == s
}

// Unwrap returns the underlying error of a connection error.
func (e *Error) Unwrap() error { return e.Err }

func usageError(op, format string, args ...any) *Error {
	return &Error{Category: CategoryUsage, Operation: op, Message: fmt.Sprintf(format, args...)}
}
