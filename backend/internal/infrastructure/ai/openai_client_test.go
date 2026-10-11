package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ailms/backend/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIClientGenerateQuizUsesResponsesStructuredOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/responses", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "gpt-5.4-mini", payload["model"])
		text, ok := payload["text"].(map[string]any)
		require.True(t, ok)
		format, ok := text["format"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "json_schema", format["type"])
		assert.Equal(t, true, format["strict"])
		schema, ok := format["schema"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "object", schema["type"])

		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"{\"questions\":[{\"question\":\"What is Go?\",\"position\":1,\"answers\":[{\"answer\":\"A language\",\"is_correct\":true},{\"answer\":\"A database\",\"is_correct\":false},{\"answer\":\"A browser\",\"is_correct\":false},{\"answer\":\"An OS\",\"is_correct\":false}]}]}"}]}]}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(config.OpenAIConfig{
		APIKey:          "test-key",
		BaseURL:         server.URL,
		Model:           "gpt-5.4-mini",
		TimeoutSeconds:  5,
		MaxOutputTokens: 200,
	})

	questions, err := client.GenerateQuiz(context.Background(), "Go is a programming language.", 1)
	require.NoError(t, err)
	require.Len(t, questions, 1)
	assert.Equal(t, "What is Go?", questions[0].Question)
	require.Len(t, questions[0].Answers, 4)
	assert.True(t, questions[0].Answers[0].IsCorrect)
}
