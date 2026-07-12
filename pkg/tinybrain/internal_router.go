package tinybrain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// InternalRouter calls a local FastAPI classifier at /classify. It's configurable
// (endpoint and candidate labels are variables) so the operator can change the
// underlying model or label set without touching the engine.
type InternalRouter struct {
	Endpoint string
	Client   *http.Client
	Labels   []string
}

// NewInternalRouter creates a router with explicit endpoint and label set.
func NewInternalRouter(endpoint string, labels []string, timeout time.Duration) *InternalRouter {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &InternalRouter{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: timeout},
		Labels:   append([]string{}, labels...),
	}
}

// NewInternalRouterDefault returns a router with sensible defaults. The label set
// is kept as a variable on the struct so callers can mutate it at runtime.
func NewInternalRouterDefault() *InternalRouter {
	labels := []string{
		"desenvolvimento de software, programacao, go, backend, code review, banco de dados, refatoracao",
		"assunto geral, conversas casuais, cultura, geografia, historia, entretenimento",
	}
	return NewInternalRouter("http://127.0.0.1:8008/classify", labels, 2*time.Second)
}

type internalRequest struct {
	Text            string   `json:"text"`
	CandidateLabels []string `json:"candidate_labels"`
}

type internalResponse struct {
	Labels   []string  `json:"labels"`
	Scores   []float64 `json:"scores"`
	TopLabel string    `json:"top_label"`
}

func (r *InternalRouter) DetectIntent(ctx context.Context, text string) (Intent, error) {
	reqBody := internalRequest{
		Text:            text,
		CandidateLabels: r.Labels,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return IntentGeneral, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return IntentGeneral, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return IntentGeneral, fmt.Errorf("falha ao conectar no microservico classifier: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return IntentGeneral, fmt.Errorf("status de erro do microservico: %d", resp.StatusCode)
	}

	var ir internalResponse
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return IntentGeneral, err
	}

	// Map top label to our Intent enums. We compare to the first label (programming)
	if ir.TopLabel == r.Labels[0] {
		if len(ir.Scores) > 0 {
			fmt.Printf("[TinyBrain Internal] Match PROGRAMAÇÃO com %.2f%% de certeza\n", ir.Scores[0]*100)
		} else {
			fmt.Printf("[TinyBrain Internal] Match PROGRAMAÇÃO (sem score)\n")
		}
		return IntentGoProgramming, nil
	}

	if len(ir.Scores) > 0 {
		fmt.Printf("[TinyBrain Internal] Match GENERAL com %.2f%% de certeza\n", ir.Scores[0]*100)
	} else {
		fmt.Printf("[TinyBrain Internal] Match GENERAL (sem score)\n")
	}
	return IntentGeneral, nil
}
