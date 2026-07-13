package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"ada-love-ai/backend"
)

func main() {
	dbPath := ""
	// replicate engine.getOSConfigDir logic for locating config dir
	switch runtime.GOOS {
	case "linux":
		dbPath = filepath.Join(os.Getenv("HOME"), ".config", "ada-love-ai")
	case "darwin":
		dbPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ada-love-ai")
	case "windows":
		dbPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "ada-love-ai")
	default:
		dbPath = "config"
	}
	os.MkdirAll(dbPath, 0755)
	path := filepath.Join(dbPath, "ada_love.db")
	fmt.Printf("Using db: %s\n", path)
	s, err := backend.NewStore(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewStore error: %v\n", err)
		os.Exit(2)
	}
	rows, err := s.ListFixedModelRows()
	if err != nil {
		fmt.Printf("ListFixedModelRows error: %v\n", err)
	} else {
		b, _ := json.MarshalIndent(rows, "", "  ")
		fmt.Printf("fixed models: %s\n", string(b))
	}
	tools, err := s.GetFixedModelRowTools(11)
	if err != nil {
		fmt.Printf("GetFixedModelRowTools error: %v\n", err)
	} else {
		fmt.Printf("tools for 11: %v\n", tools)
	}
	s.Close()
}
