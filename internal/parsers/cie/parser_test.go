package cie

import (
	"illuminate/internal/formats/cie"
	"testing"
)

func TestParser_DetectFormat(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name               string
		data               string
		expectedConfidence float64
		expectedVersion    string
	}{
		{
			name: "Valid CIE i-table format",
			data: `   1   0   0        Test LED 17W 1000 lm
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`,
			expectedConfidence: 0.3,
			expectedVersion:    string(cie.VersionITable),
		},
		{
			name: "Valid CIE with different description",
			data: `   1   0   0        P14W LED 2172.2 lms
 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192
 192 192 192 192 192 192 192 192 192 192 191 190 191 192 192 193 194`,
			expectedConfidence: 0.3,
			expectedVersion:    string(cie.VersionITable),
		},
		{
			name:               "Invalid format - not CIE",
			data:               "This is not a CIE file",
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
		{
			name:               "Empty data",
			data:               "",
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
		{
			name: "Invalid header format",
			data: `invalid header format
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`,
			expectedConfidence: 0.0,
			expectedVersion:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence, version := parser.DetectFormat([]byte(tt.data))

			if confidence < tt.expectedConfidence {
				t.Errorf("DetectFormat() confidence = %v, want >= %v", confidence, tt.expectedConfidence)
			}

			if version != tt.expectedVersion {
				t.Errorf("DetectFormat() version = %v, want %v", version, tt.expectedVersion)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	validCIEData := `   1   0   0        Test LED 17W 1000 lm
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`

	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "Valid CIE data",
			data:    validCIEData,
			wantErr: false,
		},
		{
			name:    "Invalid data - empty",
			data:    "",
			wantErr: true,
		},
		{
			name:    "Invalid data - insufficient lines",
			data:    "   1   0   0        Test LED",
			wantErr: true,
		},
		{
			name: "Invalid header format",
			data: `invalid header
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse([]byte(tt.data))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Error("Parse() returned nil result")
				return
			}

			// Basic validation of parsed data
			if result.Metadata.Description == "" {
				t.Error("Expected non-empty description")
			}

			if result.Photometry.PhotometryType != "C" {
				t.Errorf("PhotometryType = %v, want 'C'", result.Photometry.PhotometryType)
			}

			if len(result.Photometry.VerticalAngles) == 0 {
				t.Error("Expected non-empty vertical angles")
			}

			if len(result.Photometry.HorizontalAngles) == 0 {
				t.Error("Expected non-empty horizontal angles")
			}

			if len(result.Photometry.CandelaValues) == 0 {
				t.Error("Expected non-empty candela values")
			}
		})
	}
}

func TestGetSupportedVersions(t *testing.T) {
	parser := NewParser()
	versions := parser.GetSupportedVersions()

	expectedVersions := []string{string(cie.Version102), string(cie.VersionITable)}

	if len(versions) != len(expectedVersions) {
		t.Errorf("GetSupportedVersions() returned %d versions, want %d", len(versions), len(expectedVersions))
	}

	for _, expected := range expectedVersions {
		found := false
		for _, version := range versions {
			if version == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedVersions() missing version %s", expected)
		}
	}
}

func TestParser_Validate(t *testing.T) {
	parser := NewParser()

	// This test uses the shared validator, so we just test that it's called correctly
	validCIEData := `   1   0   0        Test LED 17W 1000 lm
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135
 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135 135`

	data, err := parser.Parse([]byte(validCIEData))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	err = parser.Validate(data)
	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}
