package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type StreamChoice struct {
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason"`
}

type StreamResponse struct {
	ID      string         `json:"id"`
	Choices []StreamChoice `json:"choices"`
}

type ChatService struct {
	apiKey     string
	model      string
	baseURL    string
	persona    string
	messages   []Message
	maxHistory int
}

type Config struct {
	APIKey  string
	Model   string
	Persona string
}

func NewChatService(cfg Config) *ChatService {
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个友好的AI助手，扮演用户的身份回答问题。\n\n用户信息：\n" + cfg.Persona + "\n\n回复要求：\n1. 以第一人称回答问题，就像你就是这个用户本人\n2. 回答要真实、专业、友好\n3. 对于你不确定的问题，如实说明\n4. 保持对话简洁有条理\n5. 如果涉及技术问题，可以适当展开说明",
		},
	}

	return &ChatService{
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
		baseURL:    "https://api.deepseek.com/v1",
		persona:    cfg.Persona,
		messages:   messages,
		maxHistory: 20,
	}
}

func (s *ChatService) Chat(ctx context.Context, userMessage string) (string, error) {
	s.messages = append(s.messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	if len(s.messages) > s.maxHistory {
		s.messages = append([]Message{s.messages[0]}, s.messages[len(s.messages)-s.maxHistory+1:]...)
	}

	reqBody := ChatRequest{
		Model:    s.model,
		Messages: s.messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误: %s", string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("没有收到回复")
	}

	assistantMessage := chatResp.Choices[0].Message.Content

	s.messages = append(s.messages, Message{
		Role:    "assistant",
		Content: assistantMessage,
	})

	return assistantMessage, nil
}

func (s *ChatService) ChatStream(ctx context.Context, userMessage string, onChunk func(string)) error {
	s.messages = append(s.messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	reqBody := ChatRequest{
		Model:    s.model,
		Messages: s.messages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误: %s", string(body))
	}

	reader := bufio.NewReader(resp.Body)
	fullContent := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			break
		}

		line = strings.TrimSpace(line)
		if line == "" || line == "data: [DONE]" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
		}

		var streamResp StreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
			chunk := streamResp.Choices[0].Delta.Content
			fullContent += chunk
			onChunk(chunk)
		}
	}

	s.messages = append(s.messages, Message{
		Role:    "assistant",
		Content: fullContent,
	})

	return nil
}

func (s *ChatService) GetHistory() []Message {
	return s.messages[1:]
}

func (s *ChatService) ClearHistory() {
	s.messages = []Message{s.messages[0]}
}

func (s *ChatService) UpdatePersona(persona string) {
	s.persona = persona
	s.messages[0] = Message{
		Role:    "system",
		Content: "你是一个友好的AI助手，扮演用户的身份回答问题。\n\n用户信息：\n" + persona + "\n\n回复要求：\n1. 以第一人称回答问题，就像你就是这个用户本人\n2. 回答要真实、专业、友好\n3. 对于你不确定的问题，如实说明\n4. 保持对话简洁有条理\n5. 如果涉及技术问题，可以适当展开说明",
	}
}
