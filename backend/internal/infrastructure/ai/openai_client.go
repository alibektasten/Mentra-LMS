package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ailms/backend/internal/domain/entities"
	"github.com/ailms/backend/internal/infrastructure/config"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/ailms/backend/pkg/logger"
	"github.com/google/uuid"
)

// OpenAIClient uses the Responses API. It stays behind the existing AIService
// interface so no API key or provider-specific logic reaches the frontend.
type OpenAIClient struct {
	apiKey          string
	baseURL         string
	model           string
	maxOutputTokens int
	httpClient      *http.Client
}

type openAIResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Type    string `json:"type"`
		Content []struct {
			Type    string `json:"type"`
			Text    string `json:"text"`
			Refusal string `json:"refusal"`
		} `json:"content"`
	} `json:"output"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
	Usage *struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenAIClient(cfg config.OpenAIConfig) *OpenAIClient {
	return &OpenAIClient{
		apiKey:          cfg.APIKey,
		baseURL:         strings.TrimRight(cfg.BaseURL, "/"),
		model:           cfg.Model,
		maxOutputTokens: cfg.MaxOutputTokens,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
	}
}

func (c *OpenAIClient) generate(ctx context.Context, instructions, input string, schema map[string]any) (string, error) {
	startedAt := time.Now()
	payload := map[string]any{
		"model":             c.model,
		"instructions":      instructions,
		"input":             input,
		"max_output_tokens": c.maxOutputTokens,
		"reasoning": map[string]string{
			"effort": "none",
		},
	}
	if schema != nil {
		payload["text"] = map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "lms_ai_response",
				"strict": true,
				"schema": schema,
			},
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", apperrors.InternalError("failed to marshal OpenAI request")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/responses", bytes.NewReader(data))
	if err != nil {
		return "", apperrors.InternalError("failed to create OpenAI request")
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", apperrors.InternalError(fmt.Sprintf("OpenAI service unreachable: %s", err.Error()))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperrors.InternalError("failed to read OpenAI response")
	}
	var response openAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", apperrors.InternalError("failed to parse OpenAI response")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if response.Error != nil && response.Error.Message != "" {
			return "", apperrors.InternalError("OpenAI service error: " + response.Error.Message)
		}
		return "", apperrors.InternalError(fmt.Sprintf("OpenAI service returned status %d", resp.StatusCode))
	}
	if response.Usage != nil {
		logger.Default().Info("ai_provider_usage",
			"provider", "openai",
			"model", c.model,
			"input_tokens", response.Usage.InputTokens,
			"output_tokens", response.Usage.OutputTokens,
			"total_tokens", response.Usage.TotalTokens,
			"latency_ms", time.Since(startedAt).Milliseconds(),
		)
	}
	if response.OutputText != "" {
		return response.OutputText, nil
	}
	for _, output := range response.Output {
		for _, content := range output.Content {
			if content.Type == "output_text" && content.Text != "" {
				return content.Text, nil
			}
			if content.Type == "refusal" && content.Refusal != "" {
				return "", apperrors.InternalError("OpenAI declined this request: " + content.Refusal)
			}
		}
	}
	return "", apperrors.InternalError("OpenAI returned no text output")
}

func (c *OpenAIClient) GenerateLessonSummary(ctx context.Context, lessonContent string) (string, error) {
	return c.generate(ctx,
		"You are a helpful learning assistant. Write concise, accurate study material.",
		fmt.Sprintf("Summarize this lesson in 2-3 concise paragraphs. Return only the summary.\n\n%s", lessonContent), nil)
}

func (c *OpenAIClient) GenerateQuiz(ctx context.Context, lessonContent string, numQuestions int) ([]entities.QuizQuestion, error) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"questions"},
		"properties": map[string]any{
			"questions": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"question", "position", "answers"},
					"properties": map[string]any{
						"question": map[string]any{"type": "string"},
						"position": map[string]any{"type": "integer"},
						"answers": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type":                 "object",
								"additionalProperties": false,
								"required":             []string{"answer", "is_correct"},
								"properties": map[string]any{
									"answer":     map[string]any{"type": "string"},
									"is_correct": map[string]any{"type": "boolean"},
								},
							},
						},
					},
				},
			},
		},
	}
	response, err := c.generate(ctx,
		"You generate fair, factually accurate multiple-choice quizzes for students. Use exactly four answers per question and exactly one correct answer.",
		fmt.Sprintf("Create exactly %d questions based only on this lesson. Return them in the questions field:\n\n%s", numQuestions, lessonContent), schema)
	if err != nil {
		return nil, err
	}

	type answerJSON struct {
		Answer    string `json:"answer"`
		IsCorrect bool   `json:"is_correct"`
	}
	type questionJSON struct {
		Question string       `json:"question"`
		Position int          `json:"position"`
		Answers  []answerJSON `json:"answers"`
	}
	var payload struct {
		Questions []questionJSON `json:"questions"`
	}
	if err := json.Unmarshal([]byte(response), &payload); err != nil {
		return nil, apperrors.InternalError("failed to parse OpenAI quiz response")
	}
	questions := payload.Questions

	now := time.Now()
	result := make([]entities.QuizQuestion, len(questions))
	for i, question := range questions {
		answers := make([]entities.QuizAnswer, len(question.Answers))
		for j, answer := range question.Answers {
			answers[j] = entities.QuizAnswer{ID: uuid.New(), Answer: answer.Answer, IsCorrect: answer.IsCorrect, CreatedAt: now, UpdatedAt: now}
		}
		result[i] = entities.QuizQuestion{ID: uuid.New(), Question: question.Question, Position: question.Position, Answers: answers, CreatedAt: now, UpdatedAt: now}
	}
	return result, nil
}

func (c *OpenAIClient) GenerateRemediation(ctx context.Context, lessonContent string, wrongQuestions []string) (string, error) {
	return c.generate(ctx,
		"You are an encouraging learning assistant. Be specific, concise, and accurate.",
		fmt.Sprintf("A student missed these questions:\n%s\nLesson:\n%s\n\nIn 2-3 sentences, explain what to review and how to approach it.", strings.Join(wrongQuestions, "\n"), lessonContent), nil)
}

func (c *OpenAIClient) GenerateAssignmentFeedback(ctx context.Context, assignmentTitle, description, submissionText string) (*entities.AssignmentFeedback, error) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"strengths", "gaps", "improvements", "overall"},
		"properties": map[string]any{
			"strengths":    stringArraySchema(),
			"gaps":         stringArraySchema(),
			"improvements": stringArraySchema(),
			"overall":      map[string]any{"type": "string"},
		},
	}
	response, err := c.generate(ctx,
		"You give constructive academic feedback. Do not assign a grade, invent facts, or make personal judgements.",
		fmt.Sprintf("Assignment title: %s\nAssignment description: %s\n\nStudent submission:\n%s", assignmentTitle, description, submissionText), schema)
	if err != nil {
		return nil, err
	}
	type feedbackJSON struct {
		Strengths    []string `json:"strengths"`
		Gaps         []string `json:"gaps"`
		Improvements []string `json:"improvements"`
		Overall      string   `json:"overall"`
	}
	var feedback feedbackJSON
	if err := json.Unmarshal([]byte(response), &feedback); err != nil {
		return nil, apperrors.InternalError("failed to parse OpenAI feedback response")
	}
	return &entities.AssignmentFeedback{Strengths: feedback.Strengths, Gaps: feedback.Gaps, Improvements: feedback.Improvements, Overall: feedback.Overall}, nil
}

func (c *OpenAIClient) GenerateFlashcards(ctx context.Context, lessonContent string, numCards int) ([]entities.Flashcard, error) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"cards"},
		"properties": map[string]any{
			"cards": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"required":             []string{"term", "definition"},
					"properties": map[string]any{
						"term":       map[string]any{"type": "string"},
						"definition": map[string]any{"type": "string"},
					},
				},
			},
		},
	}
	response, err := c.generate(ctx,
		"You create concise, accurate study flashcards. Avoid duplicating concepts.",
		fmt.Sprintf("Create exactly %d flashcards from this lesson. Return them in the cards field:\n\n%s", numCards, lessonContent), schema)
	if err != nil {
		return nil, err
	}
	type flashcardJSON struct {
		Term       string `json:"term"`
		Definition string `json:"definition"`
	}
	var payload struct {
		Cards []flashcardJSON `json:"cards"`
	}
	if err := json.Unmarshal([]byte(response), &payload); err != nil {
		return nil, apperrors.InternalError("failed to parse OpenAI flashcard response")
	}
	cards := payload.Cards
	result := make([]entities.Flashcard, len(cards))
	for i, card := range cards {
		result[i] = entities.Flashcard{Term: card.Term, Definition: card.Definition}
	}
	return result, nil
}

func (c *OpenAIClient) GenerateProgressInsights(ctx context.Context, progress []entities.LessonProgress) (string, error) {
	completed := 0
	var totalScore float64
	scored := 0
	for _, item := range progress {
		if item.CompletedAt != nil {
			completed++
		}
		if item.Score != nil {
			totalScore += *item.Score
			scored++
		}
	}
	average := 0.0
	if scored > 0 {
		average = totalScore / float64(scored)
	}
	return c.generate(ctx,
		"You are an encouraging learning assistant. Give specific, concise study guidance without making claims beyond the supplied progress.",
		fmt.Sprintf("A student completed %d of %d lessons and has an average score of %.1f%%. Provide 2-3 personalized study insights.", completed, len(progress), average), nil)
}

func stringArraySchema() map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}}
}
