package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestCtx(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("time over")
			return
		default:
			fmt.Println("sleep")
			time.Sleep(20 * time.Second)
			fmt.Println("sleep over")
		}
	}

}
