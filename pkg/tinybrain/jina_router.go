package tinybrain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// JinaRouter implements intent classification by calling a local FastAPI/Jina
// microservice that exposes a /classify endpoint using a zero-shot classifier.
// This avoids using an LLM chat call for routing and is deterministic & fast.
type JinaRouter struct {
	endpoint string
	client   *http.Client
}

// DefaultJinaEndpoint is the local classifier microservice URL.
const DefaultJinaEndpoint = "http://127.0.0.1:8008/classify"

func NewJinaRouter() *JinaRouter {
	return NewJinaRouterWithEndpoint(DefaultJinaEndpoint)
}

func NewJinaRouterWithEndpoint(endpoint string) *JinaRouter {
	if endpoint == "" {
		endpoint = DefaultJinaEndpoint
	}
	return &JinaRouter{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 2 * time.Second},
	}
}

type jinaRequest struct {
	Text            string   `json:"text"`
	CandidateLabels []string `json:"candidate_labels"`
}

type jinaResponse struct {
	Labels   []string  `json:"labels"`
	Scores   []float64 `json:"scores"`
	TopLabel string    `json:"top_label"`
}

func (r *JinaRouter) DetectIntent(ctx context.Context, text string) (Intent, error) {
	labels := []string{
		"desenvolvimento de software, programacao, go, backend, code review, banco de dados, refatoracao",
		"assunto geral, conversas casuais, cultura, geografia, historia, entretenimento",
	}

	reqBody := jinaRequest{
		Text:            text,
		CandidateLabels: labels,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return IntentGeneral, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return IntentGeneral, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return IntentGeneral, fmt.Errorf("falha ao conectar no microservico Jina: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IntentGeneral, fmt.Errorf("status de erro do microservico: %d", resp.StatusCode)
	}

	var jr jinaResponse
	if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
		return IntentGeneral, err
	}

	if jr.TopLabel == labels[0] {
		if len(jr.Scores) > 0 {
			fmt.Printf("[TinyBrain Jina] Match PROGRAMAÇÃO com %.2f%% de certeza\n", jr.Scores[0]*100)
		} else {
			fmt.Printf("[TinyBrain Jina] Match PROGRAMAÇÃO (sem score)\n")
		}
		return IntentGoProgramming, nil
	}

	if len(jr.Scores) > 0 {
		fmt.Printf("[TinyBrain Jina] Match GENERAL com %.2f%% de certeza\n", jr.Scores[0]*100)
	} else {
		fmt.Printf("[TinyBrain Jina] Match GENERAL (sem score)\n")
	}
	return IntentGeneral, nil
}
