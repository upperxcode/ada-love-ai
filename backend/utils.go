package backend

import (
	"fmt"
	"os"
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

func writePanicLog(log string) {
	f, err := os.OpenFile("/tmp/ada-panic.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return
	}
	defer f.WriteString(log)
	f.Close()
}
