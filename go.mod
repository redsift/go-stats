module github.com/redsift/go-stats

go 1.13

replace go.opentelemetry.io/otel => github.com/redsift/opentelemetry-go v0.13.1-beta

replace go.opentelemetry.io/otel/exporters/otlp => github.com/redsift/opentelemetry-go/exporters/otlp v0.13.0

replace go.opentelemetry.io/otel/sdk => github.com/redsift/opentelemetry-go/sdk v0.13.1-beta

require (
	github.com/PagerDuty/godspeed v0.0.0-20180224001232-122876cde329
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc
	github.com/kr/pretty v0.1.0 // indirect
	github.com/redsift/go-cfg v0.1.0
	github.com/redsift/go-errs v0.1.0
	github.com/redsift/go-foodfans v0.9.0 // indirect
	github.com/redsift/go-rstid v1.0.0
	github.com/tinylib/msgp v1.1.2 // indirect
	go.opentelemetry.io/otel v0.13.1-beta
	go.opentelemetry.io/otel/exporters/otlp v0.13.0
	go.opentelemetry.io/otel/sdk v0.13.0
	golang.org/x/net v0.0.0-20191002035440-2ec189313ef0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)
