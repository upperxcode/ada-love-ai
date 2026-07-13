package hardwaretools

import (
	"testing"
)

func TestI2CToolName(t *testing.T) {
	tool := NewI2CTool()
	if tool.Name() != "i2c" {
		t.Errorf("Name() = %q, want i2c", tool.Name())
	}
}

func TestI2CToolDescription(t *testing.T) {
	tool := NewI2CTool()
	desc := tool.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestI2CToolParameters(t *testing.T) {
	tool := NewI2CTool()
	params := tool.Parameters()
	if params == nil {
		t.Error("Parameters() should not return nil")
	}
}

func TestNewI2CTool(t *testing.T) {
	tool := NewI2CTool()
	if tool == nil {
		t.Error("NewI2CTool() should not return nil")
	}
}
