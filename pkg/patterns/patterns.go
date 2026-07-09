package patterns

import "strings"

type Pattern struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Category         string   `json:"category"`
	Group            string   `json:"group"`
	Scope            string   `json:"scope"` // empty=global, mobile/web/backend
	Description      string   `json:"description"`
	IncompatibleWith []string `json:"incompatibleWith"`
	GoldenRules      []string `json:"goldenRules"`
}

type PatternRepository struct {
	Store map[string][]Pattern
}

func NewRepository() *PatternRepository {
	repo := &PatternRepository{Store: make(map[string][]Pattern)}
	repo.setupDefaults()
	return repo
}

func (r *PatternRepository) setupDefaults() {
	architectures := []Pattern{
		{ID: "clean_architecture", Name: "Clean Architecture", Category: "Architecture", Description: "Independência de frameworks e alta testabilidade."},
		{ID: "crud", Name: "CRUD", Category: "Architecture", Description: "Manipulação direta de entidades."},
		{ID: "event_sourcing", Name: "Event Sourcing", Category: "Architecture", Description: "Histórico de eventos imutáveis."},
		{ID: "mvc", Name: "MVC", Category: "Architecture", Description: "Coordenação via Controller."},
		{ID: "mvp", Name: "MVP", Category: "Architecture", Description: "Presenter e View passiva.", Scope: "mobile"},
		{ID: "mvi", Name: "MVI", Category: "Architecture", Description: "Fluxo de dados unidirecional.", Scope: "mobile"},
		{ID: "adr", Name: "ADR", Category: "Architecture", Description: "Action-Domain-Responder.", Scope: "backend"},
		{ID: "viper", Name: "VIPER", Category: "Architecture", Description: "Separação modular extrema.", Scope: "mobile"},
		{ID: "cqrs", Name: "CQRS", Category: "Architecture", Description: "Command Query Responsibility Segregation.", Scope: "backend"},
		{ID: "custom", Name: "Custom / Service Locator", Category: "Architecture", Description: "Estrutura personalizada."},
	}
	philosophies := []Pattern{
		{ID: "solid", Name: "SOLID", Category: "Philosophy", Description: "Alta testabilidade e baixo acoplamento."},
		{ID: "dry", Name: "DRY", Category: "Philosophy", Description: "Evita duplicação de lógica."},
		{ID: "kiss", Name: "KISS", Category: "Philosophy", Description: "Mantenha o código simples."},
		{ID: "yagni", Name: "YAGNI", Category: "Philosophy", Description: "Não implemente o que não é necessário."},
	}

	designPatterns := []Pattern{
		{ID: "factory", Name: "Factory Method", Category: "DesignPattern", Group: "Creational"},
		{ID: "builder", Name: "Builder", Category: "DesignPattern", Group: "Creational"},
		{ID: "singleton", Name: "Singleton", Category: "DesignPattern", Group: "Creational"},
		{ID: "adapter", Name: "Adapter", Category: "DesignPattern", Group: "Structural"},
		{ID: "facade", Name: "Facade", Category: "DesignPattern", Group: "Structural"},
		{ID: "observer", Name: "Observer", Category: "DesignPattern", Group: "Behavioral"},
		{ID: "strategy", Name: "Strategy", Category: "DesignPattern", Group: "Behavioral"},
	}

	dataPatterns := []Pattern{
		{ID: "repository", Name: "Repository", Category: "Data", Group: "Access"},
		{ID: "dao", Name: "DAO", Category: "Data", Group: "Access"},
		{ID: "dto", Name: "DTO", Category: "Data", Group: "Representation"},
		{ID: "entity", Name: "Entity", Category: "Data", Group: "Representation"},
		{ID: "active_record", Name: "Active Record", Category: "Data", Group: "Access", Scope: "backend"},
	}

	stateManagements := []Pattern{
		{ID: "bloc", Name: "BLoC", Category: "StateManagement", Scope: "mobile"},
		{ID: "provider", Name: "Provider", Category: "StateManagement", Scope: "mobile"},
		{ID: "riverpod", Name: "Riverpod", Category: "StateManagement", Scope: "mobile"},
		{ID: "getx", Name: "GetX", Category: "StateManagement", Scope: "mobile"},
		{ID: "mobx", Name: "MobX", Category: "StateManagement", Scope: "mobile"},
		{ID: "signals", Name: "Signals", Category: "StateManagement", Scope: "mobile"},
		{ID: "redux", Name: "Redux", Category: "StateManagement", Scope: "web"},
		{ID: "context_api", Name: "Context API", Category: "StateManagement", Scope: "web"},
		{ID: "vuex", Name: "Vuex / Pinia", Category: "StateManagement", Scope: "web"},
		{ID: "none", Name: "None / Custom", Category: "StateManagement"},
	}

	dataStrategies := []Pattern{
		{ID: "sql", Name: "Relational (SQL)", Category: "DataStrategy"},
		{ID: "nosql", Name: "Non-Relational (NoSQL)", Category: "DataStrategy"},
		{ID: "remote", Name: "Remote Only (API)", Category: "DataStrategy"},
		{ID: "custom", Name: "Custom / Mixed", Category: "DataStrategy"},
	}

	allPatterns := [][]Pattern{architectures, philosophies, designPatterns, dataPatterns, stateManagements, dataStrategies}
	supportedLangs := []string{"flutter", "dart", "go", "python", "javascript", "typescript"}

	for _, lang := range supportedLangs {
		var langPatterns []Pattern
		isMobile := lang == "flutter" || lang == "dart"
		isWeb := lang == "javascript" || lang == "typescript"
		isBackend := lang == "go" || lang == "python"

		for _, group := range allPatterns {
			for _, p := range group {
				switch p.Scope {
				case "":
					langPatterns = append(langPatterns, p)
				case "mobile":
					if isMobile {
						langPatterns = append(langPatterns, p)
					}
				case "web":
					if isWeb || isMobile {
						langPatterns = append(langPatterns, p)
					}
				case "backend":
					if isBackend {
						langPatterns = append(langPatterns, p)
					}
				}
			}
		}
		r.Store[lang] = langPatterns
	}
}

