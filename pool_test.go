package AwesomePool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPool_Do(t *testing.T) {
	gr := New(context.Background(), &Config{
		Concurrent:  10,
		IdleTimeout: 10,
	})
	gr.Do(func(i context.Context) {
		fmt.Println("haha")
	})
	time.Sleep(5 * time.Second)
}
