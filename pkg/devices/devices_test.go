package devices

import (
	"testing"

	"ada-love-ai/pkg/state"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		want   bool
	}{
		{
			name: "disabled service",
			cfg: Config{
				Enabled:    false,
				MonitorUSB: false,
			},
			want: false,
		},
		{
			name: "enabled but no sources",
			cfg: Config{
				Enabled:    true,
				MonitorUSB: false,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateMgr := state.NewManager("")
			svc := NewService(tt.cfg, stateMgr)
			if svc == nil {
				t.Fatal("NewService() should not return nil")
			}
			if svc.enabled != tt.want && tt.cfg.Enabled {
				t.Errorf("enabled = %v, want %v", svc.enabled, tt.want)
			}
		})
	}
}

func TestServiceSetBus(t *testing.T) {
	stateMgr := state.NewManager("")
	svc := NewService(Config{Enabled: false}, stateMgr)

	svc.SetBus(nil)
	svc.mu.RLock()
	if svc.bus != nil {
		t.Error("bus should be nil after SetBus(nil)")
	}
	svc.mu.RUnlock()
}