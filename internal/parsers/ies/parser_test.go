package ies

import (
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}

	versions := parser.GetSupportedVersions()
	expectedVersions := []string{string(VersionLM631995), string(VersionLM632002)}

	if len(versions) != len(expectedVersions) {
		t.Errorf("Expected %d supported versions, got %d", len(expectedVersions), len(versions))
	}

	for i, expected := range expectedVersions {
		if i >= len(versions) || versions[i] != expected {
			t.Errorf("Expected version %s at index %d, got %s", expected, i, versions[i])
		}
	}
}

func TestDetectFormat(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name               string
		data               string
		expectedConfidence float64
		expectedVersion    string
	}{
		{
			name:               "LM-63-2002 format",
			data:               "IESNA:LM-63-2002\n[TEST] 123\nTILT=NONE\n",
			expectedConfidence: 0.95,
			expectedVersion:    string(VersionLM632002),
		},
		{
			name:               "LM-63-1995 format",
			data:               "IESNA:LM-63-1995\n[TEST] 123\nTILT=NONE\n",
			expectedConfidence: 0.95,
			expectedVersion:    string(VersionLM631995),
		},
		{
			name:               "IESNA91 format",
			data:               "IESNA91\n[TEST] 123\nTILT=NONE\n",
			expectedConfidence: 0.90,
			expectedVersion:    string(VersionLM631995),
		},
		{
			name:               "TILT line present",
			data:               "[TEST] 123\n[MANUFAC] Test\nTILT=NONE\n1 2 3 4 5\n",
			expectedConfidence: 0.3, // 0.3 for TILT only (numeric data doesn't reach threshold)
			expectedVersion:    "",
		},
		{
			name:               "Not IES format",
			data:               "This is not an IES file\nJust some text\n",
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
		{
			name:               "Empty data",
			data:               "",
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence, version := parser.DetectFormat([]byte(tt.data))

			if confidence < tt.expectedConfidence-0.1 || confidence > tt.expectedConfidence+0.1 {
				t.Errorf("Expected confidence around %f, got %f", tt.expectedConfidence, confidence)
			}

			if version != tt.expectedVersion {
				t.Errorf("Expected version %s, got %s", tt.expectedVersion, version)
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	parser := NewParser()

	testData := []string{
		"IESNA:LM-63-2002",
		"[TEST] 12345",
		"[MANUFAC] Test Manufacturer",
		"[LUMCAT] TEST-001",
		"[LUMINAIRE] Test Luminaire Description",
		"[MORE] Additional description",
		"[TESTLAB] Test Laboratory",
		"TILT=NONE",
	}

	iesFile := &IESFile{
		Header: IESHeader{
			Keywords: make(map[string]string),
		},
	}

	lineIndex, err := parser.parseHeader(testData, iesFile)
	if err != nil {
		t.Fatalf("parseHeader failed: %v", err)
	}

	if lineIndex != len(testData) {
		t.Errorf("Expected lineIndex %d, got %d", len(testData), lineIndex)
	}

	if iesFile.Header.Version != VersionLM632002 {
		t.Errorf("Expected version %s, got %s", VersionLM632002, iesFile.Header.Version)
	}

	if iesFile.Header.TiltData != "TILT=NONE" {
		t.Errorf("Expected TILT=NONE, got %s", iesFile.Header.TiltData)
	}

	expectedKeywords := map[string]string{
		"TEST":      "12345",
		"MANUFAC":   "Test Manufacturer",
		"LUMCAT":    "TEST-001",
		"LUMINAIRE": "Test Luminaire Description Additional description",
		"TESTLAB":   "Test Laboratory",
	}

	for key, expectedValue := range expectedKeywords {
		if value, exists := iesFile.Header.Keywords[key]; !exists {
			t.Errorf("Expected keyword %s not found", key)
		} else if value != expectedValue {
			t.Errorf("Expected keyword %s value %s, got %s", key, expectedValue, value)
		}
	}
}

func TestParsePhotometricData(t *testing.T) {
	parser := NewParser()

	testData := []string{
		"1 1000.0 1.0 3 2 1 1 1.0 2.0 3.0",
		"1.0 1.0 100.0",
		"0.0 45.0 90.0",
		"0.0 180.0",
		"100.0 200.0",
		"150.0 250.0",
		"200.0 300.0",
	}

	iesFile := &IESFile{}

	err := parser.parsePhotometricData(testData, iesFile)
	if err != nil {
		t.Fatalf("parsePhotometricData failed: %v", err)
	}

	// Verify parsed values
	if iesFile.Photometric.NumberOfLamps != 1 {
		t.Errorf("Expected NumberOfLamps 1, got %d", iesFile.Photometric.NumberOfLamps)
	}

	if iesFile.Photometric.LumensPerLamp != 1000.0 {
		t.Errorf("Expected LumensPerLamp 1000.0, got %f", iesFile.Photometric.LumensPerLamp)
	}

	if iesFile.Photometric.CandelaMultiplier != 1.0 {
		t.Errorf("Expected CandelaMultiplier 1.0, got %f", iesFile.Photometric.CandelaMultiplier)
	}

	if iesFile.Photometric.NumVerticalAngles != 3 {
		t.Errorf("Expected NumVerticalAngles 3, got %d", iesFile.Photometric.NumVerticalAngles)
	}

	if iesFile.Photometric.NumHorizontalAngles != 2 {
		t.Errorf("Expected NumHorizontalAngles 2, got %d", iesFile.Photometric.NumHorizontalAngles)
	}

	expectedVerticalAngles := []float64{0.0, 45.0, 90.0}
	if len(iesFile.Photometric.VerticalAngles) != len(expectedVerticalAngles) {
		t.Errorf("Expected %d vertical angles, got %d", len(expectedVerticalAngles), len(iesFile.Photometric.VerticalAngles))
	}

	for i, expected := range expectedVerticalAngles {
		if i < len(iesFile.Photometric.VerticalAngles) && iesFile.Photometric.VerticalAngles[i] != expected {
			t.Errorf("Expected vertical angle[%d] %f, got %f", i, expected, iesFile.Photometric.VerticalAngles[i])
		}
	}

	expectedHorizontalAngles := []float64{0.0, 180.0}
	if len(iesFile.Photometric.HorizontalAngles) != len(expectedHorizontalAngles) {
		t.Errorf("Expected %d horizontal angles, got %d", len(expectedHorizontalAngles), len(iesFile.Photometric.HorizontalAngles))
	}

	for i, expected := range expectedHorizontalAngles {
		if i < len(iesFile.Photometric.HorizontalAngles) && iesFile.Photometric.HorizontalAngles[i] != expected {
			t.Errorf("Expected horizontal angle[%d] %f, got %f", i, expected, iesFile.Photometric.HorizontalAngles[i])
		}
	}

	// Verify candela values matrix
	expectedCandelaValues := [][]float64{
		{100.0, 200.0},
		{150.0, 250.0},
		{200.0, 300.0},
	}

	if len(iesFile.Photometric.CandelaValues) != len(expectedCandelaValues) {
		t.Errorf("Expected %d candela value rows, got %d", len(expectedCandelaValues), len(iesFile.Photometric.CandelaValues))
	}

	for i, expectedRow := range expectedCandelaValues {
		if i >= len(iesFile.Photometric.CandelaValues) {
			continue
		}

		actualRow := iesFile.Photometric.CandelaValues[i]
		if len(actualRow) != len(expectedRow) {
			t.Errorf("Expected %d candela values in row %d, got %d", len(expectedRow), i, len(actualRow))
			continue
		}

		for j, expected := range expectedRow {
			if j < len(actualRow) && actualRow[j] != expected {
				t.Errorf("Expected candela value[%d][%d] %f, got %f", i, j, expected, actualRow[j])
			}
		}
	}
}

func TestParseFloatArray(t *testing.T) {
	parser := NewParser()

	testData := []string{
		"1.0 2.0 3.0",
		"4.0 5.0",
		"6.0",
	}

	var result []float64
	lineIndex := 0
	expectedCount := 6

	err := parser.parseFloatArray(testData, &lineIndex, &result, expectedCount)
	if err != nil {
		t.Fatalf("parseFloatArray failed: %v", err)
	}

	expected := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	if len(result) != len(expected) {
		t.Errorf("Expected %d values, got %d", len(expected), len(result))
	}

	for i, expectedValue := range expected {
		if i < len(result) && result[i] != expectedValue {
			t.Errorf("Expected value[%d] %f, got %f", i, expectedValue, result[i])
		}
	}

	if lineIndex != len(testData) {
		t.Errorf("Expected lineIndex %d, got %d", len(testData), lineIndex)
	}
}

func TestParseFloatArrayInsufficientData(t *testing.T) {
	parser := NewParser()

	testData := []string{
		"1.0 2.0",
	}

	var result []float64
	lineIndex := 0
	expectedCount := 5

	err := parser.parseFloatArray(testData, &lineIndex, &result, expectedCount)
	if err == nil {
		t.Fatal("Expected error for insufficient data, got nil")
	}

	if !strings.Contains(err.Error(), "expected 5 values, got 2") {
		t.Errorf("Expected error about insufficient values, got: %v", err)
	}
}

func TestIsNumericLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "All numeric",
			line:     "1.0 2.0 3.0 4.0",
			expected: true,
		},
		{
			name:     "Mostly numeric",
			line:     "1.0 2.0 text 4.0",
			expected: true, // 3/4 = 0.75 > 0.5, so this is considered numeric
		},
		{
			name:     "No numeric",
			line:     "text more text",
			expected: false,
		},
		{
			name:     "Empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "Single number",
			line:     "42.5",
			expected: true,
		},
		{
			name:     "Integers and floats",
			line:     "1 2.5 3 4.0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNumericLine(tt.line)
			if result != tt.expected {
				t.Errorf("Expected %v for line '%s', got %v", tt.expected, tt.line, result)
			}
		})
	}
}

