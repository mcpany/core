// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPage implements browserPage for testing
type MockPage struct {
	mock.Mock
}

func (m *MockPage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPage) Title() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockPage) Goto(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockPage) Click(selector string) error {
	args := m.Called(selector)
	return args.Error(0)
}

func (m *MockPage) Fill(selector, value string) error {
	args := m.Called(selector, value)
	return args.Error(0)
}

func (m *MockPage) Screenshot() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockPage) Content() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockPage) Evaluate(script string) (interface{}, error) {
	args := m.Called(script)
	return args.Get(0), args.Error(1)
}

func TestHandleToolExecution(t *testing.T) {
	t.Run("browser_navigate", func(t *testing.T) {
		mockPage := new(MockPage)
		mockPage.On("Goto", "http://example.com").Return(nil)

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		args := map[string]interface{}{"url": "http://example.com"}
		res, err := u.handleToolExecution(context.Background(), "browser_navigate", args)
		assert.NoError(t, err)
		assert.Equal(t, "navigated to http://example.com", res["result"])
		mockPage.AssertExpectations(t)
	})

	t.Run("browser_navigate_error", func(t *testing.T) {
		mockPage := new(MockPage)
		mockPage.On("Goto", "http://example.com").Return(errors.New("failed"))

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		args := map[string]interface{}{"url": "http://example.com"}
		_, err := u.handleToolExecution(context.Background(), "browser_navigate", args)
		assert.Error(t, err)
		mockPage.AssertExpectations(t)
	})

	t.Run("browser_click", func(t *testing.T) {
		mockPage := new(MockPage)
		mockPage.On("Click", "#btn").Return(nil)

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		args := map[string]interface{}{"selector": "#btn"}
		res, err := u.handleToolExecution(context.Background(), "browser_click", args)
		assert.NoError(t, err)
		assert.Equal(t, "clicked #btn", res["result"])
		mockPage.AssertExpectations(t)
	})

	t.Run("browser_fill", func(t *testing.T) {
		mockPage := new(MockPage)
		mockPage.On("Fill", "#input", "text").Return(nil)

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		args := map[string]interface{}{"selector": "#input", "value": "text"}
		res, err := u.handleToolExecution(context.Background(), "browser_fill", args)
		assert.NoError(t, err)
		assert.Equal(t, "filled #input", res["result"])
		mockPage.AssertExpectations(t)
	})

	t.Run("browser_screenshot", func(t *testing.T) {
		mockPage := new(MockPage)
		data := []byte("image data")
		mockPage.On("Screenshot").Return(data, nil)

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		res, err := u.handleToolExecution(context.Background(), "browser_screenshot", nil)
		assert.NoError(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString(data), res["image_base64"])
		mockPage.AssertExpectations(t)
	})

	t.Run("browser_content", func(t *testing.T) {
		mockPage := new(MockPage)
		content := "<html></html>"
		mockPage.On("Content").Return(content, nil)

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		res, err := u.handleToolExecution(context.Background(), "browser_content", nil)
		assert.NoError(t, err)
		assert.Equal(t, content, res["content"])
		mockPage.AssertExpectations(t)
	})

	t.Run("browser_evaluate", func(t *testing.T) {
		mockPage := new(MockPage)
		result := "result"
		mockPage.On("Evaluate", "alert(1)").Return(result, nil)

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		args := map[string]interface{}{"script": "alert(1)"}
		res, err := u.handleToolExecution(context.Background(), "browser_evaluate", args)
		assert.NoError(t, err)
		assert.Equal(t, result, res["result"])
		mockPage.AssertExpectations(t)
	})

	t.Run("unknown_tool", func(t *testing.T) {
		u := NewUpstream()
		u.SetPageForTest(new(MockPage))
		_, err := u.handleToolExecution(context.Background(), "unknown", nil)
		assert.Error(t, err)
	})

	t.Run("not_initialized", func(t *testing.T) {
		u := NewUpstream()
		_, err := u.handleToolExecution(context.Background(), "browser_navigate", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

func TestCheckHealth(t *testing.T) {
	t.Run("Healthy", func(t *testing.T) {
		mockPage := new(MockPage)
		mockPage.On("Title").Return("Page Title", nil)

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		err := u.CheckHealth(context.Background())
		assert.NoError(t, err)
		mockPage.AssertExpectations(t)
	})

	t.Run("Unhealthy", func(t *testing.T) {
		mockPage := new(MockPage)
		mockPage.On("Title").Return("", errors.New("browser crashed"))

		u := NewUpstream()
		u.SetPageForTest(mockPage)

		err := u.CheckHealth(context.Background())
		assert.Error(t, err)
		mockPage.AssertExpectations(t)
	})

	t.Run("NotInitialized", func(t *testing.T) {
		u := NewUpstream()
		err := u.CheckHealth(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}
