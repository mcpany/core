/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sourcegraph/jsonrpc2"
)

type AddParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResult struct {
	Result int `json:"result"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type stdioHandler struct{}

func (h *stdioHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Notif {
		return
	}

	var result interface{}
	var err error

	switch req.Method {
	case "tools/list":
		result = []Tool{
			{Name: "add", Description: "Adds two integers"},
		}
	case "tools/call":
		var callParams struct {
			ToolName string          `json:"tool_name"`
			Inputs   json.RawMessage `json:"inputs"`
		}
		if err = json.Unmarshal(*req.Params, &callParams); err != nil {
			break
		}

		if callParams.ToolName == "add" {
			var params AddParams
			if err = json.Unmarshal(callParams.Inputs, &params); err == nil {
				result = AddResult{Result: params.A + params.B}
			}
		} else {
			err = fmt.Errorf("tool not found: %s", callParams.ToolName)
		}
	default:
		err = &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
	}

	if err != nil {
		respErr, ok := err.(*jsonrpc2.Error)
		if !ok {
			respErr = &jsonrpc2.Error{Message: err.Error()}
		}
		if err := conn.Reply(ctx, req.ID, respErr); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to send error reply:", err)
		}
		return
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to send reply:", err)
	}
}

type stdrwc struct{}

func (s *stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (s *stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (s *stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

func main() {
	handler := &stdioHandler{}
	<-jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(&stdrwc{}, jsonrpc2.VSCodeObjectCodec{}), handler).DisconnectNotify()
}
