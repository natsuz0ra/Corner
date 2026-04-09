package tools

import (
	"fmt"
	"slimebot/internal/logging"
	"sync"
)

// globalRegistry is the process-wide tool registry.
var (
	globalRegistry = &registry{tools: make(map[string]Tool)}
)

// registry holds registered tools by name.
type registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// Register adds a tool to the global registry; call from each tool's init().
// Panics on duplicate names so every tool name stays unique.
func Register(tool Tool) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	name := tool.Name()
	if _, exists := globalRegistry.tools[name]; exists {
		panic(fmt.Sprintf("duplicate tool name: %s", name))
	}
	globalRegistry.tools[name] = tool
	logging.Info("tool_registered", "name", name, "commands", len(tool.Commands()))
}

// Get returns a registered tool by name.
func Get(name string) (Tool, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	t, ok := globalRegistry.tools[name]
	return t, ok
}

// All returns every registered tool.
func All() []Tool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	result := make([]Tool, 0, len(globalRegistry.tools))
	for _, t := range globalRegistry.tools {
		result = append(result, t)
	}
	return result
}
