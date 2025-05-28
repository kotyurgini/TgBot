package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"tgbot/internal/ai"
	"time"
)

const ChatCompletions = "chat/completions"

type OpenAI struct {
	client http.Client
	host   string
	token  string
}

func New(host string, token string, timeout time.Duration) ai.ChatModel {
	return &OpenAI{
		client: http.Client{Timeout: timeout},
		host:   host,
		token:  token,
	}
}

func (api *OpenAI) GetStreamMessages(request ai.ChatRequest) (<-chan string, error) {
	bData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	endpoint, err := url.JoinPath(api.host, ChatCompletions)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+api.token)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Print("Ошибка: статус", resp.Status)
		bd, _ := io.ReadAll(resp.Body)
		log.Print(string(bd))
		_ = resp.Body.Close()
		return nil, errors.New("Некорректный код ответа")
	}

	ch := make(chan string)
	scanner := bufio.NewScanner(resp.Body)
	go func() {
		defer func() { _ = resp.Body.Close() }()
		defer close(ch)
		for scanner.Scan() {
			line := scanner.Bytes()

			if len(line) == 0 {
				continue
			}
			if len(line) >= 6 {
				line = line[6:]
			}
			var chunk ChatCompletionChunk
			if err = json.Unmarshal(line, &chunk); err != nil {
				log.Print("Ошибка десириализации: ", err)
				return
			}
			for _, v := range chunk.Choices {
				if v.FinishReason != "" {
					return
				}
				if v.Delta.Content != "" {
					ch <- v.Delta.Content
				}
			}
		}
	}()

	return ch, nil
}
