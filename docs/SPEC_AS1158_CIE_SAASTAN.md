# AS1158 and CIE/SAASTAN Specification Document

## Overview

This document outlines the requirements for cie/saastan photometric data as per AS/NZS 1158 standards for the illuminate project.

---

## 1. Standards Reference

### 1.1 AS/NZS 1158 Series

| Standard | Title | Purpose |
|----------|-------|---------|
| AS/NZS 1158.0:2005 | Introduction | Definitions, lighting categories |
| AS/NZS 1158.1.1:2022 | Category V Lighting | Vehicular traffic - Performance & design |
| AS/NZS 1158.1.2:2010 | Category V Guide | Design, installation, operation, maintenance |
| AS/NZS 1158.2:2020 | Computer Procedures | LTP calculation methods (includes SAA STAN) |
| AS/NZS 1158.3.1:2020 | Category P Lighting | Pedestrian area lighting - Performance & design |
| AS/NZS 1158.4:2015 | Pedestrian Crossings | Lighting of pedestrian crossings |
| AS/NZS 1158.5 | Tunnels | Tunnels and underpasses |

### 1.2 Related Standards

| Standard | Title |
|----------|-------|
| CIE 102:1993 | Recommended File Format for Electronic Transfer of Luminaire Photometric Data |
| CIE 115:2010 | Lighting of roads for motor and pedestrian traffic |
| EN 13201-5 | Road Lighting - Energy Performance Indicators |

---

## 2. Lighting Categories

### 2.1 Category V (Vehicular Traffic)

**Definition**: Roads where the visual requirements of motorists are dominant.

**Typical Applications**:
- Traffic routes
- Main roads
- Arterial roads
- High-speed corridors

**Light Technical Parameters (LTPs)**:
- Average carriageway luminance (L) - cd/m²
- Overall uniformity (U₁)
- Longitudinal uniformity (U₂)
- Threshold increment (TI) - glare metric
- Disability glare (G)
- Surround ratio (SR)

### 2.2 Category P (Pedestrian Area)

**Definition**: Roads where the visual requirements of pedestrians are dominant.

**Typical Applications**:
- Local roads
- Shopping precincts
- Pedestrian paths
- Car parks
- Public spaces

**Light Technical Parameters (LTPs)**:
- Average horizontal illuminance (E) - lux
- Minimum illuminance (Eₘᵢₙ)
- Illuminance uniformity (U₂)
- Vertical illuminance (for facial recognition)

---

## 3. Photometric Data Formats

### 3.1 CIE Format (CIE 102)

The illuminate project supports CIE format files (.cie).

**File Structure**:
```
<symmetry_flag> <format_type> <0> <luminaire_name> <luminous_flux>
<candela_row_1>
<candela_row_2>
...
<candela_row_n>
```

**Header Line**:
- Column 1-3: Symmetry flag (1 = no symmetry, 2 = bilateral, 3 = rotational)
- Column 4-5: Format type (0 = intensities per 1000 lumens)
- Column 7+: Luminaire description and luminous flux

**Data Matrix**:
- Vertical angles: 0° to 170° in 10° increments (18 rows typically)
- Horizontal angles: 0° to 340° in 10° increments (36 columns typically)
- Values: Candela per 1000 lamp lumens

**Symmetry Types**:
| Flag | Description |
|------|-------------|
| 1 | No symmetry (full C-gon) |
| 2 | Bilateral symmetry (C-2) |
| 3 | Rotational symmetry (C-1) |

### 3.2 SAASTAN Format

SAASTAN is the Australian variant of CIE format used in AS/NZS 1158 calculations.

**Key Differences from Standard CIE**:
- Vertical angles: 0° to 170° in 10° increments
- Horizontal angles: 0° to 360° (full rotation for Type C photometry)
- Intensity values normalized to 1000 lumens
- Additional metadata for road surface classification

### 3.3 Other Supported Formats

| Format | Extension | Standard |
|--------|----------|----------|
| IES | .ies | IESNA LM-63 |
| LDT | .ldt | Eulumdat |

---

## 4. Database Schema Requirements

### 4.1 Luminaire Model

```go
type Luminaire struct {
    ID               int64
    Manufacturer     string      // Luminaire manufacturer
    Model            string      // Model number
    CatalogNumber    string      // Catalog/reference number
    LuminaireDesc    string      // Full description
    LampType         string      // Light source type
    LampCatalog      string      // Lamp catalog number
    Ballast          string      // Ballast type
    TestLab          string      // Testing laboratory
    TestNumber       string      // Test report number
    IssueDate        string      // Report issue date
    TestDate         string      // Test date
    LuminaireCandela string      // Candela classification
    LampPosition     string      // Lamp positioning
    Symmetry         int         // Symmetry type (1, 2, or 3)
    PhotometricType  PhotometricType  // C=1, B=2, A=3
    UnitsType        UnitsType   // Metric or Imperial
    ConversionFactor float64     // Unit conversion
    InputWatts       float64     // Power consumption (W)
    LuminousFlux     float64     // Total luminous flux (lm)
    ColorTemp        int         // CCT (K)
    CRI              int         // Color Rendering Index
    FormatType       string      // "CIE", "IES", "LDT"
    SymmetryFlag     int         // CIE symmetry flag
    FileHash         string      // SHA256 of file
    OriginalFilename string      // Original file name
}
```

