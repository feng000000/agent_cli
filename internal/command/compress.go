package command

import "context"
import "fmt"
import "slices"

import agentctx "myagent/internal/context"
import "myagent/internal/runtime"

type CompressCommand struct{}

func (c CompressCommand) Name() string {
	return "compress"
}

// Desc command 作用描述
func (c CompressCommand) Desc() string {
	return "compress context"
}

// Exec
func (c CompressCommand) Exec(
	ctx context.Context,
	state *runtime.AgentState,
	args []string,
) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("too many params, except 1 got %d", len(args))
	}

	var mode string
	if len(args) == 0 {
		mode = "hybrid"
	}
	mode = args[0]

	validMode := []string{"truncate", "summarize", "hybrid"}
	if slices.Contains(validMode, mode) {
		return "",
			fmt.Errorf("invalid compress mode, except [truncate, summarize, hybrid], got %v", mode)
	}

	switch mode {
	case "truncate":
		(&agentctx.TruncateCompressor{}).Compress(ctx, state)
	case "summarize":
		(&agentctx.LLMCompressor{}).Compress(ctx, state)
	case "hybrid":
		(&agentctx.HybridCompressor{}).Compress(ctx, state)
	}

	return "Context has been compressed", nil
}
