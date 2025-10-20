package domain

// LLMRequest представляет запрос к LLM сервису
type LLMRequest struct {
    Description string
    Text        string
}

// LLMResponse представляет ответ от LLM сервиса
type LLMResponse struct {
    Result string
}