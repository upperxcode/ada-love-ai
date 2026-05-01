package backend

import (
	"fmt"
)

// UniquifyName garante que um nome seja único dentro de um conjunto, adicionando um sufixo numérico se necessário.
func UniquifyName(name string, exists func(string) bool) string {
	if !exists(name) {
		return name
	}

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s %d", name, i)
		if !exists(candidate) {
			return candidate
		}
	}
}
