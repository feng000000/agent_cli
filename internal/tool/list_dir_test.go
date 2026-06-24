package tool

import (
	"fmt"
	"testing"
	"myagent/pkg/logger"
)

func TestListDir(t *testing.T) {

	if err := logger.InitLogger("DEBUG", ""); err != nil {
		panic(err)
	}
	tool := ListDirTool{}

	res := make(chan string)
	tool.Execute("{\"path\": \".\"}", res)

	fmt.Printf("res:\n%v\n", <- res)
}
