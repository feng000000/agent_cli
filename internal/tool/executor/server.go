package executor

import "os"
import "net"
import "errors"
import "net/http"
import "encoding/json"

import "agentcli/internal/tool"
import "agentcli/pkg/logger"


type ToolServer struct {
	tools []tool.InternalTool
}

// func serveHTTPS(addr, certFile, keyFile string) error {
// 	server := &http.Server{
// 		Addr:              addr,
// 		Handler:           newHandler(),
// 		ReadHeaderTimeout: 5 * time.Second,
// 	}

// 	return server.ListenAndServeTLS(certFile, keyFile)
// }

func (ts *ToolServer) serveUnix(socketPath string) error {
	if err := os.Remove(socketPath); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer listener.Close()

	if err := os.Chmod(socketPath, 0o600); err != nil {
		return err
	}

	server := &http.Server{
		Handler: newToolHandler(),
	}

	return server.Serve(listener)
}

func (ts *ToolServer) newToolHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tools/list", ts.listHandler)
	mux.HandleFunc("POST /v1/tools/execute", ts.executeHandler)
	return mux
}

// listHandler 列出工具信息
func (ts *ToolServer) listHandler(w http.ResponseWriter, r *http.Request) {
	tool_content := []any{}
	for _, t := range ts.tools {
		tool_content = append(
			tool_content,
			map[string]string{
				"name": t.Name(),
				"desc": t.Desc(),
			},
		)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(Response{Content: tool_content})
	if err != nil {
		logger.Errorf("encode response failed: %v", err)
	}
}

func (ts *ToolServer) executeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO: execute Tool
	resp := Response{
		Content: "executed: "
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// 客户端可能已断开，此时通常只记录日志
		logger.Errorf("encode response: %v", err)
	}
}
