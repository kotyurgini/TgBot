package openai

import "tgbot/internal/ai"

type GPTModel string

const (
	Chat4oMini GPTModel = "gpt-4o-mini"
	ImageDalle GPTModel = "dalle"
)

type ChatCompletion struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

type Choice struct {
	Index        int        `json:"index"`
	Message      ai.Message `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens"`
	CompletionTokens        int                     `json:"completion_tokens"`
	TotalTokens             int                     `json:"total_tokens"`
	PromptTokensDetails     PromptTokensDetails     `json:"prompt_tokens_details"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
}

type ChatCompletionChunk struct {
	ID      string                      `json:"id"`
	Choices []ChatCompletionChunkChoice `json:"choices"`
}

type ChatCompletionChunkChoice struct {
	Delta        ChatCompletionChunkDelta `json:"delta"`
	FinishReason string                   `json:"finish_reason"`
}

type ChatCompletionChunkDelta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
