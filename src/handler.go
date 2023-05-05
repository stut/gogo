package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type RedirectHandler struct {
	site               string
	metricsEnabled     bool
	requestLogEnabled  bool
	temporaryRedirects map[string]string
	permanentRedirects map[string]string
}

const (
	REDIRECT_ENV_TEMP_PREFIX = "GOGO_TEMP_"
	REDIRECT_ENV_PERM_PREFIX = "GOGO_PERM_"
)

func CreateRedirectHandler(site string, metricsEnabled bool, requestLogEnabled bool) RedirectHandler {
	res := RedirectHandler{
		site:               site,
		metricsEnabled:     metricsEnabled,
		requestLogEnabled:  requestLogEnabled,
		temporaryRedirects: make(map[string]string),
		permanentRedirects: make(map[string]string),
	}

	for _, prefix := range []string{REDIRECT_ENV_TEMP_PREFIX, REDIRECT_ENV_PERM_PREFIX} {
		for _, s := range os.Environ() {
			bits := strings.SplitN(s, "=", 2)
			if strings.HasPrefix(bits[0], prefix) {
				slug := strings.ToLower(bits[0][len(prefix):])
				if prefix == REDIRECT_ENV_TEMP_PREFIX {
					res.temporaryRedirects[slug] = bits[1]
				} else {
					res.permanentRedirects[slug] = bits[1]
				}
			}
		}
	}

	return res
}

func (handler RedirectHandler) GetRedirectCount() int {
	return len(handler.temporaryRedirects) + len(handler.permanentRedirects)
}

func (handler RedirectHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var timer *prometheus.Timer = nil
	if handler.metricsEnabled {
		timer = prometheus.NewTimer(httpDuration.WithLabelValues(handler.site))
	}

	slug := strings.ToLower(req.URL.Path[1:])

	statusCode := http.StatusFound
	content := notFoundContent

	dest, ok := handler.temporaryRedirects[slug]
	if !ok {
		statusCode = http.StatusMovedPermanently
		dest, ok = handler.permanentRedirects[slug]
		if !ok {
			statusCode = http.StatusNotFound
			content = notFoundContent
		}
	}

	if statusCode != http.StatusNotFound {
		w.Header().Add("Location", dest)
		content = []byte(strings.ReplaceAll(string(foundContent), "DEST_URL", dest))
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(content)

	handler.logRequest(req, statusCode, len(content))

	if handler.metricsEnabled {
		responseStatus.WithLabelValues(handler.site, strconv.Itoa(statusCode)).Inc()
		if statusCode != http.StatusNotFound {
			httpRequests.WithLabelValues(handler.site, slug).Inc()
		}
		timer.ObserveDuration()
	}
}
