package tools

// ToolProfile defines a collection of tools that can be registered together.
type ToolProfile struct {
	Name        string
	Description string
	Tools       []Tool
}

// ToolProfiles contains predefined tool collections.
var ToolProfiles = map[string]ToolProfile{
	"development": {
		Name:        "Development",
		Description: "Tools for software development workflow",
		Tools:       []Tool{
			// File system tools
			// These will be populated by the caller with actual tool instances
		},
	},
	"web_search": {
		Name:        "Web Search",
		Description: "Tools for web search and information retrieval",
		Tools:       []Tool{
			// Web search tools
		},
	},
	"full": {
		Name:        "Full",
		Description: "All available tools",
		Tools:       []Tool{},
	},
}

// GetProfile returns a tool profile by name.
// Returns an empty profile if the profile doesn't exist.
func GetProfile(name string) ToolProfile {
	if profile, ok := ToolProfiles[name]; ok {
		return profile
	}
	return ToolProfile{Name: name, Description: "", Tools: nil}
}

// ListProfiles returns all available profile names.
func ListProfiles() []string {
	names := make([]string, 0, len(ToolProfiles))
	for name := range ToolProfiles {
		names = append(names, name)
	}
	return names
}

// RegisterProfile registers all tools from a profile.
// Tools that are already registered are skipped.
func (r *ToolRegistry) RegisterProfile(profileName string) {
	profile := GetProfile(profileName)
	if len(profile.Tools) == 0 {
		return
	}
	r.RegisterTools(profile.Tools, profileName)
}
