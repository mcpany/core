package observability

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// MeterProvider is a global meter provider for the application.
var MeterProvider metric.MeterProvider

// Start initializes the OpenTelemetry SDK and starts the Prometheus HTTP server.
func Start(ctx context.Context, addr string) (*http.Server, error) {
	// Create a new Prometheus exporter.
	exporter, err := otelprometheus.New()
	if err != nil {
		return nil, err
	}

	// Create a new meter provider with the Prometheus exporter.
	MeterProvider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))

	// Create a new HTTP server to expose the /metrics endpoint.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt-in to exposing OpenMetrics.
			EnableOpenMetrics: true,
		},
	))
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start the HTTP server in a separate goroutine.
	go func() {
		_ = server.ListenAndServe()
	}()

	return server, nil
}
