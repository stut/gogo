package main

// 10 years old, no answer to the licensing question as of usage.
// https://gist.github.com/cespare/3985516
// With modifications by Stuart Dallas [2023-01-23].

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// ip - - [datetime] "method url protocol" status bytes "referer" "user-agent"
	ApacheFormatPattern = "%s - - [%s] \"%s\" %d %d \"%s\" \"%s\"\n"
)

func (handler RedirectHandler) logRequest(r *http.Request, statusCode int, byteCount int) {
	if !handler.requestLogEnabled {
		return
	}

	clientIP := r.Header.Get("X-Forwarded-For")
	if len(clientIP) == 0 {
		clientIP = r.RemoteAddr
		if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
			clientIP = clientIP[:colon]
		}
	} else {
		if strings.Contains(clientIP, ",") {
			clientIP = strings.Split(clientIP, ",")[0]
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, ApacheFormatPattern,
		clientIP,
		time.Now().Format("02/Jan/2006 03:04:05"),
		strings.ReplaceAll(fmt.Sprintf("%s %s %s", r.Method, r.RequestURI, r.Proto), "\"", "\\\""),
		statusCode,
		byteCount,
		strings.ReplaceAll(r.Header.Get("Referer"), "\"", "\\\""),
		strings.ReplaceAll(r.Header.Get("User-Agent"), "\"", "\\\""))
}
