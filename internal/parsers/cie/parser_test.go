package cie

import (
	"testing"
)

func TestParser_DetectFormat(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name               string
		input              string
		expectedConfidence float64
		expectedVersion    string
	}{
		{
			name: "Valid CIE i-table format",
			input: `   1   0   0        OSL0526 PLED II 17W AE 3000K 2172.2 lms
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`,
			expectedConfidence: 0.3,
			expectedVersion:    string(VersionITable),
		},
		{
			name: "Valid CIE with different description",
			input: `   1   1   0        StreetLED3 17W 4K Aero P2DG220923057-10 - 2458 lm
 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192
 192 192 192 192 192 192 192 192 192 192 191 190 191 192 192 193 194`,
			expectedConfidence: 0.3,
			expectedVersion:    string(VersionITable),
		},
		{
			name: "Invalid format - not enough fields in header",
			input: `1 0
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`,
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
		{
			name: "Invalid format - non-numeric header",
			input: `abc def ghi jkl description
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`,
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
		{
			name: "Invalid format - wrong number of data values per row",
			input: `   1   0   0        Test luminaire
 135 135 135 135 135
 135 135 135 135 135`,
			expectedConfidence: 0.0, // No confidence due to wrong data format
			expectedVersion:    "",
		},
		{
			name:               "Empty input",
			input:              "",
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
		{
			name:               "Single line only",
			input:              `   1   0   0        Test luminaire`,
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence, version := parser.DetectFormat([]byte(tt.input))

			if confidence < tt.expectedConfidence {
				t.Errorf("DetectFormat() confidence = %v, want >= %v", confidence, tt.expectedConfidence)
			}

			if tt.expectedVersion != "" && version != tt.expectedVersion {
				t.Errorf("DetectFormat() version = %v, want %v (confidence was %v)", version, tt.expectedVersion, confidence)
			}
		})
	}
}

func TestParser_parseHeader(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name        string
		headerLine  string
		expectError bool
		expected    CIEHeader
	}{
		{
			name:       "Valid header with description",
			headerLine: "   1   0   0        OSL0526 PLED II 17W AE 3000K 2172.2 lms",
			expected: CIEHeader{
				FormatType:   1,
				SymmetryType: 0,
				Reserved:     0,
				Description:  "OSL0526 PLED II 17W AE 3000K 2172.2 lms",
			},
		},
		{
			name:       "Valid header with different values",
			headerLine: "   1   1   0        StreetLED3 17W 4K Aero P2DG220923057-10 - 2458 lm",
			expected: CIEHeader{
				FormatType:   1,
				SymmetryType: 1,
				Reserved:     0,
				Description:  "StreetLED3 17W 4K Aero P2DG220923057-10 - 2458 lm",
			},
		},
		{
			name:       "Minimal valid header",
			headerLine: "1 0 0 Test",
			expected: CIEHeader{
				FormatType:   1,
				SymmetryType: 0,
				Reserved:     0,
				Description:  "Test",
			},
		},
		{
			name:        "Invalid header - too few fields",
			headerLine:  "1 0 0",
			expectError: true,
		},
		{
			name:        "Invalid header - non-numeric format type",
			headerLine:  "abc 0 0 Test",
			expectError: true,
		},
		{
			name:        "Invalid header - non-numeric symmetry type",
			headerLine:  "1 abc 0 Test",
			expectError: true,
		},
		{
			name:        "Invalid header - non-numeric reserved field",
			headerLine:  "1 0 abc Test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cieFile := &CIEFile{}
			err := parser.parseHeader(tt.headerLine, cieFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseHeader() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseHeader() unexpected error: %v", err)
				return
			}

			if cieFile.Header.FormatType != tt.expected.FormatType {
				t.Errorf("FormatType = %v, want %v", cieFile.Header.FormatType, tt.expected.FormatType)
			}
			if cieFile.Header.SymmetryType != tt.expected.SymmetryType {
				t.Errorf("SymmetryType = %v, want %v", cieFile.Header.SymmetryType, tt.expected.SymmetryType)
			}
			if cieFile.Header.Reserved != tt.expected.Reserved {
				t.Errorf("Reserved = %v, want %v", cieFile.Header.Reserved, tt.expected.Reserved)
			}
			if cieFile.Header.Description != tt.expected.Description {
				t.Errorf("Description = %v, want %v", cieFile.Header.Description, tt.expected.Description)
			}
		})
	}
}

