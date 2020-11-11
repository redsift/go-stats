package tracer

import (
	"context"
	"github.com/redsift/go-rstid"
	"go.opentelemetry.io/otel/api/global"
	"testing"
	"time"
)

func TestSpansWithCtx(t *testing.T) {
	_, err := InitTracingProvider("127.0.0.1:55680", "stats-test")
	if err != nil {
		t.Fatal(err)
	}

	tracer := global.TracerProvider().Tracer("redsift/trace-test")


	ctx, span := tracer.Start(context.Background(), "test-span-normal-1")
	time.Sleep(time.Second*2)
	span.End()

	ctx, span = tracer.Start(ctx, "test-span-normal-2")
	time.Sleep(time.Second*1)
	span.End()


	ctx, span = tracer.Start(ctx, "test-span-normal-3")
	time.Sleep(time.Second*1)
	span.End()

	// give some time for collector to send data
	time.Sleep(time.Minute * 2)
}

func TestSpansWithoutCtx(t *testing.T) {
	_, err := InitTracingProvider("127.0.0.1:55680", "stats-test")
	if err != nil {
		t.Fatal(err)
	}
	reqID := rstid.Generate("stats-test-4")

	tracer := global.TracerProvider().Tracer("redsift/trace-test")

	ctx, err := ContextWithIDs(context.Background(), reqID.String(), 1)
	if err != nil {
		t.Fatal(err)
	}
	_, span := tracer.Start(ctx, "test-span-custom-1", )
	time.Sleep(time.Second * 2)
	span.End()

	ctx, err = ContextWithIDs(context.Background(), reqID.String(), 1)
	if err != nil {
		t.Fatal(err)
	}
	_, span = tracer.Start(ctx, "test-span-custom-2")
	time.Sleep(time.Second * 1)
	span.End()

	ctx, err = ContextWithIDs(context.Background(), reqID.String(), 1)
	if err != nil {
		t.Fatal(err)
	}
	_, span = tracer.Start(ctx, "test-span-custom-3")
	time.Sleep(time.Second * 1)
	span.End()

	// give some time for collector to send data
	time.Sleep(time.Minute * 2)
}