### 4.2 Photometric Data Model

```go
type PhotometricData struct {
    ID                  int64
    LuminaireID         int64
    VerticalAngles      string    // JSON array of angles (°)
    HorizontalAngles    string    // JSON array of angles (°)
    CandelaValues       string    // JSON 2D array
    NumVerticalAngles   int
    NumHorizontalAngles int
}
```

---

## 5. Calculation Requirements

### 5.1 AS/NZS 1158.2 LTP Calculations

#### Category V Calculations (from AS/NZS 1158.2:2020)

1. **Luminance-based LTPs**:
   - Average carriageway luminance (L)
   - Luminance uniformity (U₁ = Lₘᵢₙ/Lₐᵥₑ)
   - Longitudinal uniformity (U₂ = Lₘᵢₙ/Lₘₐₓ on centerline)

2. **Illuminance-based LTPs** (alternative):
   - Average illuminance (E)
   - Illuminance uniformity (U₁ = Eₘᵢₙ/Eₐᵥₑ)

3. **Glare LTPs**:
   - Threshold increment (TI)
   - Disability glare (G)

4. **Flux-based LTPs**:
   - Utilance
   - Total installed flux

#### Category P Calculations (from AS/NZS 1158.2:2020)

1. **Illuminance-based LTPs**:
   - Average horizontal illuminance (Eₐᵥₑ)
   - Minimum illuminance (Eₘᵢₙ)
   - Illuminance uniformity (U₂ = Eₘᵢₙ/Eₐᵥₑ)

2. **Flux-based LTPs**:
   - Waste light ratio
   - Surround illumination

### 5.2 Road Surface Classification (r-tables)

AS/NZS 1158 uses road surface reflectance classifications:

| Class | Description | Q0 Value |
|------|-------------|----------|
| R1 | Dry road surface, Portland cement concrete | 0.10 |
| R2 | Dry road surface, hot rolled asphalt | 0.07 |
| R3 | Wet road surface, wearing course | 0.05 |
| R4 | Asphalt with exposed aggregate | 0.08 |

### 5.3 Maintenance Factors

| Factor | Description | Typical Values |
|--------|-------------|----------------|
| LLMF | Lamp Lumen Maintenance | 0.70-0.95 |
| LSF | Lamp Survival Factor | 0.90-0.98 |
| LMFD | Luminaire Maintenance Factor | 0.60-0.95 |
| MF | Overall Maintenance Factor | 0.50-0.85 |

---

## 6. Energy Performance (AS/NZS 1158.3.1:2020)

### 6.1 Power Density Indicator (PDI)

```
PDI = Total installed load power / Task area
```

Units: W/m²

### 6.2 Annual Energy Consumption

Calculated using:
- Operating hours per year
- Control factor (dimming schedules)
- Energy tariff

---

## 7. File Format Specifications

### 7.1 CIE File Example

```
   1   1   0        StreetLED3 17W 4K Aero P2DG220923057-10 - 2458 lm
  192 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192 192
  192 192 192 192 192 192 192 192 192 192 192 191 190 191 192 192 193 194
  ...
```

- Line 1: Header with symmetry, format, name, and flux
- Lines 2+: Candela values (cd/1000lm) at 10° intervals

### 7.2 Parsing Requirements

For illuminate to correctly parse CIE/SAASTAN files:

1. **Symmetry Detection**: Parse flag from header
2. **Flux Extraction**: Extract luminous flux from description
3. **Matrix Construction**: Build 2D candela matrix
4. **Angle Generation**: 0-170° vertical, 0-340° horizontal

---

## 8. Implementation Notes

### 8.1 Current illuminate Parser

Located: `internal/parser/cie.go`

- Parses CIE format files
- Extracts metadata: symmetry, format type, name, luminous flux
- Constructs vertical angles (0° to 170°, step 10°)
- Constructs horizontal angles (0° to 340°, step 10°)
- Calculates file hash for deduplication

### 8.2 Supported Extensions

```go
func GetSupportedExtensions() []string {
    return []string{".ies", ".cie", ".ldt"}
}
```

---

## 9. References

1. AS/NZS 1158.0:2005 - Introduction
2. AS/NZS 1158.1.1:2022 - Category V Lighting
3. AS/NZS 1158.2:2020 - Computer Procedures
4. AS/NZS 1158.3.1:2020 - Category P Lighting
5. CIE 102:1993 - Recommended File Format for Electronic Transfer of Luminaire Photometric Data
6. CIE 115:2010 - Lighting of roads for motor and pedestrian traffic

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| Candela (cd) | SI unit of luminous intensity |
| Lumen (lm) | SI unit of luminous flux |
| Lux (lx) | SI unit of illuminance (lm/m²) |
| Luminance (cd/m²) | Luminous intensity per unit area |
| CCT | Correlated Color Temperature |
| CRI | Color Rendering Index |
| LTP | Light Technical Parameter |
| SAA STAN | Standards Australia lighting calculation program |
| CIE | International Commission on Illumination |
| TI | Threshold Increment (glare metric) |
| U₁ | Overall uniformity |
| U₂ | Longitudinal/minimum uniformity |

---

*Document Version: 1.0*
*Date: 2026-03-09*
*Source: AS/NZS 1158 Series, CIE Standards*
