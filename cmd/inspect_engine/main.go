package main

import (
	"encoding/json"
	"fmt"
	"os"

	"ada-love-ai/backend"
)

func main() {
	eng, err := backend.NewEngine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewEngine error: %v\n", err)
		os.Exit(2)
	}
	cfg := eng.GetAdaConfig()
	b, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Printf("ADA CONFIG:\n%s\n", string(b))
	// try listing fixed models raw
	if eng.DB() != nil {
		rows, err := eng.DB().ListFixedModelRows()
		if err != nil {
			fmt.Printf("FIXED MODELS RAW: error: %v\n", err)
		} else {
			rb, _ := json.MarshalIndent(rows, "", "  ")
			fmt.Printf("FIXED MODELS RAW:\n%s\n", string(rb))
		}
		// providers
		prov, perr := eng.DB().GetProvidersFull()
		if perr != nil {
			fmt.Printf("PROVIDERS RAW: error: %v\n", perr)
		} else {
			pb, _ := json.MarshalIndent(prov, "", "  ")
			fmt.Printf("PROVIDERS RAW:\n%s\n", string(pb))
		}
		_ = eng.DB().Close()
	}
}
