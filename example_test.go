package logicCache

import (
	"context"
	"time"

	"github.com/vtopc/wcache"
)

func Example() {
	c := wcache.New(context.Background(), 100*time.Millisecond, wcache.PrintlnOnExpire)
	// put value with custom TTL:
	c.SetWithTTL("2", "to expire second", 200*time.Millisecond)
	// put value with default TTL:
	c.Set("1", "to expire first")

	time.Sleep(300 * time.Millisecond)

	// Output:
	// 1: to expire first
	// 2: to expire second
}

func ExampleCache_Done() {
	ctx, cancel := context.WithCancel(context.Background())
	c := wcache.New(ctx, time.Hour, wcache.PrintlnOnExpire)

	c.Set("1", "my value") // should expire in an hour
	// but will expire after context cancellation
	cancel()
	<-c.Done()

	// Output:
	// 1: my value
}
