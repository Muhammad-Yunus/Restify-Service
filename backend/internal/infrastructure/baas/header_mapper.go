package baas

import (
	"encoding/json"
	"net/http"
)

// HeaderMapper extracts and maps HTTP headers to endpoint parameters.
type HeaderMapper struct {
	authHeader   string
	paramHeaders map[string]string // header name -> param name
}

// NewHeaderMapper parses auth header name and optional param header mappings from endpoint config.
func NewHeaderMapper(authHeader string, paramHeadersJSON []byte) (*HeaderMapper, error) {
	hm := &HeaderMapper{
		authHeader: authHeader,
	}

	if len(paramHeadersJSON) > 0 {
		var headers map[string]string
		if err := json.Unmarshal(paramHeadersJSON, &headers); err != nil {
			return nil, err
		}
		hm.paramHeaders = headers
	}

	if hm.authHeader == "" {
		hm.authHeader = "Authorization"
	}

	return hm, nil
}

// ExtractAuth returns the authentication value from the request.
func (hm *HeaderMapper) ExtractAuth(r *http.Request) string {
	return r.Header.Get(hm.authHeader)
}

// ExtractParams reads custom header-to-param mappings from the request.
func (hm *HeaderMapper) ExtractParams(r *http.Request) map[string]string {
	if len(hm.paramHeaders) == 0 {
		return nil
	}

	params := make(map[string]string, len(hm.paramHeaders))
	for headerName, paramName := range hm.paramHeaders {
		if value := r.Header.Get(headerName); value != "" {
			params[paramName] = value
		}
	}
	return params
}

// ExtractAll returns auth and params together for function/procedure calls.
func (hm *HeaderMapper) ExtractAll(r *http.Request) (string, map[string]string) {
	return hm.ExtractAuth(r), hm.ExtractParams(r)
}
