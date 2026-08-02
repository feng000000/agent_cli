package executor

import "os"
import "net"
import "errors"
import "time"
import "net/http"
import "encoding/json"

import "agentcli/pkg/logger"


// TODO: tool executor server 启动服务器监听端口
//   /api/v1/tool/list
//   /api/v1/tool/execute

func serveHTTPS(addr, certFile, keyFile string) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server.ListenAndServeTLS(certFile, keyFile)
}

func serveUnix(socketPath string) error {
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
		Handler: newHandler(),
	}

	return server.Serve(listener)
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
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

	resp := Response{
		Result: "executed: " + req.Tool,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// 客户端可能已断开，此时通常只记录日志
		logger.Errorf("encode response: %v", err)
	}
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/tools/execute", executeHandler)
	return mux
}
