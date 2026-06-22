package handler

import "myagent/internal/runtime"
import "myagent/pkg/logger"


func HandleCommand(ctx *runtime.LoopContext) error {
	// TODO: implement
	logger.Errorf("got command: %v (not implemented)\n", ctx.Query)
	return nil
}
