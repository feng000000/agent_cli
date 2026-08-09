package executor

import "os"
import "fmt"
import "net"
import "errors"
import "time"
import "net/http"
import "encoding/json"

import "github.com/modelcontextprotocol/go-sdk/mcp"

import "agentcli/internal/tool"
import "agentcli/pkg/logger"


// TODO: 实现 mcp server

type ToolServer struct {
	handler http.Handler
}

func (ts *ToolServer) Run(addr string) {
	http.ListenAndServe(addr, ts.handler)
}

func NewServer() *ToolServer {
	// Create a server with a single tool.
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)


	// // Run the server over stdin/stdout, until the client disconnects.
	// if err := server.Run(context.Background(), &mcp.StreamableServerTransport{}); err != nil {
	// 	log.Fatal(err)
	// }

	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return server
		},
		&mcp.StreamableHTTPOptions{
			Stateless: true,
		},
	)

	return &ToolServer{
		handler: handler,
	}
}