func TestParser_reshapeIntensityData(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    [][]float64
		expected [][]float64
	}{
		{
			name: "Already correct shape",
			input: func() [][]float64 {
				// Create 19x16 matrix
				data := make([][]float64, 19)
				for i := range data {
					data[i] = make([]float64, 16)
					for j := range data[i] {
						data[i][j] = float64(i*16 + j)
					}
				}
				return data
			}(),
			expected: func() [][]float64 {
				// Same 19x16 matrix
				data := make([][]float64, 19)
				for i := range data {
					data[i] = make([]float64, 16)
					for j := range data[i] {
						data[i][j] = float64(i*16 + j)
					}
				}
				return data
			}(),
		},
		{
			name: "Single row with all data",
			input: [][]float64{
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			},
			expected: func() [][]float64 {
				data := make([][]float64, 19)
				value := 1.0
				for i := range data {
					data[i] = make([]float64, 16)
					for j := range data[i] {
						if value <= 20 {
							data[i][j] = value
							value++
						} else {
							data[i][j] = 0
						}
					}
				}
				return data
			}(),
		},
		{
			name:  "Empty input",
			input: [][]float64{},
			expected: func() [][]float64 {
				data := make([][]float64, 19)
				for i := range data {
					data[i] = make([]float64, 16)
					// All zeros
				}
				return data
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.reshapeIntensityData(tt.input)

			// Check dimensions
			if len(result) != 19 {
				t.Errorf("Result should have 19 rows, got %d", len(result))
			}

			for i, row := range result {
				if len(row) != 16 {
					t.Errorf("Row %d should have 16 columns, got %d", i, len(row))
				}
			}

			// For the "already correct shape" test, check values
			if tt.name == "Already correct shape" {
				for i := range result {
					for j := range result[i] {
						expected := float64(i*16 + j)
						if result[i][j] != expected {
							t.Errorf("Value at [%d][%d] = %v, want %v", i, j, result[i][j], expected)
						}
					}
				}
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	// Test with a minimal valid CIE file
	validCIEData := `   1   0   0        Test LED 17W 1000 lm
 100 100 100 100 100 100 100 100 100 100 100 100 100 100 100 100 100
  90  90  90  90  90  90  90  90  90  90  90  90  90  90  90  90  90
  80  80  80  80  80  80  80  80  80  80  80  80  80  80  80  80  80
  70  70  70  70  70  70  70  70  70  70  70  70  70  70  70  70  70
  60  60  60  60  60  60  60  60  60  60  60  60  60  60  60  60  60
  50  50  50  50  50  50  50  50  50  50  50  50  50  50  50  50  50
  40  40  40  40  40  40  40  40  40  40  40  40  40  40  40  40  40
  30  30  30  30  30  30  30  30  30  30  30  30  30  30  30  30  30
  20  20  20  20  20  20  20  20  20  20  20  20  20  20  20  20  20
  10  10  10  10  10  10  10  10  10  10  10  10  10  10  10  10  10
   5   5   5   5   5   5   5   5   5   5   5   5   5   5   5   5   5
   3   3   3   3   3   3   3   3   3   3   3   3   3   3   3   3   3
   2   2   2   2   2   2   2   2   2   2   2   2   2   2   2   2   2
   1   1   1   1   1   1   1   1   1   1   1   1   1   1   1   1   1
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0
   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0   0`

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Valid CIE file",
			input:       validCIEData,
			expectError: false,
		},
		{
			name:        "Invalid header",
			input:       "invalid header\n100 100 100",
			expectError: true,
		},
		{
			name:        "Empty input",
			input:       "",
			expectError: true,
		},
		{
			name:        "Header only",
			input:       "1 0 0 Test",
			expectError: true, // No photometric data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse([]byte(tt.input))

			if tt.expectError {
				if err == nil {
					t.Errorf("Parse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Parse() returned nil result")
				return
			}

			// Basic validation of the result
			if result.Photometry.PhotometryType != "C" {
				t.Errorf("Expected photometry type C, got %s", result.Photometry.PhotometryType)
			}

			if len(result.Photometry.VerticalAngles) != 19 {
				t.Errorf("Expected 19 vertical angles, got %d", len(result.Photometry.VerticalAngles))
			}

			if len(result.Photometry.HorizontalAngles) != 16 {
				t.Errorf("Expected 16 horizontal angles, got %d", len(result.Photometry.HorizontalAngles))
			}

			if len(result.Photometry.CandelaValues) != 19 {
				t.Errorf("Expected 19 rows of candela values, got %d", len(result.Photometry.CandelaValues))
			}

			for i, row := range result.Photometry.CandelaValues {
				if len(row) != 16 {
					t.Errorf("Row %d should have 16 candela values, got %d", i, len(row))
				}
			}
		})
	}
}

func TestExtractLuminousFlux(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    float64
	}{
		{
			name:        "Standard format with lms",
			description: "OSL0526 PLED II 17W AE 3000K 2172.2 lms",
			expected:    2172.2,
		},
		{
			name:        "Standard format with lm",
			description: "StreetLED3 17W 4K Aero P2DG220923057-10 - 2458 lm",
			expected:    2458,
		},
		{
			name:        "Simple format",
			description: "Test LED 1000 lm",
			expected:    1000,
		},
		{
			name:        "No luminous flux",
			description: "Test LED 17W",
			expected:    0,
		},
		{
			name:        "Empty description",
			description: "",
			expected:    0,
		},
		{
			name:        "Flux at end",
			description: "LED Luminaire 1500lm",
			expected:    1500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLuminousFlux(tt.description)
			if result != tt.expected {
				t.Errorf("extractLuminousFlux(%q) = %v, want %v", tt.description, result, tt.expected)
			}
		})
	}
}

