package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StdioTransport implements mcp.Transport for a local command,
// capturing stderr to provide better error messages on failure.
type StdioTransport struct {
	Command *exec.Cmd
}

// Connect starts the command and returns a connection.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (t *StdioTransport) Connect(_ context.Context) (mcp.Connection, error) {
	log := logging.GetLogger()

	stdin, err := t.Command.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := t.Command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := t.Command.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Capture stderr
	stderrCapture := &tailBuffer{limit: 4096}
	// We also want to log stderr to the application logs
	// Note: slog.LevelError is imported from "log/slog"
	logWriter := &slogWriter{log: log, level: slog.LevelError}

	multiStderr := io.MultiWriter(stderrCapture, logWriter)

	conn := &stdioConn{
		stdin:         stdin,
		stdout:        stdout,
		cmd:           t.Command,
		stderrCapture: stderrCapture,
		decoder:       json.NewDecoder(stdout),
		encoder:       json.NewEncoder(stdin),
	}
	conn.wg.Add(1)

	go func() {
		defer conn.wg.Done()
		if _, err := io.Copy(multiStderr, stderr); err != nil && err != io.EOF {
			log.Error("Failed to copy stderr", "error", err)
		}
	}()

	if err := t.Command.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	return conn, nil
}

type stdioConn struct {
	stdin         io.WriteCloser
	stdout        io.ReadCloser
	cmd           *exec.Cmd
	stderrCapture *tailBuffer
	decoder       *json.Decoder
	encoder       *json.Encoder
	mutex         sync.Mutex
	closed        bool
	wg            sync.WaitGroup
}

// Read reads a JSON-RPC message from the standard output of the command.
//
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (c *stdioConn) Read(_ context.Context) (jsonrpc.Message, error) {
	var raw json.RawMessage
	if err := c.decoder.Decode(&raw); err != nil {
		if err == io.EOF {
			// Check if process has exited
			if exitErr := c.checkExit(); exitErr != nil {
				// Process exited with error
				stderr := c.stderrCapture.String()
				if stderr != "" {
					return nil, fmt.Errorf("process exited with error: %v. Stderr: %s", exitErr, stderr)
				}
				return nil, fmt.Errorf("process exited with error: %v", exitErr)
			}
			// Normal EOF?
			stderr := c.stderrCapture.String()
			if stderr != "" {
				return nil, fmt.Errorf("connection closed. Stderr: %s", stderr)
			}
		}
		return nil, err
	}

	var header struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message header: %w", err)
	}

	var msg jsonrpc.Message
	isRequest := header.Method != ""

	if isRequest {
		req := &jsonrpc.Request{}
		if err := json.Unmarshal(raw, req); err != nil {
			// Handle ID unmarshal issue if necessary (copied from docker_transport)
			type requestAnyID struct {
				Method string          `json:"method"`
				Params json.RawMessage `json:"params,omitempty"`
				ID     any             `json:"id,omitempty"`
			}
			var rAny requestAnyID
			if err2 := json.Unmarshal(raw, &rAny); err2 != nil {
				return nil, fmt.Errorf("failed to unmarshal request: %w (and %v)", err2, err)
			}
			req = &jsonrpc.Request{
				Method: rAny.Method,
				Params: rAny.Params,
			}
			if err := setUnexportedID(&req.ID, rAny.ID); err != nil {
				logging.GetLogger().Error("Failed to set unexported ID on request", "error", err)
			}
			msg = req
		} else {
			msg = req
		}
	} else {
		resp := &jsonrpc.Response{}
		if err := json.Unmarshal(raw, resp); err != nil {
			type responseAnyID struct {
				Result json.RawMessage `json:"result,omitempty"`
				Error  *transportError `json:"error,omitempty"`
				ID     any             `json:"id,omitempty"`
			}
			var rAny responseAnyID
			if err2 := json.Unmarshal(raw, &rAny); err2 != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w (and %v)", err2, err)
			}
			resp = &jsonrpc.Response{
				Result: rAny.Result,
			}
			if rAny.Error != nil {
				resp.Error = rAny.Error
			}
			if err := setUnexportedID(&resp.ID, rAny.ID); err != nil {
				logging.GetLogger().Error("Failed to set unexported ID on response", "error", err)
			}
			msg = resp
		} else {
			msg = resp
		}
	}

	return msg, nil
}

// Write writes a JSON-RPC message to the standard input of the command.
//
// _ is an unused parameter.
// msg is the msg.
//
// Returns an error if the operation fails.
func (c *stdioConn) Write(_ context.Context, msg jsonrpc.Message) error {
	var method string
	var params any
	var result any
	var errorObj any
	var id any

	if req, ok := msg.(*jsonrpc.Request); ok {
		method = req.Method
		params = req.Params
		id = fixID(req.ID)
	} else if resp, ok := msg.(*jsonrpc.Response); ok {
		result = resp.Result
		errorObj = resp.Error
		id = fixID(resp.ID)
	}

	wire := map[string]any{
		"jsonrpc": "2.0",
	}
	if method != "" {
		wire["method"] = method
	}
	if params != nil {
		wire["params"] = params
	}
	if id != nil {
		wire["id"] = id
	}
	if result != nil {
		wire["result"] = result
	}
	if errorObj != nil {
		wire["error"] = errorObj
	}

	return c.encoder.Encode(wire)
}

// Close terminates the command and closes the streams.
//
// Returns an error if the operation fails.
func (c *stdioConn) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	_ = c.stdin.Close()
	_ = c.stdout.Close()
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	return nil
}

// SessionID returns a static session ID for the stdio connection.
//
// Returns the result.
func (c *stdioConn) SessionID() string {
	return "stdio-session"
}

func (c *stdioConn) checkExit() error {
	// Wait for the command to finish.
	// If it's already finished, Wait will return the error.
	// However, Wait can only be called once.
	// Since we are in Read() loop, we might have multiple reads?
	// But if we hit EOF, the stream is done, so calling Wait is appropriate.
	// Wait for the stderr copier goroutine to finish to ensure we captured all stderr
	c.wg.Wait()
	err := c.cmd.Wait()
	return err
}
