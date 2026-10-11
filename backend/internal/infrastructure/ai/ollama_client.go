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
	"github.com/google/uuid"
)

type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

type ollamaRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	Stream    bool   `json:"stream"`
	KeepAlive string `json:"keep_alive,omitempty"`
	Think     *bool  `json:"think,omitempty"`
}

// extractJSON strips markdown code fences that LLMs sometimes wrap around JSON,
// then extracts the first JSON object or array from the response.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// Strip ```json ... ``` or ``` ... ```
	if strings.HasPrefix(s, "```") {
		end := strings.LastIndex(s, "```")
		if end > 3 {
			s = s[3:end]
			// Remove optional language tag on first line
			if nl := strings.Index(s, "\n"); nl != -1 {
				s = s[nl+1:]
			}
			s = strings.TrimSpace(s)
		}
	}
	// Find the first [ or { and the matching last ] or }
	start := strings.IndexAny(s, "[{")
	if start == -1 {
		return s
	}
	open := rune(s[start])
	close := map[rune]rune{'[': ']', '{': '}'}[open]
	end := strings.LastIndexByte(s, byte(close))
	if end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func NewOllamaClient(cfg config.OllamaConfig) *OllamaClient {
	return &OllamaClient{
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
	}
}

func (c *OllamaClient) generate(ctx context.Context, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:     c.model,
		Prompt:    prompt,
		Stream:    false,
		KeepAlive: "1h",
	}
	// Qwen 3 enables a long reasoning phase by default. Disabling it for these
	// short, structured LMS actions keeps the local fallback responsive.
	if strings.HasPrefix(strings.ToLower(c.model), "qwen3") {
		think := false
		reqBody.Think = &think
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", apperrors.InternalError("failed to marshal ollama request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(data))
	if err != nil {
		return "", apperrors.InternalError("failed to create ollama request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", apperrors.InternalError(fmt.Sprintf("AI service unreachable: %s", err.Error()))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperrors.InternalError("failed to read AI response")
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", apperrors.InternalError("failed to parse AI response")
	}
	if ollamaResp.Error != "" {
		return "", apperrors.InternalError(fmt.Sprintf("AI service error: %s", ollamaResp.Error))
	}
	if resp.StatusCode != 200 {
		return "", apperrors.InternalError(fmt.Sprintf("AI service returned status %d", resp.StatusCode))
	}
	return ollamaResp.Response, nil
}

// Warmup sends a minimal request to Ollama to pre-load the model into memory,
// avoiding a cold-start delay on the first real request.
func (c *OllamaClient) Warmup(ctx context.Context) {
	_, _ = c.generate(ctx, "hi")
}

func (c *OllamaClient) GenerateLessonSummary(ctx context.Context, lessonContent string) (string, error) {
	prompt := fmt.Sprintf(`Summarize the following lesson content in 2-3 concise paragraphs:

%s

Provide only the summary, no additional commentary.`, lessonContent)
	return c.generate(ctx, prompt)
}

