package pkg

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/polarisjrex0406/static-proxy/config"
	"github.com/polarisjrex0406/static-proxy/constants"
)

const (
	strHeaderBasicRealm = "Basic realm=\"\"\r\n\r\n"
)

var hopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // spelling per https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func ListenHTTP(ctx context.Context, httpServer *http.Server) error {
	go httpServer.ListenAndServe() //nolint:errcheck

	<-ctx.Done()
	return httpServer.Shutdown(ctx)
}

func HandlerHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodConnect {
		w.WriteHeader(http.StatusHTTPVersionNotSupported)
		return
	}

	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	username, password, err := extractCredentials(req, req)
	if err != nil || username != cfg.Proxy.Credentials.User || password != cfg.Proxy.Credentials.Password {
		w.WriteHeader(http.StatusProxyAuthRequired)
		w.Header().Add(constants.HeaderProxyAuthenticate, strHeaderBasicRealm)
		return
	}

	client := &http.Client{}
	// When a http.Request is sent through an http.Client, RequestURI should not
	// be set (see documentation of this field).
	req.RequestURI = ""

	removeHopHeaders(req.Header)
	removeConnectionHeaders(req.Header)

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		appendHostToXForwardHeader(req.Header, clientIP)
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(req.RemoteAddr, " ", resp.Status)

	removeHopHeaders(resp.Header)
	removeConnectionHeaders(resp.Header)

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func removeHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

// removeConnectionHeaders removes hop-by-hop headers listed in the "Connection"
// header of h. See RFC 7230, section 6.1
func removeConnectionHeaders(h http.Header) {
	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = strings.TrimSpace(sf); sf != "" {
				h.Del(sf)
			}
		}
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}
