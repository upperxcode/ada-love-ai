export namespace backend {
	
	export class WorkspaceTemplate {
	    id: number;
	    name: string;
	    description: string;
	    personality: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceTemplate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.personality = source["personality"];
	        this.created_at = source["created_at"];
	    }
	}
	export class StackItem {
	    name: string;
	    example?: string;
	
	    static createFrom(source: any = {}) {
	        return new StackItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.example = source["example"];
	    }
	}
	export class SpecWizardConfig {
	    id: string;
	    name: string;
	    description?: string;
	    expert_language_plugin?: string;
	    prd?: string;
	    functional_requirements?: string[];
	    non_functional_requirements?: string[];
	    persistence?: string;
	    architecture?: string;
	    engineering_philosophies?: string[];
	    design_patterns?: string[];
	    data_patterns?: string[];
	    stack_config?: StackItem[];
	    business_state_management?: string;
	    business_api_contract?: string;
	    business_customization_details?: string;
	    business_final_adjustments?: string;
	    business_architecture_recommendations?: string;
	    color: string;
	    icon: string;
	    architecture_health: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new SpecWizardConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.expert_language_plugin = source["expert_language_plugin"];
	        this.prd = source["prd"];
	        this.functional_requirements = source["functional_requirements"];
	        this.non_functional_requirements = source["non_functional_requirements"];
	        this.persistence = source["persistence"];
	        this.architecture = source["architecture"];
	        this.engineering_philosophies = source["engineering_philosophies"];
	        this.design_patterns = source["design_patterns"];
	        this.data_patterns = source["data_patterns"];
	        this.stack_config = this.convertValues(source["stack_config"], StackItem);
	        this.business_state_management = source["business_state_management"];
	        this.business_api_contract = source["business_api_contract"];
	        this.business_customization_details = source["business_customization_details"];
	        this.business_final_adjustments = source["business_final_adjustments"];
	        this.business_architecture_recommendations = source["business_architecture_recommendations"];
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.architecture_health = source["architecture_health"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MCPServerUI {
	    command: string;
	    args?: string[];
	    env?: Record<string, string>;
	    url: string;
	    enabled: boolean;
	    icon: string;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new MCPServerUI(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.args = source["args"];
	        this.env = source["env"];
	        this.url = source["url"];
	        this.enabled = source["enabled"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	    }
	}
	export class ToolProfile {
	    id: number;
	    name: string;
	    description?: string;
	    color: string;
	    icon: string;
	    tools: string[];
	
	    static createFrom(source: any = {}) {
	        return new ToolProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.tools = source["tools"];
	    }
	}
	export class ModelSettings {
	    context_size?: number;
	    temperature?: number;
	    max_tokens?: number;
	    top_p?: number;
	    type?: string;
	    vision?: boolean;
	    embedding?: boolean;
	    tools?: boolean;
	    free?: boolean;
	    thinking?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.context_size = source["context_size"];
	        this.temperature = source["temperature"];
	        this.max_tokens = source["max_tokens"];
	        this.top_p = source["top_p"];
	        this.type = source["type"];
	        this.vision = source["vision"];
	        this.embedding = source["embedding"];
	        this.tools = source["tools"];
	        this.free = source["free"];
	        this.thinking = source["thinking"];
	    }
	}
	export class ProviderApiKey {
	    key: string;
	    user_key?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderApiKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.user_key = source["user_key"];
	    }
	}
	export class ProviderConfig {
	    icon: string;
	    color: string;
	    api_url: string;
	    api_key?: string;
	    api_keys?: ProviderApiKey[];
	    type_connection: string;
	    models: Record<string, ModelSettings>;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.api_url = source["api_url"];
	        this.api_key = source["api_key"];
	        this.api_keys = this.convertValues(source["api_keys"], ProviderApiKey);
	        this.type_connection = source["type_connection"];
	        this.models = this.convertValues(source["models"], ModelSettings, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExtraModelConfig {
	    context_size: number;
	    temperature: number;
	    max_tokens: number;
	    top_p: number;
	
	    static createFrom(source: any = {}) {
	        return new ExtraModelConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.context_size = source["context_size"];
	        this.temperature = source["temperature"];
	        this.max_tokens = source["max_tokens"];
	        this.top_p = source["top_p"];
	    }
	}
	export class SubagentsConfig {
	    allow_agents?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SubagentsConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.allow_agents = source["allow_agents"];
	    }
	}
	export class AgentConfig {
	    id: number;
	    name: string;
	    description: string;
	    provider: string;
	    model: string;
	    provider_id?: number;
	    model_id?: number;
	    type: string;
	    icon: string;
	    color: string;
	    max_iterations?: number;
	    temperature?: number;
	    delegates?: string[];
	    system_prompt?: string;
	    subagents?: SubagentsConfig;
	
	    static createFrom(source: any = {}) {
	        return new AgentConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.provider_id = source["provider_id"];
	        this.model_id = source["model_id"];
	        this.type = source["type"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.max_iterations = source["max_iterations"];
	        this.temperature = source["temperature"];
	        this.delegates = source["delegates"];
	        this.system_prompt = source["system_prompt"];
	        this.subagents = this.convertValues(source["subagents"], SubagentsConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkerConfig {
	    id: number;
	    name: string;
	    persona: string;
	    response_language: string;
	    icon: string;
	    color: string;
	    connection_type: string;
	    command: string;
	    arguments: string;
	    environment: string;
	    connection_name?: string;
	    connection_config?: string;
	    language?: string;
	    inherit_folders: boolean;
	    inherit_knowledge: boolean;
	    inherit_skills: boolean;
	    inherit_tools: boolean;
	    inherit_persona: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WorkerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.persona = source["persona"];
	        this.response_language = source["response_language"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.connection_type = source["connection_type"];
	        this.command = source["command"];
	        this.arguments = source["arguments"];
	        this.environment = source["environment"];
	        this.connection_name = source["connection_name"];
	        this.connection_config = source["connection_config"];
	        this.language = source["language"];
	        this.inherit_folders = source["inherit_folders"];
	        this.inherit_knowledge = source["inherit_knowledge"];
	        this.inherit_skills = source["inherit_skills"];
	        this.inherit_tools = source["inherit_tools"];
	        this.inherit_persona = source["inherit_persona"];
	    }
	}
	export class WorkspaceConfig {
	    id: number;
	    nome: string;
	    description: string;
	    max_prompt: number;
	    max_content: number;
	    commit: boolean;
	    spec_provider: string;
	    spec_wizard_id: string;
	    personality: string;
	    routing_rules: string;
	    color: string;
	    icon: string;
	    title: string;
	    summary: string;
	    path: string;
	    folders: string[];
	    knowledge: string[];
	    worker_names: string[];
	    agents: string[];
	    skills: string[];
	    tools: string[];
	    enabled: boolean;
	    max_prompt_send: number;
	    commit_changes: boolean;
	    max_context_length: number;
	    spec_wizard: string;
	    embedding_model: string;
	    embedding_provider: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nome = source["nome"];
	        this.description = source["description"];
	        this.max_prompt = source["max_prompt"];
	        this.max_content = source["max_content"];
	        this.commit = source["commit"];
	        this.spec_provider = source["spec_provider"];
	        this.spec_wizard_id = source["spec_wizard_id"];
	        this.personality = source["personality"];
	        this.routing_rules = source["routing_rules"];
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.path = source["path"];
	        this.folders = source["folders"];
	        this.knowledge = source["knowledge"];
	        this.worker_names = source["worker_names"];
	        this.agents = source["agents"];
	        this.skills = source["skills"];
	        this.tools = source["tools"];
	        this.enabled = source["enabled"];
	        this.max_prompt_send = source["max_prompt_send"];
	        this.commit_changes = source["commit_changes"];
	        this.max_context_length = source["max_context_length"];
	        this.spec_wizard = source["spec_wizard"];
	        this.embedding_model = source["embedding_model"];
	        this.embedding_provider = source["embedding_provider"];
	    }
	}
	export class AdaConfig {
	    active_workspace_path: string;
	    active_workspace_index: number;
	    workspaces: WorkspaceConfig[];
	    // Go type: struct { ModelName string "json:\"model_name\""; Provider string "json:\"provider\""; EmbeddingModel string "json:\"embedding_model\""; EmbeddingProvider string "json:\"embedding_provider\""; Tools []string "json:\"tools\"" }
	    tiny_brain: any;
	    // Go type: struct { ModelName string "json:\"model_name\""; Provider string "json:\"provider\""; Tools []string "json:\"tools\"" }
	    classifier: any;
	    embedding_model: string;
	    embedding_provider: string;
	    image_model: string;
	    image_provider: string;
	    spec_model: string;
	    spec_provider: string;
	    spec_tools: string[];
	    workers: WorkerConfig[];
	    worker_categories: string[];
	    agents: AgentConfig[];
	    agent_categories: string[];
	    provider_keys: Record<string, string>;
	    provider_bases: Record<string, string>;
	    model_settings: Record<string, ExtraModelConfig>;
	    model_list: config.ModelConfig[];
	    providers?: Record<string, ProviderConfig>;
	    tool_profiles?: ToolProfile[];
	    mcp_servers?: Record<string, MCPServerUI>;
	    spec_wizards: SpecWizardConfig[];
	    templates: WorkspaceTemplate[];
	
	    static createFrom(source: any = {}) {
	        return new AdaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.active_workspace_path = source["active_workspace_path"];
	        this.active_workspace_index = source["active_workspace_index"];
	        this.workspaces = this.convertValues(source["workspaces"], WorkspaceConfig);
	        this.tiny_brain = this.convertValues(source["tiny_brain"], Object);
	        this.classifier = this.convertValues(source["classifier"], Object);
	        this.embedding_model = source["embedding_model"];
	        this.embedding_provider = source["embedding_provider"];
	        this.image_model = source["image_model"];
	        this.image_provider = source["image_provider"];
	        this.spec_model = source["spec_model"];
	        this.spec_provider = source["spec_provider"];
	        this.spec_tools = source["spec_tools"];
	        this.workers = this.convertValues(source["workers"], WorkerConfig);
	        this.worker_categories = source["worker_categories"];
	        this.agents = this.convertValues(source["agents"], AgentConfig);
	        this.agent_categories = source["agent_categories"];
	        this.provider_keys = source["provider_keys"];
	        this.provider_bases = source["provider_bases"];
	        this.model_settings = this.convertValues(source["model_settings"], ExtraModelConfig, true);
	        this.model_list = this.convertValues(source["model_list"], config.ModelConfig);
	        this.providers = this.convertValues(source["providers"], ProviderConfig, true);
	        this.tool_profiles = this.convertValues(source["tool_profiles"], ToolProfile);
	        this.mcp_servers = this.convertValues(source["mcp_servers"], MCPServerUI, true);
	        this.spec_wizards = this.convertValues(source["spec_wizards"], SpecWizardConfig);
	        this.templates = this.convertValues(source["templates"], WorkspaceTemplate);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ToolCall {
	    id: string;
	    type: string;
	    // Go type: struct { Name string "json:\"name\""; Arguments string "json:\"arguments\"" }
	    function: any;
	
	    static createFrom(source: any = {}) {
	        return new ToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.function = this.convertValues(source["function"], Object);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChatMessage {
	    id: number;
	    role: string;
	    content: string;
	    tool_calls?: ToolCall[];
	    tool_call_id?: string;
	    // Go type: time
	    time: any;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.tool_calls = this.convertValues(source["tool_calls"], ToolCall);
	        this.tool_call_id = source["tool_call_id"];
	        this.time = this.convertValues(source["time"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ChatSession {
	    id: string;
	    workspace_id: string;
	    worker_name: string;
	    parent_session_id: string;
	    title: string;
	    summary: string;
	    model: string;
	    provider: string;
	    mode: string;
	    thinking: string;
	    summarized_context: string;
	    // Go type: time
	    summarized_at: any;
	    last_summarized_msg_id: number;
	    messages: ChatMessage[];
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    pinned: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ChatSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspace_id = source["workspace_id"];
	        this.worker_name = source["worker_name"];
	        this.parent_session_id = source["parent_session_id"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.model = source["model"];
	        this.provider = source["provider"];
	        this.mode = source["mode"];
	        this.thinking = source["thinking"];
	        this.summarized_context = source["summarized_context"];
	        this.summarized_at = this.convertValues(source["summarized_at"], null);
	        this.last_summarized_msg_id = source["last_summarized_msg_id"];
	        this.messages = this.convertValues(source["messages"], ChatMessage);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.pinned = source["pinned"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConnectionDefinition {
	    name: string;
	    type: string;
	    command: string;
	    description: string;
	    icon: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.command = source["command"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	    }
	}
	export class ConnectionTestResult {
	    success: boolean;
	    message: string;
	    latency_ms: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.latency_ms = source["latency_ms"];
	    }
	}
	
	export class FixedModel {
	    id: number;
	    name: string;
	    provider: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new FixedModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	    }
	}
	
	
	
	
	export class ProviderModel {
	    id: string;
	    name: string;
	    vision?: boolean;
	    embedding?: boolean;
	    tools?: boolean;
	    free?: boolean;
	    thinking?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProviderModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.vision = source["vision"];
	        this.embedding = source["embedding"];
	        this.tools = source["tools"];
	        this.free = source["free"];
	        this.thinking = source["thinking"];
	    }
	}
	export class ProviderTestResult {
	    ok: boolean;
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class SearchResult {
	    name: string;
	    display_name: string;
	    registry_name: string;
	    summary: string;
	    description: string;
	    slug: string;
	    version: string;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.display_name = source["display_name"];
	        this.registry_name = source["registry_name"];
	        this.summary = source["summary"];
	        this.description = source["description"];
	        this.slug = source["slug"];
	        this.version = source["version"];
	        this.score = source["score"];
	    }
	}
	export class SkillFullInfo {
	    name: string;
	    description?: string;
	    version?: string;
	    registry?: string;
	    url?: string;
	    markdown?: string;
	    raw?: string;
	    line_count?: number;
	    char_count?: number;
	    tags?: string[];
	
	    static createFrom(source: any = {}) {
	        return new SkillFullInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.version = source["version"];
	        this.registry = source["registry"];
	        this.url = source["url"];
	        this.markdown = source["markdown"];
	        this.raw = source["raw"];
	        this.line_count = source["line_count"];
	        this.char_count = source["char_count"];
	        this.tags = source["tags"];
	    }
	}
	
	
	
	
	
	export class ToolUIInfo {
	    name: string;
	    description: string;
	    category: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolUIInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.enabled = source["enabled"];
	    }
	}
	
	

}

export namespace config {
	
	export class SecureString {
	
	
	    static createFrom(source: any = {}) {
	        return new SecureString(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class ModelConfig {
	    model_name: string;
	    provider: string;
	    model: string;
	    api_base?: string;
	    proxy?: string;
	    fallbacks?: string[];
	    auth_method?: string;
	    connect_mode?: string;
	    workspace?: string;
	    rpm?: number;
	    max_tokens_field?: string;
	    request_timeout?: number;
	    thinking_level?: string;
	    extra_body?: Record<string, any>;
	    custom_headers?: Record<string, string>;
	    api_keys: SecureString[];
	    enabled?: boolean;
	    user_agent?: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model_name = source["model_name"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.api_base = source["api_base"];
	        this.proxy = source["proxy"];
	        this.fallbacks = source["fallbacks"];
	        this.auth_method = source["auth_method"];
	        this.connect_mode = source["connect_mode"];
	        this.workspace = source["workspace"];
	        this.rpm = source["rpm"];
	        this.max_tokens_field = source["max_tokens_field"];
	        this.request_timeout = source["request_timeout"];
	        this.thinking_level = source["thinking_level"];
	        this.extra_body = source["extra_body"];
	        this.custom_headers = source["custom_headers"];
	        this.api_keys = this.convertValues(source["api_keys"], SecureString);
	        this.enabled = source["enabled"];
	        this.user_agent = source["user_agent"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class SubCommandInfo {
	    name: string;
	    description: string;
	    args_usage: string;
	
	    static createFrom(source: any = {}) {
	        return new SubCommandInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.args_usage = source["args_usage"];
	    }
	}
	export class CommandInfo {
	    name: string;
	    description: string;
	    usage: string;
	    aliases: string[];
	    sub_commands: SubCommandInfo[];
	
	    static createFrom(source: any = {}) {
	        return new CommandInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.usage = source["usage"];
	        this.aliases = source["aliases"];
	        this.sub_commands = this.convertValues(source["sub_commands"], SubCommandInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class architecture {
	    id: string;
	    name: string;
	    description: string;
	    best_for: string[];
	    aliases: string[];
	
	    static createFrom(source: any = {}) {
	        return new architecture(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.best_for = source["best_for"];
	        this.aliases = source["aliases"];
	    }
	}

}

export namespace registry {
	
	export class TestConfig {
	    command: string;
	    fail_prompt: string;
	
	    static createFrom(source: any = {}) {
	        return new TestConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.command = source["command"];
	        this.fail_prompt = source["fail_prompt"];
	    }
	}
	export class ExpertPlugin {
	    id: string;
	    name: string;
	    description: string;
	    endpoint: string;
	    triggers: string[];
	    language: string;
	    start_command: string;
	    dependency_endpoint: string;
	    test_config?: TestConfig;
	
	    static createFrom(source: any = {}) {
	        return new ExpertPlugin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.endpoint = source["endpoint"];
	        this.triggers = source["triggers"];
	        this.language = source["language"];
	        this.start_command = source["start_command"];
	        this.dependency_endpoint = source["dependency_endpoint"];
	        this.test_config = this.convertValues(source["test_config"], TestConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

