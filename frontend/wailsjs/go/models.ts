export namespace backend {
	
	export class ToolProfile {
	    id: number;
	    name: string;
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
	export class ProviderConfig {
	    api_url: string;
	    api_key: string;
	    type_connection: string;
	    models: Record<string, ModelSettings>;
	
	    static createFrom(source: any = {}) {
	        return new ProviderConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_url = source["api_url"];
	        this.api_key = source["api_key"];
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
	export class AgentConfig {
	    name: string;
	    description: string;
	    provider: string;
	    model: string;
	    type: string;
	    icon: string;
	    color: string;
	    max_iterations?: number;
	    temperature?: number;
	    delegates?: string[];
	    system_prompt?: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.type = source["type"];
	        this.icon = source["icon"];
	        this.color = source["color"];
	        this.max_iterations = source["max_iterations"];
	        this.temperature = source["temperature"];
	        this.delegates = source["delegates"];
	        this.system_prompt = source["system_prompt"];
	    }
	}
	export class WorkerConfig {
	    name: string;
	    persona: string;
	    language: string;
	    connection_type: string;
	    connection_name: string;
	    connection_config: string;
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
	        this.name = source["name"];
	        this.persona = source["persona"];
	        this.language = source["language"];
	        this.connection_type = source["connection_type"];
	        this.connection_name = source["connection_name"];
	        this.connection_config = source["connection_config"];
	        this.inherit_folders = source["inherit_folders"];
	        this.inherit_knowledge = source["inherit_knowledge"];
	        this.inherit_skills = source["inherit_skills"];
	        this.inherit_tools = source["inherit_tools"];
	        this.inherit_persona = source["inherit_persona"];
	    }
	}
	export class WorkspaceConfig {
	    title: string;
	    description: string;
	    path: string;
	    folders: string[];
	    personality: string;
	    knowledge: string[];
	    workspace_agents: WorkerConfig[];
	    skills: string[];
	    tools: string[];
	    enabled: boolean;
	    color: string;
	    icon: string;
	    max_prompt_send: number;
	    commit_changes: boolean;
	    max_context_length: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.path = source["path"];
	        this.folders = source["folders"];
	        this.personality = source["personality"];
	        this.knowledge = source["knowledge"];
	        this.workspace_agents = this.convertValues(source["workspace_agents"], WorkerConfig);
	        this.skills = source["skills"];
	        this.tools = source["tools"];
	        this.enabled = source["enabled"];
	        this.color = source["color"];
	        this.icon = source["icon"];
	        this.max_prompt_send = source["max_prompt_send"];
	        this.commit_changes = source["commit_changes"];
	        this.max_context_length = source["max_context_length"];
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
	export class AdaConfig {
	    active_workspace_path: string;
	    active_workspace_index: number;
	    workspaces: WorkspaceConfig[];
	    // Go type: struct { ModelName string "json:\"model_name\""; Provider string "json:\"provider\""; EmbeddingModel string "json:\"embedding_model\""; EmbeddingProvider string "json:\"embedding_provider\"" }
	    tiny_brain: any;
	    embedding_model: string;
	    embedding_provider: string;
	    image_model: string;
	    image_provider: string;
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
	
	    static createFrom(source: any = {}) {
	        return new AdaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.active_workspace_path = source["active_workspace_path"];
	        this.active_workspace_index = source["active_workspace_index"];
	        this.workspaces = this.convertValues(source["workspaces"], WorkspaceConfig);
	        this.tiny_brain = this.convertValues(source["tiny_brain"], Object);
	        this.embedding_model = source["embedding_model"];
	        this.embedding_provider = source["embedding_provider"];
	        this.image_model = source["image_model"];
	        this.image_provider = source["image_provider"];
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
	    title: string;
	    summary: string;
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
	        this.title = source["title"];
	        this.summary = source["summary"];
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

