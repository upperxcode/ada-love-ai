package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"ada-love-ai/backend"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: set_router -endpoint <url> [-name <name>] [-labels <comma|json>]\n")
	flag.PrintDefaults()
}

func main() {
	name := flag.String("name", "default", "router config name to save (default 'default')")
	endpoint := flag.String("endpoint", "", "router endpoint URL (required)")
	labelsArg := flag.String("labels", "", "labels as comma-separated list or JSON array (e.g. 'prog,general' or '[\"a\",\"b\"]')")
	rType := flag.String("type", "http-classifier", "router type (http-classifier|llm-tinybrain)")
	backendModel := flag.String("backend", "", "backend model name (e.g. jina-embeddings-v5-text-nano-classification)")
	flag.Usage = usage
	flag.Parse()

	if strings.TrimSpace(*endpoint) == "" {
		fmt.Fprintln(os.Stderr, "-endpoint is required")
		usage()
		os.Exit(2)
	}

	var labels []string
	if strings.TrimSpace(*labelsArg) == "" {
		// Use defaults if not provided
		labels = []string{
			"desenvolvimento de software, programacao, go, backend, code review, banco de dados, refatoracao",
			"assunto geral, conversas casuais, cultura, geografia, historia, entretenimento",
		}
	} else {
		// Try JSON first
		s := strings.TrimSpace(*labelsArg)
		if strings.HasPrefix(s, "[") {
			if err := json.Unmarshal([]byte(s), &labels); err != nil {
				fmt.Fprintf(os.Stderr, "failed to parse labels JSON: %v\n", err)
				os.Exit(2)
			}
		} else {
			// comma-separated
			parts := strings.Split(s, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			labels = parts
		}
	}

	eng, err := backend.NewEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create engine: %v\n", err)
		os.Exit(2)
	}

	if err := eng.SetRouterConfig(*name, *endpoint, labels, *rType, *backendModel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set router config: %v\n", err)
		// try to close DB if available
		if eng.DB() != nil {
			eng.DB().Close()
		}
		os.Exit(2)
	}

	fmt.Printf("router config %q set -> %s (labels=%d)\n", *name, *endpoint, len(labels))

	// close DB cleanly before exit
	if eng.DB() != nil {
		eng.DB().Close()
	}
}
