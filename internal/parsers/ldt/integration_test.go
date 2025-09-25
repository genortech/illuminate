package ldt

import (
	"os"
	"testing"
)

func TestLDTFormatDetection(t *testing.T) {
	// Test with the actual sample LDT file
	data, err := os.ReadFile("../../../tests/samples/102-0136.ldt")
	if err != nil {
		t.Skipf("Sample LDT file not found: %v", err)
	}

	parser := NewParser()

	// Test format detection
	confidence, version := parser.DetectFormat(data)
	if confidence < 0.5 {
		t.Errorf("Low confidence for LDT format detection: %f", confidence)
	}
	if version != string(Version10) {
		t.Errorf("Expected version %s, got %s", Version10, version)
	}

	t.Logf("LDT format detection successful with confidence: %f", confidence)
}
