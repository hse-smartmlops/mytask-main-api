package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"reports_llm_ms/internal/config"
	"reports_llm_ms/internal/domain"
)

// webuiLLMService реализует доменный интерфейс LLMService через WebUI API
type webuiLLMService struct {
    config *config.Config
    client *http.Client
}

// WebUI структуры запросов/ответов
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type ChatCompletionRequest struct {
    Model    string    `json:"model"`
    Stream   bool      `json:"stream"`
    Messages []Message `json:"messages"`
}

type ChatCompletionResponse struct {
    Choices []Choice `json:"choices"`
}

type Choice struct {
    Message Message `json:"message"`
}

type StreamResponse struct {
    Choices []StreamChoice `json:"choices"`
}

type StreamChoice struct {
    Delta   Message `json:"delta"`
    Message Message `json:"message"`
}

// LLMService интерфейс для работы с LLM
type LLMService interface {
    Process(ctx context.Context, req *domain.LLMRequest) (*domain.LLMResponse, error)
    ProcessStream(ctx context.Context, req *domain.LLMRequest) (<-chan *domain.LLMResponse, <-chan error)
}

// NewWebUILLMService создает новый экземпляр LLM сервиса
func NewWebUILLMService(cfg *config.Config) LLMService {
    return &webuiLLMService{
        config: cfg,
        client: &http.Client{
            Timeout: 120 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 100,
                IdleConnTimeout:     90 * time.Second,
            },
        },
    }
}

// createPayload создает payload для WebUI API
func (s *webuiLLMService) createPayload(description, text string, stream bool) ChatCompletionRequest {
    systemMessage := Message{
        Role: "system",
        Content: `Ты — работник разработчик web/ml/LLM.
Отвечай на русском языке.
я тебе гооврю контекст - описание задачи и что я написал в отчете проделанной работы за день.
ты должен на основе контекста переписать отчет более обширно. пример: написал множество вариаций lstm, catboost на предсказание дельты.
твой ответ - разработаны и протестированы различные архитектуры LSTM и модели CatBoost для предсказания дельты. Экспериментировались с параметрами LSTM (количество слоев, нейронов, функции активации, оптимизаторы) и CatBoost (глубина дерева, learning rate, регуляризация).
Проведен сравнительный анализ производительности моделей на тестовом наборе данных.
пиши специальных знаков. без пункта дальнейшие действия. текст, который ты расширил из того, что я сделал`,
    }

    userMessage := Message{
        Role:    "user",
        Content: fmt.Sprintf("описание задачи:\n%s\nчто я написал:\n%s", description, text),
    }

    return ChatCompletionRequest{
        Model:  s.config.WebUIModel,
        Stream: stream,
        Messages: []Message{
            systemMessage,
            userMessage,
        },
    }
}

// Process обрабатывает синхронный запрос к LLM
func (s *webuiLLMService) Process(ctx context.Context, req *domain.LLMRequest) (*domain.LLMResponse, error) {
    payload := s.createPayload(req.Description, req.Text, false)
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal payload: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.WebUIURL, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Authorization", "Bearer "+s.config.WebUIToken)
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
    }

    var response ChatCompletionResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    if len(response.Choices) == 0 {
        return nil, fmt.Errorf("no choices in response")
    }

    return &domain.LLMResponse{
        Result: response.Choices[0].Message.Content,
    }, nil
}

// ProcessStream обрабатывает потоковый запрос к LLM
func (s *webuiLLMService) ProcessStream(ctx context.Context, req *domain.LLMRequest) (<-chan *domain.LLMResponse, <-chan error) {
    resultChan := make(chan *domain.LLMResponse)
    errChan := make(chan error, 1)

    go func() {
        defer close(resultChan)
        defer close(errChan)

        payload := s.createPayload(req.Description, req.Text, true)
        
        jsonData, err := json.Marshal(payload)
        if err != nil {
            errChan <- fmt.Errorf("failed to marshal payload: %w", err)
            return
        }

        httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.WebUIURL, bytes.NewBuffer(jsonData))
        if err != nil {
            errChan <- fmt.Errorf("failed to create request: %w", err)
            return
        }

        httpReq.Header.Set("Authorization", "Bearer "+s.config.WebUIToken)
        httpReq.Header.Set("Content-Type", "application/json")

        resp, err := s.client.Do(httpReq)
        if err != nil {
            errChan <- fmt.Errorf("request failed: %w", err)
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            body, _ := io.ReadAll(resp.Body)
            errChan <- fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
            return
        }

        // Обрабатываем Server-Sent Events (SSE)
        reader := NewSSEReader(resp.Body)
        chunkCount := 0

        for {
            select {
            case <-ctx.Done():
                errChan <- ctx.Err()
                return
            default:
                line, err := reader.ReadEvent()
                if err != nil {
                    if err == io.EOF {
                        return
                    }
                    errChan <- fmt.Errorf("failed to read stream: %w", err)
                    return
                }

                if line == "[DONE]" {
                    return
                }

                var streamResp StreamResponse
                if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
                    continue
                }

                if len(streamResp.Choices) == 0 {
                    continue
                }

                content := ""
                if streamResp.Choices[0].Delta.Content != "" {
                    content = streamResp.Choices[0].Delta.Content
                } else if streamResp.Choices[0].Message.Content != "" {
                    content = streamResp.Choices[0].Message.Content
                }

                if content != "" {
                    chunkCount++
                    resultChan <- &domain.LLMResponse{Result: content}
                }
            }
        }
    }()

    return resultChan, errChan
}

// SSEReader для чтения Server-Sent Events
type SSEReader struct {
    reader *strings.Reader
}

func NewSSEReader(body io.Reader) *SSEReader {
    data, _ := io.ReadAll(body)
    return &SSEReader{
        reader: strings.NewReader(string(data)),
    }
}

func (r *SSEReader) ReadEvent() (string, error) {
    var line strings.Builder
    
    for {
        char, err := r.reader.ReadByte()
        if err != nil {
            return "", err
        }
        
        if char == '\n' {
            eventLine := line.String()
            line.Reset()
            
            if strings.HasPrefix(eventLine, "data: ") {
                return strings.TrimPrefix(eventLine, "data: "), nil
            }
        } else {
            line.WriteByte(char)
        }
    }
}