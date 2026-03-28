// Package otelwrap provides otelwrap functionality.
package otelwrap

import (
	"github.com/aawadallak/go-core-kit/core/conf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Tracer retorna um tracer global (do provider configurado via otel.SetTracerProvider).
// NÃO cria provider novo. Só pega do registro global.
func Tracer(name ...string) trace.Tracer {
	if len(name) > 0 && name[0] != "" {
		return otel.Tracer(name[0])
	}

	return otel.Tracer(conf.Instance().
		GetString("OTEL_SERVICE_NAME"))
}

// Meter retorna um meter global (do provider configurado via otel.SetMeterProvider).
// NÃO cria provider novo. Só pega do registro global.
func Meter(name ...string) metric.Meter {
	if len(name) > 0 && name[0] != "" {
		return otel.Meter(name[0])
	}

	return otel.Meter(conf.Instance().
		GetString("OTEL_SERVICE_NAME"))
}
