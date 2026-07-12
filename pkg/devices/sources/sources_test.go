package sources

import (
	"testing"

	"ada-love-ai/pkg/devices/events"
)

func TestUSBMonitorKind(t *testing.T) {
	monitor := NewUSBMonitor()
	if monitor.Kind() != events.KindUSB {
		t.Errorf("Kind() = %v, want usb", monitor.Kind())
	}
}

func TestNewUSBMonitor(t *testing.T) {
	monitor := NewUSBMonitor()
	if monitor == nil {
		t.Error("NewUSBMonitor() should not return nil")
	}
}

func TestUSBDeviceClassLookup(t *testing.T) {
	tests := []struct {
		classCode string
		want      string
	}{
		{"01", "Audio"},
		{"03", "HID (Keyboard/Mouse/Gamepad)"},
		{"08", "Mass Storage (USB Flash Drive/Hard Disk)"},
		{"0e", "Video (Camera)"},
		{"e0", "Wireless Controller (Bluetooth)"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.classCode, func(t *testing.T) {
			got, ok := usbClassToCapability[tt.classCode]
			if tt.want == "" {
				if ok {
					t.Errorf("usbClassToCapability[%q] = %q, want not found", tt.classCode, got)
				}
			} else {
				if !ok {
					t.Errorf("usbClassToCapability[%q] not found, want %q", tt.classCode, tt.want)
				} else if got != tt.want {
					t.Errorf("usbClassToCapability[%q] = %q, want %q", tt.classCode, got, tt.want)
				}
			}
		})
	}
}