func TestExtractWattage(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    float64
	}{
		{
			name:        "Standard format",
			description: "OSL0526 PLED II 17W AE 3000K 2172.2 lms",
			expected:    17,
		},
		{
			name:        "Different wattage",
			description: "Test LED 25W 1000 lm",
			expected:    25,
		},
		{
			name:        "Decimal wattage",
			description: "Small LED 5.5W",
			expected:    5.5,
		},
		{
			name:        "No wattage",
			description: "Test LED 1000 lm",
			expected:    0,
		},
		{
			name:        "Empty description",
			description: "",
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWattage(tt.description)
			if result != tt.expected {
				t.Errorf("extractWattage(%q) = %v, want %v", tt.description, result, tt.expected)
			}
		})
	}
}

func TestGenerateStandardAngles(t *testing.T) {
	t.Run("Gamma angles", func(t *testing.T) {
		angles := generateStandardGammaAngles()

		if len(angles) != 19 {
			t.Errorf("Expected 19 gamma angles, got %d", len(angles))
		}

		// Check first few angles
		expected := []float64{0, 5, 10, 15, 20}
		for i, exp := range expected {
			if i < len(angles) && angles[i] != exp {
				t.Errorf("Gamma angle %d = %v, want %v", i, angles[i], exp)
			}
		}

		// Check last angle
		if angles[len(angles)-1] != 90 {
			t.Errorf("Last gamma angle = %v, want 90", angles[len(angles)-1])
		}
	})

	t.Run("C-plane angles", func(t *testing.T) {
		angles := generateStandardCPlaneAngles()

		if len(angles) != 16 {
			t.Errorf("Expected 16 C-plane angles, got %d", len(angles))
		}

		// Check first few angles
		expected := []float64{0, 22.5, 45, 67.5, 90}
		for i, exp := range expected {
			if i < len(angles) && angles[i] != exp {
				t.Errorf("C-plane angle %d = %v, want %v", i, angles[i], exp)
			}
		}

		// Check last angle
		if angles[len(angles)-1] != 337.5 {
			t.Errorf("Last C-plane angle = %v, want 337.5", angles[len(angles)-1])
		}
	})
}

func TestGetSupportedVersions(t *testing.T) {
	parser := NewParser()
	versions := parser.GetSupportedVersions()

	expectedVersions := []string{string(Version102), string(VersionITable)}

	if len(versions) != len(expectedVersions) {
		t.Errorf("Expected %d versions, got %d", len(expectedVersions), len(versions))
	}

	for i, expected := range expectedVersions {
		if i < len(versions) && versions[i] != expected {
			t.Errorf("Version %d = %v, want %v", i, versions[i], expected)
		}
	}
}
