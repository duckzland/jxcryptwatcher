//go:build local
// +build local

package core

import (
	"net/http"
	"net/url"
)

func AlterRequests(q url.Values, req *http.Request) {
	// Example for adding extra request for local tags
	// req.Header.Set("X-RegKey", "123334")
}
