package github

import (
	"bytes"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

var (
	errInvalidClient          = errors.New("error, invalid service client connection")
	errorNotStruct            = errors.New("error, not a struct")
	errorMappingJson          = errors.New("error while trying to map json response")
	errorRateLimitLowLevel    = errors.New("error, rate limit low level")
	errorRateLimitReached     = errors.New("error, rate limite for the current token is reached.")
	errorResponseIsNull       = errors.New("error while receiving the response of the http request.")
	errorLanguageNotSupported = errors.New("isLanguageSupported: invalid prjLangs type")
	errorMarshallingResponse  = errors.New("error while trying to marshall the api response, entity object is nil")
	errTooManyCall            = errors.New("API rate limit exceeded")
	errUnavailable            = errors.New("resource unavailable")
	errRuntime                = errors.New("runtime error")
	errInvalidArgs            = errors.New("invalid arguments")
	errNilArg                 = errors.New("nil argument")
	errInvalidParamType       = errors.New("invalid parameter type")
)

type invalidStructError struct {
	message string
	fields  []string
}

func newInvalidStructError(msg string) *invalidStructError {
	return &invalidStructError{message: msg, fields: []string{}}
}

func (e *invalidStructError) AddField(f string) *invalidStructError {
	e.fields = append(e.fields, f)
	return e
}

func (e invalidStructError) FieldsLen() int {
	return len(e.fields)
}

func (e invalidStructError) Error() string {
	buf := bytes.NewBufferString(e.message)
	buf.WriteString("{ ")
	buf.WriteString(strings.Join(e.fields, ", "))
	buf.WriteString(" }\n")

	return buf.String()
}

func isTemporaryError(err error, wait bool) bool {
	defer funcTrack(time.Now())

	if err == nil {
		return false
	}
	// Get the underlying error, if this is a Wrapped error by the github.com/pkg/errors package.
	// If not, this will just return the error itself.
	underlyingErr := errors.Cause(err)
	switch underlyingErr.(type) {
	case *github.RateLimitError:
		return true
	case *github.AbuseRateLimitError:
		if wait {
			time.Sleep(2 * time.Second)
		}
		return true
	default:
		if strings.Contains(err.Error(), "abuse detection") {
			if wait {
				time.Sleep(2 * time.Second)
			}
			return true
		}
		if strings.Contains(err.Error(), "try again") {
			if wait {
				time.Sleep(2 * time.Second)
			}
			return true
		}
		return false
	}
}