func (c *OllamaClient) GenerateQuiz(ctx context.Context, lessonContent string, numQuestions int) ([]entities.QuizQuestion, error) {
	prompt := fmt.Sprintf(`Based on the following lesson content, generate %d quiz questions in JSON format.
Each question should have 4 answer options with exactly one correct answer.

Lesson content:
%s

Return ONLY valid JSON in this exact format:
[
  {
    "question": "Question text here",
    "position": 1,
    "answers": [
      {"answer": "Option A", "is_correct": false},
      {"answer": "Option B", "is_correct": true},
      {"answer": "Option C", "is_correct": false},
      {"answer": "Option D", "is_correct": false}
    ]
  }
]`, numQuestions, lessonContent)

	response, err := c.generate(ctx, prompt)
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

	var questions []questionJSON
	if err := json.Unmarshal([]byte(extractJSON(response)), &questions); err != nil {
		return nil, apperrors.InternalError("failed to parse AI quiz response")
	}

	now := time.Now()
	result := make([]entities.QuizQuestion, len(questions))
	for i, q := range questions {
		answers := make([]entities.QuizAnswer, len(q.Answers))
		for j, a := range q.Answers {
			answers[j] = entities.QuizAnswer{
				ID:        uuid.New(),
				Answer:    a.Answer,
				IsCorrect: a.IsCorrect,
				CreatedAt: now,
				UpdatedAt: now,
			}
		}
		result[i] = entities.QuizQuestion{
			ID:        uuid.New(),
			Question:  q.Question,
			Position:  q.Position,
			Answers:   answers,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return result, nil
}

func (c *OllamaClient) GenerateRemediation(ctx context.Context, lessonContent string, wrongQuestions []string) (string, error) {
	questionsText := ""
	for i, q := range wrongQuestions {
		questionsText += fmt.Sprintf("%d. %s\n", i+1, q)
	}
	prompt := fmt.Sprintf(`A student answered the following quiz questions incorrectly:

%s
Lesson content they studied:
%s

In 2-3 sentences, explain which concepts to review and how to approach them. Be specific and encouraging.`,
		questionsText, lessonContent)
	return c.generate(ctx, prompt)
}

func (c *OllamaClient) GenerateAssignmentFeedback(ctx context.Context, assignmentTitle, description, submissionText string) (*entities.AssignmentFeedback, error) {
	prompt := fmt.Sprintf(`You are reviewing a student's assignment submission.

Assignment title: %s
Assignment description: %s

Student submission:
%s

Provide structured feedback as JSON only:
{
  "strengths": ["point 1", "point 2"],
  "gaps": ["gap 1", "gap 2"],
  "improvements": ["suggestion 1", "suggestion 2"],
  "overall": "2-3 sentence summary"
}

Return ONLY valid JSON, no other text.`, assignmentTitle, description, submissionText)

	response, err := c.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	type feedbackJSON struct {
		Strengths    []string `json:"strengths"`
		Gaps         []string `json:"gaps"`
		Improvements []string `json:"improvements"`
		Overall      string   `json:"overall"`
	}
	var fb feedbackJSON
	if err := json.Unmarshal([]byte(extractJSON(response)), &fb); err != nil {
		return nil, apperrors.InternalError("failed to parse AI feedback response")
	}
	return &entities.AssignmentFeedback{
		Strengths:    fb.Strengths,
		Gaps:         fb.Gaps,
		Improvements: fb.Improvements,
		Overall:      fb.Overall,
	}, nil
}

func (c *OllamaClient) GenerateFlashcards(ctx context.Context, lessonContent string, numCards int) ([]entities.Flashcard, error) {
	prompt := fmt.Sprintf(`Based on the following lesson content, generate %d flashcards as JSON.

Lesson content:
%s

Return ONLY valid JSON in this exact format:
[
  {"term": "Term or concept", "definition": "Clear, concise definition"}
]`, numCards, lessonContent)

	response, err := c.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	type flashcardJSON struct {
		Term       string `json:"term"`
		Definition string `json:"definition"`
	}
	var cards []flashcardJSON
	if err := json.Unmarshal([]byte(extractJSON(response)), &cards); err != nil {
		return nil, apperrors.InternalError("failed to parse AI flashcard response")
	}
	result := make([]entities.Flashcard, len(cards))
	for i, card := range cards {
		result[i] = entities.Flashcard{Term: card.Term, Definition: card.Definition}
	}
	return result, nil
}

func (c *OllamaClient) GenerateProgressInsights(ctx context.Context, progress []entities.LessonProgress) (string, error) {
	completedCount := 0
	var totalScore float64
	var scoreCount int
	for _, p := range progress {
		if p.CompletedAt != nil {
			completedCount++
		}
		if p.Score != nil {
			totalScore += *p.Score
			scoreCount++
		}
	}
	avgScore := 0.0
	if scoreCount > 0 {
		avgScore = totalScore / float64(scoreCount)
	}

	prompt := fmt.Sprintf(`A student has completed %d lessons out of %d total tracked, with an average score of %.1f%%.
Provide 2-3 sentences of personalized learning insights and encouragement based on this progress.`,
		completedCount, len(progress), avgScore)

	return c.generate(ctx, prompt)
}
