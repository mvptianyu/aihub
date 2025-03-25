package agent

import (
	"log/slog"

	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/types"
)

// NewAgentConfig holds configuration for agent initialization
type NewAgentConfig struct {
	// The core.Provider this agent will use
	Provider core.Provider

	// Maximum number of steps before forcing stop
	MaxSteps int

	// Initial set of tools
	Tools []types.Tool

	// Initial system prompt
	SystemPrompt string

	Logger *slog.Logger

	Memory core.MemoryStorer
}
