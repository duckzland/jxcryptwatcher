//go:build !local
// +build !local

package core

import (
	"net/http"
	"net/url"
)

func AlterRequests(q url.Values, req *http.Request) {

}
