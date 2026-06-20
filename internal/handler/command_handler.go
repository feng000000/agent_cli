package handler

import "fmt"
import "myagent/internal/runtime"


func HandleCommand(ctx *runtime.LoopContext) error {
	// TODO: implement
	fmt.Printf("got command: %v (not implemented)\n", ctx.Query)
	return nil
}
