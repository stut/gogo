package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

var (
	VERSION = 1

	responseStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gogo_status_total",
			Help: "HTTP response status.",
		},
		[]string{"site", "status"},
	)
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gogo_requests_total",
			Help: "HTTP requests total.",
		},
		[]string{"site", "slug"})
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gogo_response_time_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"site"})
	notFoundContent = []byte("<html><head><title>404 Not Found</title></head><body><h1>404 Not Found</h1></body></html>")
	foundContent    = "<html><head><title>Redirecting...</title><meta http-equiv=\"refresh\" content=\"0;URL='DEST_URL'\" /></head><body><h1>Redirecting...</h1><p><a href=\"DEST_URL\">Click here if you are not redirected automatically.</a></p></body></html>"
)

func main() {
	var err error

	listenPort := os.Getenv("NOMAD_PORT_http")
	if len(listenPort) == 0 {
		listenPort = "3000"
	}
	site := os.Getenv("GOGO_SITE")
	if len(site) == 0 {
		site = "notset"
	}
	listenAddr := flag.String("listen-addr", fmt.Sprintf(":%s", listenPort), "Address on which to listen for HTTP requests")
	noMetrics := flag.Bool("no-metrics", false, "Disable prometheus metrics")
	healthUrl := flag.String("health-url", "/health", "Healthcheck URL")
	metricsUrl := flag.String("metrics-url", "/metrics", "Prometheus metrics URL")
	notFoundFilename := flag.String("not-found-filename", "", "Page not found content filename, uses a default page if not provided")
	foundFilename := flag.String("found-filename", "", "Redirect found content filename, uses a default page if not provided; \"DEST_URL\" will be replaced with the destination URL")
	disableApacheLogging := flag.Bool("no-request-logging", false, "Disable Apache request logging to stdout")

	flag.Parse()

	// Read the 404 content. If reading fails the default content is used.
	if len(*notFoundFilename) > 0 {
		var content []byte
		content, err = os.ReadFile(*notFoundFilename)
		if err == nil {
			notFoundContent = content
		}
	}

	// Read the page found content. If reading fails the default content is used.
	if len(*foundFilename) > 0 {
		var content []byte
		content, err = os.ReadFile(*foundFilename)
		if err == nil {
			foundContent = string(content)
		}
	}

	// Handle healthcheck requests. No metrics, no content.
	http.HandleFunc(*healthUrl, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Set up the metrics endpoint.
	if !*noMetrics {
		http.Handle(*metricsUrl, promhttp.Handler())
	}

	handler := CreateRedirectHandler(site, !(*noMetrics), !(*disableApacheLogging))
	http.Handle("/", handler)

	log.Printf("Gogo v%d", VERSION)
	log.Printf("  Listen addr: %s", *listenAddr)
	log.Printf("  Healthcheck: %s", *healthUrl)
	if !*noMetrics {
		log.Printf("  Prometheus metrics: %s", *metricsUrl)
	}
	log.Printf("  Redirect count: %d", handler.GetRedirectCount())

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
