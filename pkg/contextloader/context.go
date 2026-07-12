package contextloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxBytes = 50 * 1024

func LoadCodeContext(targetFile string) (string, error) {
	fileInfo, err := os.Stat(targetFile)
	if err != nil {
		return "", err
	}

	if fileInfo.Size() > maxBytes {
		return "", fmt.Errorf("arquivo muito grande para injeção automática (%d bytes, max %d)", fileInfo.Size(), maxBytes)
	}

	content, err := os.ReadFile(targetFile)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(targetFile)
	lang := strings.TrimPrefix(ext, ".")

	return fmt.Sprintf("Arquivo: %s\n```%s\n%s\n```", filepath.Base(targetFile), lang, string(content)), nil
}

func LoadDirectoryContext(dir string, extensions []string) (string, error) {
	var result strings.Builder
	count := 0
	const maxFiles = 10

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if count >= maxFiles {
			return filepath.SkipDir
		}

		ext := filepath.Ext(path)
		for _, allowed := range extensions {
			if ext == allowed {
				content, err := LoadCodeContext(path)
				if err == nil {
					result.WriteString(content + "\n\n")
					count++
				}
				break
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if result.Len() == 0 {
		return "", fmt.Errorf("nenhum arquivo encontrado com as extensões: %v", extensions)
	}

	return result.String(), nil
}

func GetActiveFileContext(activeFilePath string) (string, error) {
	if activeFilePath == "" {
		return "", fmt.Errorf("nenhum arquivo ativo")
	}
	return LoadCodeContext(activeFilePath)
}
