package events

import (
	"testing"
)

func TestDeviceEventFormatMessage(t *testing.T) {
	tests := []struct {
		name     string
		event    *DeviceEvent
		wantPart string
	}{
		{
			name: "add event",
			event: &DeviceEvent{
				Action:       ActionAdd,
				Kind:         KindUSB,
				Vendor:       "VendorX",
				Product:      "ProductY",
				Capabilities: "USB 3.0",
			},
			wantPart: "Connected",
		},
		{
			name: "remove event",
			event: &DeviceEvent{
				Action:  ActionRemove,
				Kind:    KindUSB,
				Vendor:  "VendorX",
				Product: "ProductY",
			},
			wantPart: "Disconnected",
		},
		{
			name: "event with serial",
			event: &DeviceEvent{
				Action:  ActionAdd,
				Kind:    KindBluetooth,
				Vendor:  "VendorX",
				Product: "ProductY",
				Serial:  "SN12345",
			},
			wantPart: "Serial: SN12345",
		},
		{
			name: "event without capabilities",
			event: &DeviceEvent{
				Action:  ActionAdd,
				Kind:    KindPCI,
				Vendor:  "VendorX",
				Product: "ProductY",
			},
			wantPart: "Type: pci",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.FormatMessage()
			if got == "" {
				t.Fatal("FormatMessage() should not return empty string")
			}
			if len(got) < 10 {
				t.Errorf("FormatMessage() = %q, too short", got)
			}
		})
	}
}

func TestActionConstants(t *testing.T) {
	if ActionAdd != "add" {
		t.Errorf("ActionAdd = %q, want add", ActionAdd)
	}
	if ActionRemove != "remove" {
		t.Errorf("ActionRemove = %q, want remove", ActionRemove)
	}
	if ActionChange != "change" {
		t.Errorf("ActionChange = %q, want change", ActionChange)
	}
}

func TestKindConstants(t *testing.T) {
	if KindUSB != "usb" {
		t.Errorf("KindUSB = %q, want usb", KindUSB)
	}
	if KindBluetooth != "bluetooth" {
		t.Errorf("KindBluetooth = %q, want bluetooth", KindBluetooth)
	}
	if KindPCI != "pci" {
		t.Errorf("KindPCI = %q, want pci", KindPCI)
	}
	if KindGeneric != "generic" {
		t.Errorf("KindGeneric = %q, want generic", KindGeneric)
	}
}
