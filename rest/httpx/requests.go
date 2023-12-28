package httpx

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/zeromicro/go-zero/core/mapping"
	"github.com/zeromicro/go-zero/core/validation"
	"github.com/zeromicro/go-zero/rest/internal/encoding"
	"github.com/zeromicro/go-zero/rest/internal/header"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

const (
	formKey           = "form"
	queryKey          = "query"
	pathKey           = "path"
	maxMemory         = 32 << 20 // 32MB
	maxBodyLen        = 8 << 20  // 8MB
	separator         = ";"
	tokensInAttribute = 2
)

var (
	formUnmarshaler  = mapping.NewUnmarshaler(formKey, mapping.WithStringValues(), mapping.WithOpaqueKeys())
	queryUnmarshaler = mapping.NewUnmarshaler(queryKey, mapping.WithStringValues(), mapping.WithOpaqueKeys())
	pathUnmarshaler  = mapping.NewUnmarshaler(pathKey, mapping.WithStringValues(), mapping.WithOpaqueKeys())
	validator        atomic.Value
)

// Validator defines the interface for validating the request.
type Validator interface {
	// Validate validates the request and parsed data.
	Validate(r *http.Request, data any) error
}

// Parse parses the request.
func Parse(r *http.Request, v any) error {
	if err := ParsePath(r, v); err != nil {
		return err
	}

	if err := ParseQuery(r, v); err != nil {
		return err
	}

	if err := ParseForm(r, v); err != nil {
		return err
	}

	if err := ParseHeaders(r, v); err != nil {
		return err
	}

	if err := ParseJsonBody(r, v); err != nil {
		return err
	}

	if valid, ok := v.(validation.Validator); ok {
		return valid.Validate()
	} else if val := validator.Load(); val != nil {
		return val.(Validator).Validate(r, v)
	}

	return nil
}

// ParseHeaders parses the headers request.
func ParseHeaders(r *http.Request, v any) error {
	return encoding.ParseRequestHeaders(r, v)
}

// ParseQuery parses the url query request.
func ParseQuery(r *http.Request, v any) error {
	params, err := GetQueryValues(r)
	if err != nil {
		return err
	}

	unmarshaler := mapping.WithOpts(queryUnmarshaler, mapping.GetUnmarshalOptions(r)...)
	return unmarshaler.Unmarshal(params, v)
}

// ParseForm parses the form request.
func ParseForm(r *http.Request, v any) error {
	params, err := GetFormValues(r)
	if err != nil {
		return err
	}

	unmarshaler := mapping.WithOpts(formUnmarshaler, mapping.GetUnmarshalOptions(r)...)
	return unmarshaler.Unmarshal(params, v)
}

// ParseHeader parses the request header and returns a map.
func ParseHeader(headerValue string) map[string]string {
	ret := make(map[string]string)
	fields := strings.Split(headerValue, separator)

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) == 0 {
			continue
		}

		kv := strings.SplitN(field, "=", tokensInAttribute)
		if len(kv) != tokensInAttribute {
			continue
		}

		ret[kv[0]] = kv[1]
	}

	return ret
}

// ParseJsonBody parses the post request which contains json in body.
func ParseJsonBody(r *http.Request, v any) error {
	opts := mapping.GetUnmarshalOptions(r)
	if withJsonBody(r) {
		reader := io.LimitReader(r.Body, maxBodyLen)
		return mapping.UnmarshalJsonReader(reader, v, opts...)
	}

	return mapping.UnmarshalJsonMap(nil, v, opts...)
}

// ParsePath parses the symbols reside in url path.
// Like http://localhost/bag/:name
func ParsePath(r *http.Request, v any) error {
	vars := pathvar.Vars(r)
	m := make(map[string]any, len(vars))
	for k, v := range vars {
		m[k] = v
	}

	unmarshaler := mapping.WithOpts(pathUnmarshaler, mapping.GetUnmarshalOptions(r)...)
	return unmarshaler.Unmarshal(m, v)
}

// SetValidator sets the validator.
// The validator is used to validate the request, only called in Parse,
// not in ParseHeaders, ParseForm, ParseHeader, ParseJsonBody, ParsePath.
func SetValidator(val Validator) {
	validator.Store(val)
}

func withJsonBody(r *http.Request) bool {
	return r.ContentLength > 0 && strings.Contains(r.Header.Get(header.ContentType), header.ApplicationJson)
}
