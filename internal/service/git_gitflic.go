package service

import (
	"context"
	models "emplacc-api/internal/domain"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// GitFlicProvider — адаптер российского хостинга GitFlic (docs.gitflic.ru).
// Реализует тот же порт CommitProvider, что и GitHub — это и есть заявленная в дипломе
// независимость от конкретного хранилища кода.
//
// ВНИМАНИЕ: точные пути/поля GitFlic API при первом боевом подключении сверить с
// docs.gitflic.ru — парсер ниже намеренно толерантен (поддерживает и плоский массив,
// и HATEOAS-обёртку `_embedded.commitList`, и несколько вариантов имён полей).
type GitFlicProvider struct {
	token   string
	baseURL string
	client  *http.Client
}

func NewGitFlicProvider(token, baseURL string) *GitFlicProvider {
	if baseURL == "" {
		baseURL = "https://api.gitflic.ru"
	}
	return &GitFlicProvider{
		token:   token,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *GitFlicProvider) Name() string { return "gitflic" }

type gfCommit struct {
	Hash         string `json:"hash"`
	ID           string `json:"id"`
	Sha          string `json:"sha"`
	Message      string `json:"message"`
	ShortMessage string `json:"shortMessage"`
	WebURL       string `json:"webUrl"`
	URL          string `json:"url"`
	Author       struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Login string `json:"username"`
	} `json:"author"`
	Timestamp       *time.Time `json:"timestamp"`
	CreateTimestamp *time.Time `json:"createTimestamp"`
	Date            *time.Time `json:"date"`
}

func (c gfCommit) sha() string {
	switch {
	case c.Hash != "":
		return c.Hash
	case c.Sha != "":
		return c.Sha
	default:
		return c.ID
	}
}
func (c gfCommit) message() string {
	if c.Message != "" {
		return c.Message
	}
	return c.ShortMessage
}
func (c gfCommit) when() time.Time {
	for _, t := range []*time.Time{c.Timestamp, c.CreateTimestamp, c.Date} {
		if t != nil {
			return *t
		}
	}
	return time.Time{}
}
func (c gfCommit) webURL() string {
	if c.WebURL != "" {
		return c.WebURL
	}
	return c.URL
}

// gfResponse покрывает и плоский массив, и HATEOAS-обёртку GitFlic.
type gfResponse struct {
	Embedded struct {
		CommitList []gfCommit `json:"commitList"`
	} `json:"_embedded"`
}

const gfMaxPages = 10

func (p *GitFlicProvider) FetchCommits(ctx context.Context, repo models.CodeRepository, since *time.Time) ([]ProviderCommit, error) {
	out := make([]ProviderCommit, 0, 100)
	for page := 0; page < gfMaxPages; page++ {
		endpoint := fmt.Sprintf("%s/project/%s/%s/commit?page=%s&size=100",
			p.baseURL, repo.Owner, repo.Name, strconv.Itoa(page))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		if p.token != "" {
			req.Header.Set("Authorization", "token "+p.token)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("gitflic %s/%s: status %d: %s", repo.Owner, repo.Name, resp.StatusCode, string(body))
		}

		batch := parseGitFlicBatch(body)
		if len(batch) == 0 {
			break
		}
		for _, c := range batch {
			if c.sha() == "" {
				continue
			}
			out = append(out, ProviderCommit{
				SHA:         c.sha(),
				Message:     c.message(),
				AuthorName:  c.Author.Name,
				AuthorEmail: c.Author.Email,
				AuthorLogin: c.Author.Login,
				URL:         c.webURL(),
				CommittedAt: c.when(),
			})
		}
		if len(batch) < 100 {
			break
		}
	}
	return out, nil
}

func parseGitFlicBatch(body []byte) []gfCommit {
	// Сначала пробуем плоский массив, затем HATEOAS-обёртку.
	var flat []gfCommit
	if err := json.Unmarshal(body, &flat); err == nil && len(flat) > 0 {
		return flat
	}
	var wrapped gfResponse
	if err := json.Unmarshal(body, &wrapped); err == nil {
		return wrapped.Embedded.CommitList
	}
	return nil
}
