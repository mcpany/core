package prompt

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestMockManagerInterface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockManagerInterface(ctrl)

	// AddPrompt
	mock.EXPECT().AddPrompt(gomock.Any()).Times(1)
	mock.AddPrompt(nil)

	// Clear
	mock.EXPECT().Clear().Times(1)
	mock.Clear()

	// ClearPromptsForService
	mock.EXPECT().ClearPromptsForService("id").Times(1)
	mock.ClearPromptsForService("id")

	// GetPrompt
	mock.EXPECT().GetPrompt("name").Return(nil, false).Times(1)
	mock.GetPrompt("name")

	// ListPrompts
	mock.EXPECT().ListPrompts().Return(nil).Times(1)
	mock.ListPrompts()

	// SetMCPServer
	mock.EXPECT().SetMCPServer(nil).Times(1)
	mock.SetMCPServer(nil)

	// UpdatePrompt
	mock.EXPECT().UpdatePrompt(gomock.Any()).Times(1)
	mock.UpdatePrompt(nil)
}
