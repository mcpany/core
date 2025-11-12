
package prompt

import "github.com/modelcontextprotocol/go-sdk/mcp"

const (
	// RoleUser represents the "user" role in a prompt message.
	RoleUser = "user"
	// RoleAssistant represents the "assistant" role in a prompt message.
	RoleAssistant = "assistant"
)

const (
	// ContentTypeText represents a text content type in a prompt message.
	ContentTypeText = "text"
	// ContentTypeImage represents an image content type in a prompt message.
	ContentTypeImage = "image"
	// ContentTypeAudio represents an audio content type in a prompt message.
	ContentTypeAudio = "audio"
	// ContentTypeResource represents a resource content type in a prompt message.
	ContentTypeResource = "resource"
)

// Argument defines an argument for a prompt, including its name,
// description, and whether it is required.
type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// Message represents a single message within a prompt, including its role and
// content.
type Message struct {
	Role    string  `json:"role"`
	Content Content `json:"content"`
}

// Content is a generic interface for the different types of content that can
// be included in a prompt message.
type Content interface{}

// TextContent represents a plain text message.
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ImageContent represents an image, with the data being a base64-encoded
// string.
type ImageContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

// AudioContent represents an audio clip, with the data being a base64-encoded
// string.
type AudioContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mimeType"`
}

// ResourceContent represents a reference to a server-side resource.
type ResourceContent struct {
	Type     string        `json:"type"`
	Resource *mcp.Resource `json:"resource"`
}
