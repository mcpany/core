package interop

import "fmt"

// Protocol represents an AI interop standard
type Protocol string

const (
	MCP Protocol = "MCP" // Model Context Protocol
	ACP Protocol = "ACP" // Agent Control Protocol
	A2A Protocol = "A2A" // Agent-to-Agent
)

// Framework represents a third-party framework
type Framework string

const (
	OpenClaw Framework = "OpenClaw"
	CrewAI   Framework = "CrewAI"
	AutoGen  Framework = "AutoGen"
)

// InteropRequest represents a standard agent request over the bus
type InteropRequest struct {
	Source    Framework
	Target    Framework
	Protocol  Protocol
	Operation string
	Payload   map[string]interface{}
}

// InteropResponse represents a standard agent response
type InteropResponse struct {
	Success bool
	Data    interface{}
	Error   error
}

// Adapter defines the Bridge pattern interface for frameworks
type Adapter interface {
	GetFramework() Framework
	SupportedProtocols() []Protocol
	TranslateRequest(req *InteropRequest) error
	Execute(req *InteropRequest) (*InteropResponse, error)
}

// BaseAdapter implements common functionality
type BaseAdapter struct {
	framework Framework
	protocols []Protocol
}

func (b *BaseAdapter) GetFramework() Framework {
	return b.framework
}

func (b *BaseAdapter) SupportedProtocols() []Protocol {
	return b.protocols
}

func (b *BaseAdapter) TranslateRequest(req *InteropRequest) error {
	// Common validation
	supported := false
	for _, p := range b.protocols {
		if p == req.Protocol {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("protocol %s not supported by %s", req.Protocol, b.framework)
	}
	return nil
}

// OpenClawAdapter
type OpenClawAdapter struct {
	BaseAdapter
}

func NewOpenClawAdapter() *OpenClawAdapter {
	return &OpenClawAdapter{
		BaseAdapter: BaseAdapter{
			framework: OpenClaw,
			protocols: []Protocol{MCP, ACP},
		},
	}
}

func (a *OpenClawAdapter) Execute(req *InteropRequest) (*InteropResponse, error) {
	if err := a.TranslateRequest(req); err != nil {
		return &InteropResponse{Success: false, Error: err}, nil
	}
	return &InteropResponse{Success: true, Data: "OpenClaw Executed: " + req.Operation}, nil
}

// CrewAIAdapter
type CrewAIAdapter struct {
	BaseAdapter
}

func NewCrewAIAdapter() *CrewAIAdapter {
	return &CrewAIAdapter{
		BaseAdapter: BaseAdapter{
			framework: CrewAI,
			protocols: []Protocol{MCP, A2A},
		},
	}
}

func (a *CrewAIAdapter) Execute(req *InteropRequest) (*InteropResponse, error) {
	if err := a.TranslateRequest(req); err != nil {
		return &InteropResponse{Success: false, Error: err}, nil
	}
	return &InteropResponse{Success: true, Data: "CrewAI Executed: " + req.Operation}, nil
}

// AutoGenAdapter
type AutoGenAdapter struct {
	BaseAdapter
}

func NewAutoGenAdapter() *AutoGenAdapter {
	return &AutoGenAdapter{
		BaseAdapter: BaseAdapter{
			framework: AutoGen,
			protocols: []Protocol{MCP, A2A, ACP},
		},
	}
}

func (a *AutoGenAdapter) Execute(req *InteropRequest) (*InteropResponse, error) {
	if err := a.TranslateRequest(req); err != nil {
		return &InteropResponse{Success: false, Error: err}, nil
	}
	return &InteropResponse{Success: true, Data: "AutoGen Executed: " + req.Operation}, nil
}
