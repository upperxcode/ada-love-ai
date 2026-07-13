package orchestrator

import (
	"strings"
)

// CandidateLabels returns the default set of labels passed to the intent
// classifier. Keep these human-friendly and capitalized for backwards
// compatibility with a default HTTP classifier that expects such labels.
// CandidateLabels retorna o conjunto de labels injetando contexto semântico.
// O formato "ID: Descrição" garante que o Qwen de 1.5B entenda a intenção
// e o Python consiga extrair o ID limpo antes de responder ao Go.
func CandidateLabels() []string {
	// TODO: No futuro, carregar essa lista dinamicamente do banco de dados (Supabase)
	return []string{
		"Go: desenvolvimento de software, criação de APIs e refatoração na linguagem Go",
		"React: criação de componentes, hooks e interfaces web em React",
		"Tester: escrita de testes unitários, testes de integração e automação",
		"Geral: conversação comum, perguntas de conhecimentos gerais ou chat comum",
	}
}

// ResolveIntentCandidates maps a normalized classifier label to a list of
// candidate agent identifiers (aliases). The returned slice contains
// candidate agent IDs or name fragments that will be normalized and
// matched against the AgentLoop registry.
// ResolveIntentCandidates mapeia o ID limpo retornado pelo classificador Python
// para os identificadores internos do AgentLoop.
func ResolveIntentCandidates(label string) []string {
	if strings.TrimSpace(label) == "" {
		return nil
	}

	// O app.py novo já limpa e padroniza, mas garantimos o comportamento aqui
	l := strings.ToLower(strings.TrimSpace(label))

	// Se o roteador indicar que é GERAL, retorna nil para o chamador assumir o fluxo comum
	if l == "geral" || l == "assunto geral" || l == "general" {
		return nil
	}

	// Como o app.py retorna exatamente o ID/Chave que mandamos no CandidateLabels,
	// nós simplificamos o mapeamento para bater direto com o registro do seu AgentLoop.
	var defaults = map[string][]string{
		"react":   {"react_agent"}, // ajuste aqui para o ID exato registrado no seu AgentLoop
		"flutter": {"flutter_agent"},
		"go":      {"go_agent"},
		"tester":  {"tester_agent"},
	}

	if v, ok := defaults[l]; ok {
		return append([]string(nil), v...)
	}

	// Fallback de segurança: se você criar um agente novo no banco (ex: "SUPABASE"),
	// ele cairá aqui e tentará buscar um agente com esse nome após a normalização.
	return []string{l}
}
