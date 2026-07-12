---
name: "Developer Guidelines"
description: "Regras de desenvolvimento e filosofia adotada pelo time"
tools: []
model: ""
maxTurns: 0
skills: []
mcpServers: []
---

KISS — Keep It Simple, Stupid

Mantenha tudo simples. Prefira soluções diretas e fáceis de entender em vez de abstrações complexas. Simplicidade facilita revisão de código, testes e manutenção a longo prazo.

Princípios práticos

- Mantenha arquivos pequenos e com responsabilidade única.
  - Cada arquivo deve ter foco claro; evite arquivos gigantes com muitas responsabilidades.
- Em Go, prefira funções/métodos pequenos e coesos.
  - Extraia funções quando um trecho de código atinge mais do que algumas linhas ou faz mais de uma coisa.
  - Coloque funções auxiliares relacionadas em arquivos separados com nomes descritivos (ex.: db_migrations.go, workspace_store.go, agent_parser.go).
- Prefira composição a herança; mantenha as interfaces pequenas e específicas.
- Escreva testes automáticos para funcionalidades críticas.
- Documente decisões arquiteturais importantes no repositório (README, docs/) — não no código apenas.
- Ao introduzir uma nova dependência, avalie custo/benefício e prefira dependências pequenas e ativas.

Boas práticas adicionais

- Nomes claros: escolha nomes descritivos para pacotes, funções e variáveis.
- Não otimize prematuramente: meça antes de otimizar.
- Faça commits pequenos e com mensagens claras — prefira revisar frequentemente.

Exemplo (Go)

- Se um serviço tem parts: inicialização, migrations, handlers → coloque cada parte em arquivo separado:
  - init.go
  - migrations.go
  - handlers.go
  - store.go

Seguindo esses princípios conseguimos código mais legível, testável e fácil de manter.