func TestParseCompleteIESFile(t *testing.T) {
	parser := NewParser()

	// Create a minimal but complete IES file
	iesData := `IESNA:LM-63-2002
[TEST] 12345
[MANUFAC] Test Manufacturer
[LUMCAT] TEST-001
[LUMINAIRE] Test Luminaire
[TESTLAB] Test Lab
TILT=NONE
1 1000.0 1.0 2 2 1 1 1.0 2.0 3.0
1.0 1.0 100.0
0.0 90.0
0.0 180.0
100.0 200.0
150.0 250.0`

	commonData, err := parser.Parse([]byte(iesData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if commonData == nil {
		t.Fatal("Parse returned nil data")
	}

	// Verify conversion to common model
	if commonData.Metadata.Manufacturer != "Test Manufacturer" {
		t.Errorf("Expected manufacturer 'Test Manufacturer', got '%s'", commonData.Metadata.Manufacturer)
	}

	if commonData.Metadata.CatalogNumber != "TEST-001" {
		t.Errorf("Expected catalog number 'TEST-001', got '%s'", commonData.Metadata.CatalogNumber)
	}

	if commonData.Photometry.PhotometryType != "C" {
		t.Errorf("Expected photometry type 'C', got '%s'", commonData.Photometry.PhotometryType)
	}

	if commonData.Photometry.LuminousFlux != 1000.0 {
		t.Errorf("Expected luminous flux 1000.0, got %f", commonData.Photometry.LuminousFlux)
	}

	if len(commonData.Photometry.VerticalAngles) != 2 {
		t.Errorf("Expected 2 vertical angles, got %d", len(commonData.Photometry.VerticalAngles))
	}

	if len(commonData.Photometry.HorizontalAngles) != 2 {
		t.Errorf("Expected 2 horizontal angles, got %d", len(commonData.Photometry.HorizontalAngles))
	}

	if len(commonData.Photometry.CandelaValues) != 2 {
		t.Errorf("Expected 2 candela value rows, got %d", len(commonData.Photometry.CandelaValues))
	}

	for i, row := range commonData.Photometry.CandelaValues {
		if len(row) != 2 {
			t.Errorf("Expected 2 candela values in row %d, got %d", i, len(row))
		}
	}
}

func TestParseInvalidData(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name string
		data string
	}{
		{
			name: "Empty data",
			data: "",
		},
		{
			name: "Invalid photometric line",
			data: `IESNA:LM-63-2002
TILT=NONE
invalid data line`,
		},
		{
			name: "Insufficient photometric fields",
			data: `IESNA:LM-63-2002
TILT=NONE
1 2 3`,
		},
		{
			name: "Missing angle data",
			data: `IESNA:LM-63-2002
TILT=NONE
1 1000.0 1.0 2 2 1 1 1.0 2.0 3.0
1.0 1.0 100.0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse([]byte(tt.data))
			if err == nil {
				t.Errorf("Expected error for invalid data, got nil")
			}
		})
	}
}
