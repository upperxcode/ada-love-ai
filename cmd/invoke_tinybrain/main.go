package main

import (
	"context"
	"fmt"
	"time"

	"ada-love-ai/backend"
)

func main() {
	eng, err := backend.NewEngine()
	if err != nil {
		fmt.Printf("NewEngine error: %v\n", err)
		return
	}
	if eng.TinyBrainRouter() == nil {
		fmt.Println("TinyBrainRouter not initialized")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	intent, err := eng.TinyBrainRouter().DetectIntent(ctx, "me dê as 5 cidades brasileiras com o maior idh")
	if err != nil {
		fmt.Printf("DetectIntent error: %v\n", err)
		return
	}
	fmt.Printf("Detected intent: %s\n", intent)
}
