package logicCache

import (
	"context"
	"fmt"
	"math"
	"time"
)

func ExampleCompareFn() {
	const key = "test"

	c := New(context.Background(), time.Hour, NoopExpire)
	c.CompareFn = math.MaxInt64

	c.Set(key, int64(1))
	c.Set(key, int64(2))

	v, _ := c.Get(key)
	fmt.Println(v)

	// Output:
	// 2
}
