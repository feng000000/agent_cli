package executor

import "context"
import "time"
import "net"
import "net/http"
import "encoding/json"

import "agentcli/internal/tool"

type ToolExecutorClient interface {
	ListTools() []string
	ExecTool(id string, name string) tool.ToolResult
}

// TODO: implement ToolExecutorClient



// example
func newUnixHTTPClient(socketPath string) *http.Client {
	transport := &http.Transport{
		DialContext: func(
			ctx context.Context,
			network string,
			address string,
		) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(
				ctx,
				"unix",
				socketPath,
			)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

func testNewClient() {

	client := newUnixHTTPClient("/run/user/1000/agent-cli/executor.sock")

	reqBody := Request{
		ToolName:      "read_file",
		Arguments: json.RawMessage(`{"path":"/tmp/a.txt"}`),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://unix/v1/tools/execute",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
}
