package ai

import (
	"fmt"
	"strings"

	"github.com/ailms/backend/internal/domain/services"
	"github.com/ailms/backend/internal/infrastructure/config"
)

// NewService creates the configured AI implementation. Configuration errors are
// raised during boot instead of surfacing as opaque errors to learners later.
func NewService(aiConfig config.AIConfig, ollamaConfig config.OllamaConfig, openAIConfig config.OpenAIConfig) (services.AIService, error) {
	switch strings.ToLower(strings.TrimSpace(aiConfig.Provider)) {
	case "ollama":
		return NewOllamaClient(ollamaConfig), nil
	case "openai":
		if strings.TrimSpace(openAIConfig.APIKey) == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required when AI_PROVIDER=openai")
		}
		return NewOpenAIClient(openAIConfig), nil
	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q; use ollama or openai", aiConfig.Provider)
	}
}
