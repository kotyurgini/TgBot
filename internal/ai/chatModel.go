package ai

type ChatModel interface {
	GetStreamMessages(request ChatRequest) (<-chan string, error)
}
