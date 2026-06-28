package main

import "fmt"
import "time"
import "context"



func foo(ctx context.Context) {
	time.Sleep(time.Second * 2)
	fmt.Println("exec foo success")
}

func main() {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	go foo(ctx)

	cancel()

	time.Sleep(time.Second * 4)
}
