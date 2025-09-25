package ldt

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
	if len(versions) != 1 || versions[0] != string(Version10) {
		t.Errorf("Expected supported versions [%s], got %v", Version10, versions)
	}
}

func TestDetectFormat(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name          string
		data          string
		expectedConf  float64
		expectedVer   string
		minConfidence float64
	}{
		{
			name: "Valid LDT file",
			data: `WE-EF;Eulumdat2
2
3
72
5
91
1
6491;24 LED, Wild Light White - 120° angle of beam
AFL120-WL [S61] IP66:LED-8/8W/2200K - 16/32W/3000K;AFL120-WL, Street and Area Lighting
102-0136
102-0136
13 Mar 2023 14:58:32/Quang
605
250
192
180
160
0
0
0
0
100.0
90.6
1.0
0
3
24
LED-8/8W/2200K - 16/32W/3000K
5800.0
2200/3000K
70&80
44.5`,
			minConfidence: 0.5,
			expectedVer:   string(Version10),
		},
		{
			name: "Invalid format - too few lines",
			data: `Line 1
Line 2
Line 3`,
			expectedConf: 0.0,
			expectedVer:  "",
		},
		{
			name: "Invalid format - not LDT",
			data: `IESNA:LM-63-2002
[TEST] Test file
[MANUFAC] Test Manufacturer
TILT=NONE
1 1000 1.0 37 1 1 2 0 0 0
1.0 1.0 100.0`,
			expectedConf: 0.0,
			expectedVer:  "",
		},
		{
			name: "European decimal format",
			data: `Company;Test
2
3
72
5,5
91
1,0
Test measurement
Test luminaire
12345
test.ldt
01 Jan 2023/TestLab
605,5
250,0
192,5`,
			minConfidence: 0.3,
			expectedVer:   string(Version10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence, version := parser.DetectFormat([]byte(tt.data))

			if tt.minConfidence > 0 {
				if confidence < tt.minConfidence {
					t.Errorf("Expected confidence >= %f, got %f", tt.minConfidence, confidence)
				}
				if version != tt.expectedVer {
					t.Errorf("Expected version %s, got %s", tt.expectedVer, version)
				}
			} else {
				if confidence != tt.expectedConf {
					t.Errorf("Expected confidence %f, got %f", tt.expectedConf, confidence)
				}
				if version != tt.expectedVer {
					t.Errorf("Expected version %s, got %s", tt.expectedVer, version)
				}
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	parser := NewParser()

	lines := []string{
		"WE-EF;Eulumdat2",
		"2",
		"3",
		"72",
		"5.0",
		"91",
		"1.0",
		"6491;24 LED, Wild Light White - 120° angle of beam",
		"AFL120-WL [S61] IP66:LED-8/8W/2200K - 16/32W/3000K;AFL120-WL, Street and Area Lighting",
		"102-0136",
		"102-0136",
		"13 Mar 2023 14:58:32/Quang",
		"605",
	}

	ldtFile := &LDTFile{}
	err := parser.parseHeader(lines, ldtFile)
	if err != nil {
		t.Fatalf("parseHeader failed: %v", err)
	}

	// Verify header values
	if ldtFile.Header.CompanyIdentification != "WE-EF;Eulumdat2" {
		t.Errorf("Expected company identification 'WE-EF;Eulumdat2', got '%s'", ldtFile.Header.CompanyIdentification)
	}
	if ldtFile.Header.TypeIndicator != 2 {
		t.Errorf("Expected type indicator 2, got %d", ldtFile.Header.TypeIndicator)
	}
	if ldtFile.Header.SymmetryIndicator != 3 {
		t.Errorf("Expected symmetry indicator 3, got %d", ldtFile.Header.SymmetryIndicator)
	}
	if ldtFile.Header.NumberOfCPlanes != 72 {
		t.Errorf("Expected 72 C planes, got %d", ldtFile.Header.NumberOfCPlanes)
	}
	if ldtFile.Header.DistanceBetweenCPlanes != 5.0 {
		t.Errorf("Expected distance between C planes 5.0, got %f", ldtFile.Header.DistanceBetweenCPlanes)
	}
	if ldtFile.Header.NumberOfLuminousIntensities != 91 {
		t.Errorf("Expected 91 luminous intensities, got %d", ldtFile.Header.NumberOfLuminousIntensities)
	}
}

func TestParseGeometry(t *testing.T) {
	parser := NewParser()

	lines := make([]string, 26)
	// Fill first 12 lines with dummy data
	for i := 0; i < 12; i++ {
		lines[i] = "dummy"
	}

	// Geometry data starts at line 13 (index 12)
	lines[12] = "605"   // Length of luminaire
	lines[13] = "250"   // Width of luminaire
	lines[14] = "192"   // Height of luminaire
	lines[15] = "180"   // Length of luminous area
	lines[16] = "160"   // Width of luminous area
	lines[17] = "0"     // Height of luminous area C0
	lines[18] = "0"     // Height of luminous area C90
	lines[19] = "0"     // Height of luminous area C180
	lines[20] = "0"     // Height of luminous area C270
	lines[21] = "100.0" // Downward flux fraction
	lines[22] = "90.6"  // Light output ratio luminaire
	lines[23] = "1.0"   // Conversion factor
	lines[24] = "0"     // Additional line
	lines[25] = "3"     // Additional line

	ldtFile := &LDTFile{}
	err := parser.parseGeometry(lines, ldtFile)
	if err != nil {
		t.Fatalf("parseGeometry failed: %v", err)
	}

	// Verify geometry values
	if ldtFile.Geometry.LengthOfLuminaire != 605 {
		t.Errorf("Expected length 605, got %f", ldtFile.Geometry.LengthOfLuminaire)
	}
	if ldtFile.Geometry.WidthOfLuminaire != 250 {
		t.Errorf("Expected width 250, got %f", ldtFile.Geometry.WidthOfLuminaire)
	}
	if ldtFile.Geometry.HeightOfLuminaire != 192 {
		t.Errorf("Expected height 192, got %f", ldtFile.Geometry.HeightOfLuminaire)
	}
	if ldtFile.Geometry.ConversionFactor != 1.0 {
		t.Errorf("Expected conversion factor 1.0, got %f", ldtFile.Geometry.ConversionFactor)
	}
}

func TestParseElectrical(t *testing.T) {
	parser := NewParser()

	lines := []string{
		"0", // DR index
		"3", // Number of lamp sets
		// Lamp set 1
		"24",
		"LED-8/8W/2200K - 16/32W/3000K",
		"5800.0",
		"2200/3000K",
		"70&80",
		"44.5",
		// Lamp set 2
		"8",
		"LED-8/8W/2200K",
		"1160.0",
		"2200K",
		"70",
		"9.5",
		// Lamp set 3
		"16",
		"LED-16/32W - 3000K",
		"4640.0",
		"3000K",
		"80",
		"35.0",
	}

	ldtFile := &LDTFile{}
	lineIndex := 0
	err := parser.parseElectrical(lines, &lineIndex, ldtFile)
	if err != nil {
		t.Fatalf("parseElectrical failed: %v", err)
	}

	// Verify electrical values
	if ldtFile.Electrical.DRIndex != 0 {
		t.Errorf("Expected DR index 0, got %d", ldtFile.Electrical.DRIndex)
	}
	if ldtFile.Electrical.NumberOfLampSets != 3 {
		t.Errorf("Expected 3 lamp sets, got %d", ldtFile.Electrical.NumberOfLampSets)
	}
	if len(ldtFile.Electrical.LampSets) != 3 {
		t.Errorf("Expected 3 lamp sets in array, got %d", len(ldtFile.Electrical.LampSets))
	}

	// Verify first lamp set
	lampSet := ldtFile.Electrical.LampSets[0]
	if lampSet.NumberOfLamps != 24 {
		t.Errorf("Expected 24 lamps in first set, got %d", lampSet.NumberOfLamps)
	}
	if lampSet.TotalLuminousFlux != 5800.0 {
		t.Errorf("Expected flux 5800.0, got %f", lampSet.TotalLuminousFlux)
	}
	if lampSet.WattageIncludingBallast != 44.5 {
		t.Errorf("Expected wattage 44.5, got %f", lampSet.WattageIncludingBallast)
	}
}

func TestParsePhotometry(t *testing.T) {
	parser := NewParser()

	// Create test data for 3 C-planes and 4 gamma angles
	lines := []string{
		// C-plane angles
		"0.0",
		"90.0",
		"180.0",
		// Gamma angles
		"0.0",
		"30.0",
		"60.0",
		"90.0",
		// Luminous intensity distribution (3 C-plane x 4 gamma = 12 values)
		// C-plane 0 (0°): values for gamma 0, 30, 60, 90
		"100.0", "95.0", "80.0", "50.0",
		// C-plane 1 (90°): values for gamma 0, 30, 60, 90
		"90.0", "85.0", "70.0", "40.0",
		// C-plane 2 (180°): values for gamma 0, 30, 60, 90
		"80.0", "75.0", "60.0", "30.0",
	}

	ldtFile := &LDTFile{
		Header: LDTHeader{
			NumberOfCPlanes:             3,
			NumberOfLuminousIntensities: 4,
		},
	}

	err := parser.parsePhotometry(lines, 0, ldtFile)
	if err != nil {
		t.Fatalf("parsePhotometry failed: %v", err)
	}

	// Verify C-plane angles
	expectedCPlanes := []float64{0.0, 90.0, 180.0}
	if len(ldtFile.Photometry.CPlaneAngles) != len(expectedCPlanes) {
		t.Errorf("Expected %d C-plane angles, got %d", len(expectedCPlanes), len(ldtFile.Photometry.CPlaneAngles))
	}
	for i, expected := range expectedCPlanes {
		if ldtFile.Photometry.CPlaneAngles[i] != expected {
			t.Errorf("Expected C-plane angle[%d] = %f, got %f", i, expected, ldtFile.Photometry.CPlaneAngles[i])
		}
	}

	// Verify gamma angles
	expectedGammas := []float64{0.0, 30.0, 60.0, 90.0}
	if len(ldtFile.Photometry.GammaAngles) != len(expectedGammas) {
		t.Errorf("Expected %d gamma angles, got %d", len(expectedGammas), len(ldtFile.Photometry.GammaAngles))
	}
	for i, expected := range expectedGammas {
		if ldtFile.Photometry.GammaAngles[i] != expected {
			t.Errorf("Expected gamma angle[%d] = %f, got %f", i, expected, ldtFile.Photometry.GammaAngles[i])
		}
	}

	// Verify luminous intensity distribution
	// Data is transposed from C-plane first to gamma first
	expectedValues := [][]float64{
		{100.0, 90.0, 80.0}, // Gamma 0: values from C-plane 0, 1, 2
		{95.0, 85.0, 75.0},  // Gamma 30: values from C-plane 0, 1, 2
		{80.0, 70.0, 60.0},  // Gamma 60: values from C-plane 0, 1, 2
		{50.0, 40.0, 30.0},  // Gamma 90: values from C-plane 0, 1, 2
	}

	if len(ldtFile.Photometry.LuminousIntensityDistribution) != len(expectedValues) {
		t.Errorf("Expected %d gamma rows, got %d", len(expectedValues), len(ldtFile.Photometry.LuminousIntensityDistribution))
	}

	for i, expectedRow := range expectedValues {
		if len(ldtFile.Photometry.LuminousIntensityDistribution[i]) != len(expectedRow) {
			t.Errorf("Expected %d values in row %d, got %d", len(expectedRow), i, len(ldtFile.Photometry.LuminousIntensityDistribution[i]))
		}
		for j, expected := range expectedRow {
			if ldtFile.Photometry.LuminousIntensityDistribution[i][j] != expected {
				t.Errorf("Expected intensity[%d][%d] = %f, got %f", i, j, expected, ldtFile.Photometry.LuminousIntensityDistribution[i][j])
			}
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"123.45", 123.45},
		{"123,45", 123.45},     // European decimal separator
		{"  123.45  ", 123.45}, // With whitespace
		{"0", 0.0},
		{"invalid", 0.0}, // Invalid input should return 0
		{"", 0.0},        // Empty input should return 0
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseFloat(tt.input)
			if result != tt.expected {
				t.Errorf("ParseFloat(%q) = %f, expected %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"  456  ", 456}, // With whitespace
		{"0", 0},
		{"invalid", 0}, // Invalid input should return 0
		{"", 0},        // Empty input should return 0
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseInt(tt.input)
			if result != tt.expected {
				t.Errorf("ParseInt(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsNumericLine(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123.45", true},
		{"123,45", true}, // European decimal
		{"123 456 789", true},
		{"123.45 67,89", true},
		{"", false},
		{"abc def", false},
		{"123 abc", false},    // Mixed - should be false since less than half are numeric
		{"123 456 abc", true}, // 2/3 numeric, which is > 0.5
		{"123 456", true},     // 2/2 numeric
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isNumericLine(tt.input)
			if result != tt.expected {
				t.Errorf("isNumericLine(%q) = %t, expected %t", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsDateLike(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"13 Mar 2023 14:58:32", true},
		{"01/01/2023", true},
		{"2023-03-13", true},
		{"13.03.2023", true},
		{"abc", false},
		{"123", false},   // Only digits, no separators
		{"/ / /", false}, // Only separators, no digits
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isDateLike(tt.input)
			if result != tt.expected {
				t.Errorf("isDateLike(%q) = %t, expected %t", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseLDTFile_Integration is covered by TestParse_Integration

func TestParse_Integration(t *testing.T) {
	parser := NewParser()

	// Create a minimal valid LDT file
	ldtData := `Company;Test
1
0
2
180.0
3
60.0
Test measurement
Test luminaire
12345
test.ldt
01 Jan 2023/TestLab
100
50
25
80
40
0
0
0
0
100.0
90.0
1.0
0
1
1
LED
1000.0
3000K
80
10.0
0.0
180.0
0.0
30.0
90.0
100.0 90.0
80.0 70.0
60.0 50.0`

	commonData, err := parser.Parse([]byte(ldtData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify conversion to common model
	if commonData.Metadata.Manufacturer != "Company" {
		t.Errorf("Expected manufacturer 'Company', got '%s'", commonData.Metadata.Manufacturer)
	}
	if commonData.Metadata.CatalogNumber != "12345" {
		t.Errorf("Expected catalog number '12345', got '%s'", commonData.Metadata.CatalogNumber)
	}
	if commonData.Photometry.PhotometryType != "C" {
		t.Errorf("Expected photometry type 'C', got '%s'", commonData.Photometry.PhotometryType)
	}
	if len(commonData.Photometry.VerticalAngles) != 3 {
		t.Errorf("Expected 3 vertical angles, got %d", len(commonData.Photometry.VerticalAngles))
	}
	if len(commonData.Photometry.HorizontalAngles) != 2 {
		t.Errorf("Expected 2 horizontal angles, got %d", len(commonData.Photometry.HorizontalAngles))
	}
}

func TestParseErrors(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name string
		data string
	}{
		{
			name: "Empty file",
			data: "",
		},
		{
			name: "Too few lines",
			data: "Line 1\nLine 2\nLine 3",
		},
		{
			name: "Invalid header - zero C planes",
			data: strings.Repeat("Test\n", 4) + "0\n" + strings.Repeat("Test\n", 20),
		},
		{
			name: "Invalid header - zero luminous intensities",
			data: strings.Repeat("Test\n", 6) + "0\n" + strings.Repeat("Test\n", 20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse([]byte(tt.data))
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
		})
	}
}
