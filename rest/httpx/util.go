package httpx

import (
	"net/http"
	"net/url"
)

const xForwardedFor = "X-Forwarded-For"

// GetFormValues returns the form values.
func GetFormValues(r *http.Request) (map[string]any, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		if err != http.ErrNotMultipart {
			return nil, err
		}
	}

	params := make(map[string]any, len(r.Form))
	for name := range r.Form {
		formValue := r.Form.Get(name)
		if len(formValue) > 0 {
			params[name] = formValue
		} else {
			params[name] = ""
		}
	}

	return params, nil
}

func GetQueryValues(r *http.Request) (map[string]any, error) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, err
	}

	params := make(map[string]any, len(queryValues))
	for name := range queryValues {
		queryValue := queryValues.Get(name)
		if len(queryValue) > 0 {
			params[name] = queryValue
		} else {
			params[name] = ""
		}
	}
	return params, nil
}

// GetRemoteAddr returns the peer address, supports X-Forward-For.
func GetRemoteAddr(r *http.Request) string {
	v := r.Header.Get(xForwardedFor)
	if len(v) > 0 {
		return v
	}

	return r.RemoteAddr
}
