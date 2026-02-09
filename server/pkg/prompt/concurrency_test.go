package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// SimpleMockPrompt for testing.
type SimpleMockPrompt struct {
	NameValue    string
	ServiceValue string
}

func (p *SimpleMockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: p.NameValue}
}

func (p *SimpleMockPrompt) Service() string {
	return p.ServiceValue
}

func (p *SimpleMockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{}, nil
}

func TestPromptManager_Concurrency(t *testing.T) {
	// This test attempts to provoke race conditions by continuously adding and listing prompts.
	// While it may not deterministicly fail without the fix, it ensures that the manager
	// can handle high concurrent load without panicking or deadlocking.

	pm := NewManager()
	var wg sync.WaitGroup
	start := make(chan struct{})

	numWriters := 10
	numReaders := 10
	duration := 2 * time.Second

	// Writers: Add prompts continuously
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start

			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			timeout := time.After(duration)

			count := 0
			for {
				select {
				case <-timeout:
					return
				case <-ticker.C:
					p := &SimpleMockPrompt{
						NameValue:    fmt.Sprintf("prompt-%d-%d", id, count),
						ServiceValue: fmt.Sprintf("service-%d", id),
					}
					pm.AddPrompt(p)
					count++
				}
			}
		}(i)
	}

	// Readers: List prompts continuously
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start

			ticker := time.NewTicker(5 * time.Millisecond)
			defer ticker.Stop()

			timeout := time.After(duration)

			for {
				select {
				case <-timeout:
					return
				case <-ticker.C:
					prompts := pm.ListPrompts()
					// Basic integrity check: slice should not be nil (though empty is fine)
					if prompts == nil {
						t.Error("ListPrompts returned nil slice")
					}
					// Check for duplicates? ListPrompts implementation allows duplicates if map allows it (but map keys are unique)
					// Verify prompt names are unique?
					seen := make(map[string]bool)
					for _, p := range prompts {
						name := p.Prompt().Name
						if seen[name] {
							t.Errorf("Duplicate prompt name in list: %s", name)
						}
						seen[name] = true
					}
				}
			}
		}(i)
	}

	close(start)
	wg.Wait()

	// Final verify
	finalList := pm.ListPrompts()
	assert.NotEmpty(t, finalList)
}

func TestPromptManager_ClearPrompts_Consistency(t *testing.T) {
	// Test that clearing prompts doesn't leave stale entries visible to readers.
	pm := NewManager()
	serviceID := "service-clear"

	// Add initial prompts
	for i := 0; i < 100; i++ {
		pm.AddPrompt(&SimpleMockPrompt{
			NameValue:    fmt.Sprintf("p-%d", i),
			ServiceValue: serviceID,
		})
	}

	var wg sync.WaitGroup
	start := make(chan struct{})

	// Writer: Clears prompts
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		time.Sleep(100 * time.Millisecond)
		pm.ClearPromptsForService(serviceID)
	}()

	// Reader: Lists prompts
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start

		for i := 0; i < 50; i++ {
			prompts := pm.ListPrompts()
			// eventually it should be empty.
			if len(prompts) == 0 {
				// Success
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	close(start)
	wg.Wait()

	assert.Empty(t, pm.ListPrompts(), "Prompts should be cleared")
}