func (r *PatternRepository) GetPatternsForLanguage(lang string) []Pattern {
	patterns, ok := r.Store[strings.ToLower(lang)]
	if !ok {
		return []Pattern{}
	}
	return patterns
}

// StackTemplate representa um template de stack para uma linguagem.
type StackTemplate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Libraries []Library `json:"libraries"`
}

// Library representa uma biblioteca dentro de um stack template.
type Library struct {
	Name         string `json:"name"`
	Mandatory    bool   `json:"mandatory"`
	UsageExample string `json:"usage_example"`
}

// GetStacksForLanguage retorna stack templates para a linguagem especificada.
func (r *PatternRepository) GetStacksForLanguage(lang string) []StackTemplate {
	switch strings.ToLower(lang) {
	case "flutter":
		return []StackTemplate{
			{
				ID:   "flutter_standard",
				Name: "Flutter Standard (Dio/GetIt/Bloc)",
				Libraries: []Library{
					{Name: "dio", Mandatory: true, UsageExample: "final dio = Dio();\nawait dio.get('/api');"},
					{Name: "get_it", Mandatory: true, UsageExample: "GetIt.I.registerSingleton(Service());"},
					{Name: "flutter_bloc", Mandatory: true, UsageExample: "BlocProvider(create: (_) => MyBloc());"},
				},
			},
			{
				ID:   "flutter_enterprise",
				Name: "Flutter Enterprise (Riverpod/Freezed)",
				Libraries: []Library{
					{Name: "flutter_riverpod", Mandatory: true, UsageExample: "final provider = StateProvider((ref) => 0);"},
					{Name: "freezed_annotation", Mandatory: true, UsageExample: "@freezed class User with _$User {...}"},
					{Name: "dio", Mandatory: true, UsageExample: "final response = await dio.get('/endpoint');"},
				},
			},
		}
	case "go":
		return []StackTemplate{
			{
				ID:   "go_standard",
				Name: "Go Standard (Chi/GORM)",
				Libraries: []Library{
					{Name: "chi", Mandatory: true, UsageExample: "r := chi.NewRouter()\nr.Get(\"/api\", handler)"},
					{Name: "gorm", Mandatory: true, UsageExample: "db.Create(&user)\nvar result User\nFirst(&result, 1)"},
				},
			},
			{
				ID:   "go_microservices",
				Name: "Go Microservices (gRPC/Redis)",
				Libraries: []Library{
					{Name: "grpc-go", Mandatory: true, UsageExample: "lis, _ := net.Listen(\"tcp\", \":50051\")\ngrpc.NewServer()"},
					{Name: "go-redis", Mandatory: true, UsageExample: "rdb := redis.NewClient(&redis.Options{Addr: \"localhost:6379\"})"},
				},
			},
		}
	case "python":
		return []StackTemplate{
			{
				ID:   "python_fastapi",
				Name: "Python FastAPI (SQLAlchemy/Pydantic)",
				Libraries: []Library{
					{Name: "fastapi", Mandatory: true, UsageExample: "@app.get(\"/api\")\nasync def read_items(): ..."},
					{Name: "sqlalchemy", Mandatory: true, UsageExample: "engine = create_engine(\"sqlite:///db.sqlite\")"},
					{Name: "pydantic", Mandatory: true, UsageExample: "class User(BaseModel):\n    name: str"},
				},
			},
		}
	case "javascript", "typescript":
		return []StackTemplate{
			{
				ID:   "react_standard",
				Name: "React Standard (Vite/Zustand/TanStack)",
				Libraries: []Library{
					{Name: "zustand", Mandatory: true, UsageExample: "const useStore = create((set) => ({ count: 0 }))"},
					{Name: "@tanstack/react-query", Mandatory: true, UsageExample: "useQuery({ queryKey: ['todos'], queryFn: fetchTodos })"},
					{Name: "axios", Mandatory: false, UsageExample: "const res = await axios.get('/api')"},
				},
			},
		}
	default:
		return []StackTemplate{}
	}
}

func (r *PatternRepository) GetMultiplePatternRules(lang string, patternIDs []string) map[string][]string {
	result := make(map[string][]string)
	langPatterns := r.Store[strings.ToLower(lang)]
	for _, id := range patternIDs {
		for _, p := range langPatterns {
			if p.ID == id {
				result[p.Name] = p.GoldenRules
				break
			}
		}
	}
	return result
}
