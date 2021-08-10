package gofr

import (
	"context"
	"fmt"
	"strings"

	"developer.zopsmart.com/go/gofr/pkg/log"

	"go.opentelemetry.io/collector/translator/conventions"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type exporter struct {
	name    string
	host    string
	port    string
	appName string
}

func TraceProvider(appName, exporterName, exporterHost, exporterPort string, logger log.Logger, config Config) *trace.TracerProvider {
	exporterName = strings.ToLower(exporterName)
	e := exporter{
		name:    exporterName,
		host:    exporterHost,
		port:    exporterPort,
		appName: appName,
	}

	switch exporterName {
	case "zipkin":
		return e.getZipkinExporter(config, logger)
	default:
		return nil
	}
}

func (e *exporter) getZipkinExporter(c Config, logger log.Logger) *trace.TracerProvider {
	url := fmt.Sprintf("http://%s:%s/api/v2/spans", e.host, e.port)

	exporter, err := zipkin.New(url, zipkin.WithSDKOptions(trace.WithSampler(trace.AlwaysSample())))
	if err != nil {
		logger.Errorf("failed to initialize zipkinExporter export pipeline: %v", err)
	}

	batcher := trace.NewBatchSpanProcessor(exporter)

	attributes := []attribute.KeyValue{
		attribute.String(conventions.AttributeTelemetrySDKName, "launcher"),
		attribute.String(conventions.AttributeTelemetrySDKLanguage, "go"),
		attribute.String(conventions.AttributeTelemetrySDKVersion, c.Get("APP_VERSION")),
		attribute.String(conventions.AttributeServiceName, c.Get("APP_NAME")),
	}

	r, err := resource.New(context.Background(), resource.WithAttributes(attributes...))
	if err != nil {
		logger.Errorf("error creating resource")
	}

	tp := trace.NewTracerProvider(trace.WithSpanProcessor(batcher), trace.WithResource(r))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp
}
