package pkg

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/polarisjrex0406/static-proxy/constants"
)

var (
	ErrNotEnoughData    = errors.New("not enough data")
	ErrMissingAuth      = errors.New("missing auth")
	ErrPurchaseNotFound = errors.New("purchase not found")
	ErrDomainBlocked    = errors.New("domain blocked")
	ErrIPNotAllowed     = errors.New("ip not allowed")
)

var (
	strBasic = []byte("Basic ")
)

func extractCredentials(originReq *http.Request, bufferedReq *http.Request) (string, string, error) {
	authStr := originReq.Header.Get(constants.HeaderProxyAuthorization)

	if authStr == "" {
		if authStr = bufferedReq.Header.Get(constants.HeaderProxyAuthorization); authStr == "" {
			return "", "", ErrMissingAuth
		}
	}

	username, password, ok := parseBasicAuth([]byte(authStr))
	if !ok {
		return "", "", ErrMissingAuth
	}

	return username, password, nil
}

func parseBasicAuth(credentials []byte) (username string, password string, ok bool) {
	if !bytes.EqualFold(credentials[:6], strBasic) {
		return
	}

	var buf = make([]byte, base64.StdEncoding.DecodedLen(len(credentials)))
	w, err := base64.StdEncoding.Decode(buf, credentials[6:])
	if err != nil {
		return
	}
	buf = buf[:w]
	s := bytes.IndexByte(buf, ':')
	if s < 0 {
		return
	}
	return string(buf[:s]), string(buf[s+1:]), true
}
