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

Política de UI e Fallbacks — não mascarar erros

FALHA NÃO DEVE SER MASCARADA POR FALBACKS AUTOMÁTICOS

- Regra: não implemente fallbacks na UI que ocultem ou mascaram problemas de persistência, migração ou carga de dados. Um dropdown vazio ou um campo sem valor deve expor claramente que há um problema no backend (log + mensagem de erro amigável), e a causa raiz deve ser corrigida.

- Motivação prática:
  - Fallbacks escondem bugs e condições de corrida. Quando a UI apresenta um valor por "fallback" não é evidente que os dados primários (por ex. tabelas normalizadas no DB) não foram carregados corretamente.
  - Isso dificulta debug e promove acúmulo de dívida técnica: correções de curto prazo viram soluções permanentes inadvertidas.

- Comportamento esperado diante de dados faltantes:
  1. A UI deve indicar claramente um estado "incompleto" (ex.: mensagem ou ícone informando "Dados ausentes — ver logs"), não preencher o campo com dados de outra fonte invisível.
  2. Registrar um log no frontend e no backend com contexto suficiente: endpoint chamado, timestamp, user action, workspace/ID e qualquer payload relevante.
  3. Fornecer um caminho de correção claro (ex.: botão "Recarregar dados", instruções para reexecutar migração, ou link para a documentação de troubleshooting).

- Procedimentos de debugging que o time deve seguir (prioritários):
  1. Verificar os logs do backend na inicialização — procurar por mensagens de migração e pelos logs: "[DB] fixed_model loaded" e "[DB] SaveFixedModelRow".
  2. Consultar diretamente a tabela no DB (sqlite3) para confirmar o conteúdo de fixed_models e fixed_model_tools.
  3. Verificar se engine já concluiu a inicialização antes de servir GetAdaConfig (evitar condição de corrida).
  4. Confirmar que provider_models foram migrados para provider_models (GetProvidersFull) e que deadaptProviderConfig mapeou os models para adaCfg.Providers.

- Quando um fallback for considerado necessário (exceção):
  - Deve ser altamente visível e temporal: exibir claramente que é um fallback (UI badge "fallback"), criar um ticket automático/alerta e expirar o fallback após X minutos.
  - Preferir sempre mostrar a falha e exigir correção do backend em vez de esconder o problema.

- Checklist de implementação segura (ao adicionar qualquer comportamento que envolva exibir modelos ou dados derivados):
  - [ ] Existe logging suficiente no backend para rastrear carga/migração dos dados?
  - [ ] O frontend valida a presença explícita do dado primário antes de renderizar (ex.: adaCfg.tiny_brain.provider !== undefined)?
  - [ ] Em caso de ausência, a UI apresenta mensagem de erro e botão de recarregar, não uma lista preenchida por outro recurso invisível.
  - [ ] Há um caminho de correção (documentado em README/docs) para a operação de migração/seed que populará as tabelas normalizadas.

Seguindo essa política evitamos mascarar problemas e mantemos a observabilidade e correção das causas raiz. Se um dropdown estiver vazio, tratamos isso como um sinal de erro a ser investigado — não como um motivo para silenciar o bug com um fallback invisível.
