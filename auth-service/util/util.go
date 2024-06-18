package util

import (
	"github.com/afiskon/promtail-client/promtail"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"
)

func HttpTraceError(err error, errMessage string, span trace.Span, loki promtail.Client, funcName, stringifyData string) {
	loki.Errorf("service_name = %s, time = %s, error = %s, data = %s\n", funcName, time.Now().String(), err, stringifyData)
	log.Printf(errMessage)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

func lokiInfo(loki promtail.Client, funcName, stringifyData string) {
	loki.Infof("service_name = %s, time = %s, data = %s\n", funcName, time.Now().String(), stringifyData)
}

func HttpTraceInfo(message string, span trace.Span, loki promtail.Client, funcName, stringifyData string) {
	loki.Infof("service_name = %s, time = %s, data = %s\n", funcName, time.Now().String(), stringifyData)
	log.Printf(message)
	span.AddEvent(message, trace.WithAttributes(attribute.String("stringify_data", stringifyData)))
}
