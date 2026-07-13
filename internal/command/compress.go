package command

import "context"
import "fmt"
import "slices"

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
	state *runtime.Session,
	args []string,
) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("too many params, except 1 got %d", len(args))
	}

	var mode string
	if len(args) == 0 {
		mode = "hybrid"
	} else {
		mode = args[0]
	}

	validMode := []string{"truncate", "summarize", "hybrid"}
	if !slices.Contains(validMode, mode) {
		return "",
			fmt.Errorf("invalid mode, except [%v], got %v", validMode, mode)
	}

	switch mode {
	case validMode[0]:
		(&runtime.TruncateCompressor{}).Compress(ctx, state)
	case validMode[1]:
		(&runtime.LLMCompressor{}).Compress(ctx, state)
	case validMode[2]:
		(&runtime.HybridCompressor{}).Compress(ctx, state)
	}

	return "Context has been compressed", nil
}
