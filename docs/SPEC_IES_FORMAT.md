# IES Format Specification Document

## Overview

This document outlines the specifications for the IES (Illuminating Engineering Society) photometric data format as supported by the illuminate project, based on ANSI/IES LM-63 standard.

---

## 1. Standard Reference

### 1.1 IES LM-63 Versions

| Version | Year | Status | Notes |
|---------|------|--------|-------|
| IES LM-63-1986 | 1986 | Superseded | Original version |
| IES LM-63-1995 | 1995 | Superseded | Second revision |
| IES LM-63-2002 | 2002 | Superseded | Third revision |
| IES LM-63-2019 | 2019 | Current | Fifth revision (latest) |

### 1.2 Related Standards

| Standard | Title |
|----------|-------|
| ANSI/IES LM-63-19 | IES Standard File Format for Electronic Transfer of Photometric Data |
| ANSI/IES TM-33-18 | Standard Format for Electronic Transfer of Luminaire Optical Data (XML - future) |
| IES LM-79-19 | Electrical and Photometric Measurements of Solid-State Lighting |
| IES LM-80-20 | Measuring Lumen Maintenance of LED Light Sources |

---

## 2. File Format Structure

### 2.1 General Structure

An IES file consists of:
1. **Header line** - Version identifier
2. **Keyword section** - Metadata about the luminaire
3. **TILT specification** - Tilt angle handling
4. **Main data line** - Photometric parameters
5. **Lamp data line** - Lamp specifications
6. **Angle arrays** - Vertical and horizontal angles
7. **Candela values** - Photometric intensity data

### 2.2 Sample File Structure

```
IESNA:LM-63-2002
[TEST] P2DG220923056-10_IESNA2002
[TESTLAB] Bay Area Compliance Labs Corp.
[TESTDATE] 2022-09-30
[ISSUEDATE] 2022-09-30 19:05:45
[MANUFAC] Sylvania Schreder
[LAMPPOSITION] 0,0
[OTHER] EVERFINE GO-R5000_V2 SYSTEM
[LUMCAT] StreetLED3 17W 4K Aeroscreen Visor
[LUMINAIRE] StreetLED3 17W 4K Aeroscreen Visor
[LAMPCAT] LED Module 4K
[LAMP] LED Module 4K 
[OTHER] Total Luminous Flux 2458 lm. Not suitable to scale for other SSL modules
TILT=NONE
1 -1 1 181 73 1 2 0.280 0.285 0.000
1.000 1 16.8
<vertical_angles>
<horizontal_angles>
<candela_values>
```

---

## 3. Header Line

### 3.1 Format

```
IESNA:LM-63-<yyyy>
```

### 3.2 Valid Versions

- `IESNA:LM-63-1986`
- `IESNA:LM-63-1991`
- `IESNA:LM-63-1995`
- `IESNA:LM-63-2002`
- `IESNA:LM-63-2019`

### 3.3 Example

```
IESNA:LM-63-2002
```

---

## 4. Keyword Section

### 4.1 Keyword Format

Keywords are enclosed in square brackets:

```
[KEYWORD] value
```

### 4.2 Standard Keywords

| Keyword | Description | Example |
|---------|-------------|---------|
| [TEST] | Test report number | P2DG220923056-10_IESNA2002 |
| [TESTLAB] | Testing laboratory | Bay Area Compliance Labs Corp. |
| [TESTDATE] | Date of test | 2022-09-30 |
| [ISSUEDATE] | Issue date | 2022-09-30 19:05:45 |
| [MANUFAC] | Manufacturer | Sylvania Schreder |
| [LUMCAT] | Luminaire catalog number | StreetLED3 17W 4K |
| [LUMINAIRE] | Luminaire description | StreetLED3 17W 4K Aeroscreen |
| [LAMPCAT] | Lamp catalog number | LED Module 4K |
| [LAMP] | Lamp type description | LED Module 4K |
| [BALLAST] | Ballast catalog/type | - |
| [BALLASTCAT] | Ballast catalog number | - |
| [LAMPPOSITION] | Lamp position (X, Y) | 0,0 |
| [OTHER] | Additional information | Total Luminous Flux 2458 lm |
| [FILEFORMAT] | File format version | - |
| [FILEGENINFO] | File generation info | - |

### 4.4 Custom Keywords

Additional keywords may be present depending on the manufacturer.

---

## 5. TILT Specification

### 5.1 Format

```
TILT=<value>
```

### 5.2 Valid Values

| Value | Description |
|-------|-------------|
| `NONE` | No tilt (horizontal operation) - most common |
| `INCLUDE` | Tilt data included in file |
| `<filename>` | External file containing tilt data |

### 5.3 Example

```
TILT=NONE
```

---

## 6. Main Data Line

### 6.1 Format

```
<number_of_lamps> <lumens_per_lamp> <candela_multiplier> <vertical_angles> <horizontal_angles> 
<photometric_type> <units_type> <width> <length> <height>
```

### 6.2 Parameters

| Position | Parameter | Description | Typical Values |
|----------|-----------|-------------|----------------|
| 1 | Number of lamps | Number of lamps in luminaire | 1 |
| 2 | Lumens per lamp | Luminous flux per lamp | -1 (means use actual) |
| 3 | Candela multiplier | Multiplier for all values | 1.0 |
| 4 | Vertical angles | Number of vertical angles | 181 |
| 5 | Horizontal angles | Number of horizontal angles | 73 |
| 6 | Photometric type | Type of photometry | 1 (Type C) |
| 7 | Units type | Units of measurement | 2 (Metric) |
| 8 | Width | Luminaire width | 0 |
| 9 | Length | Luminaire length | 0 |
| 10 | Height | Luminaire height | 0 |

### 6.3 Photometric Types

| Type | Description | Application |
|------|-------------|-------------|
| 1 | Type C (horizontal) | Road lighting, general exterior |
| 2 | Type B (beta) | Not commonly used |
| 3 | Type A (alpha) | Not commonly used |

### 6.4 Units Types

| Type | Description |
|------|-------------|
| 1 | English (feet) |
| 2 | Metric (meters) |

### 6.5 Example

```
1 -1 1 181 73 1 2 0.280 0.285 0.000
```

---

## 7. Lamp Data Line

### 7.1 Format

```
<lamp_factor> <number_of_lamps> <watts>
```

### 7.2 Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| Lamp factor | Conversion factor | 1.000 |
| Number of lamps | Number of lamps | 1 |
| Watts | Power consumption | 16.8 |

### 7.3 Example

```
1.000 1 16.8
```

---

## 8. Angle Arrays

### 8.1 Vertical Angles

- Range: 0° to 180°
- Typically measured from nadir (0°) to zenith (180°)
- Common increments: 1°, 2°, 5°, or 10°
- Must be in ascending order

### 8.2 Horizontal Angles

- Range: 0° to 360°
- Full rotation for Type C photometry
- Must be in ascending order

### 8.3 Example

```
Vertical (0-180°):
0.0 1.0 2.0 3.0 ... 180.0

Horizontal (0-360°):
0.0 5.0 10.0 15.0 ... 355.0 360.0
```

---

## 9. Candela Data Matrix

### 9.1 Organization

Candela values are organized as a 2D matrix:
- Rows = Horizontal angles
- Columns = Vertical angles

### 9.2 Data Organization

The data is organized by horizontal planes:
- First all vertical angles at horizontal angle 0°
- Then all vertical angles at horizontal angle 1°
- Continue until all horizontal angles complete

### 9.3 Units

Candela values are typically given as:
- Absolute values (actual candela), OR
- Normalized values (candela per 1000 lumens)

### 9.4 Example

```
First horizontal plane (0°):
471.33 470.89 471.07 ... 0.75

Second horizontal plane (5°):
471.33 470.97 471.35 ... 0.75
```

---

## 10. Database Schema Requirements

### 10.1 Luminaire Model

```go
type Luminaire struct {
    ID               int64           `json:"id"`
    Manufacturer     string          `json:"manufacturer"`     // [MANUFAC]
    Model            string          `json:"model"`           // [LUMCAT]
    CatalogNumber    string          `json:"catalog_number"`  // [LAMPCAT]
    LuminaireDesc    string          `json:"luminaire_description"` // [LUMINAIRE]
    LampType         string          `json:"lamp_type"`       // [LAMP]
    LampCatalog      string          `json:"lamp_catalog"`    // [LAMPCAT]
    Ballast          string          `json:"ballast"`         // [BALLAST]
    TestLab          string          `json:"test_lab"`        // [TESTLAB]
    TestNumber       string          `json:"test_number"`     // [TEST]
    IssueDate        string          `json:"issue_date"`      // [ISSUEDATE]
    TestDate         string          `json:"test_date"`       // [TESTDATE]
    LuminaireCandela string          `json:"luminaire_candela"`
    LampPosition     string          `json:"lamp_position"`   // [LAMPPOSITION]
    Symmetry         int             `json:"symmetry"`
    PhotometricType  PhotometricType `json:"photometric_type"` // 1=Type C
    UnitsType        UnitsType       `json:"units_type"`      // 1=Imperial, 2=Metric
    ConversionFactor float64         `json:"conversion_factor"`
    InputWatts       float64         `json:"input_watts"`     // From lamp data line
    LuminousFlux     float64         `json:"luminous_flux"`  // Calculated or from metadata
    ColorTemp        int             `json:"color_temp"`     // Not in IES, may need separate field
    CRI              int             `json:"cri"`            // Not in IES, may need separate field
    FormatType       string          `json:"format_type"`   // "IES (IESNA LM-63)"
    SymmetryFlag     int             `json:"symmetry_flag"`
    FileHash         string          `json:"file_hash"`
    OriginalFilename string          `json:"original_filename"`
}
```

### 10.2 PhotometricData Model

```go
type PhotometricData struct {
    ID                  int64     `json:"id"`
    LuminaireID         int64     `json:"luminaire_id"`
    VerticalAngles      string    `json:"vertical_angles"`   // JSON array
    HorizontalAngles    string    `json:"horizontal_angles"` // JSON array
    CandelaValues       string    `json:"candela_values"`    // JSON 2D array
    NumVerticalAngles   int       `json:"num_vertical_angles"`
    NumHorizontalAngles int       `json:"num_horizontal_angles"`
}
```

---

## 11. Photometric Type Details

### 11.1 Type C (Photometric Type = 1)

Most common type for road and area lighting:

- **Vertical angles**: 0° to 180° (nadir to zenith)
- **Horizontal angles**: 0° to 360° (full rotation)
- **Reference direction**: Perpendicular to luminaire mounting
- **Application**: Road lighting, parking areas, sports lighting

### 11.2 Coordinate System

For Type C:
- γ (gamma) = vertical angle from nadir (0°) to zenith (180°)
- C = horizontal angle, typically starting at 0° (perpendicular to luminaire length)

---

## 12. Parser Implementation Notes

### 12.1 Current Implementation

Located: `internal/parser/ies.go`

**Parsing Flow**:
1. Read header line (IESNA:LM-63-xxxx)
2. Parse all keyword lines into map
3. Parse TILT line
4. Parse main data line
5. Parse lamp data line
6. Parse vertical angles array
7. Parse horizontal angles array
8. Parse remaining lines as candela values

**Keyword Mapping**:
- `TEST` → TestNumber
- `TESTLAB` → TestLab
- `MANUFAC` → Manufacturer
- `ISSUEDATE` → IssueDate
- `LUMCAT` → Model
- `LUMINAIRE` → LuminaireDesc
- `LAMPCAT` → LampCatalog
- `LAMP` → LampType
- `BALLAST` → Ballast
- `LAMPPOSITION` → LampPosition

### 12.2 File Writing

The parser can also write IES files:
- Output version: IESNA:LM-63-2002
- Writes all available metadata as keywords
- TILT=NONE by default
- Outputs candela matrix in standard format

---

## 13. Supported Extensions

The illuminate project supports the following photometric formats:

| Format | Extension | Standard |
|--------|-----------|----------|
| IES | .ies | IESNA LM-63 |
| CIE | .cie | CIE 102 |
| LDT | .ldt | Eulumdat |

---

## 14. Limitations

### 14.1 Current Parser Limitations

- Limited to Type C photometry
- Does not parse TILT=INCLUDE (assumes TILT=NONE)
- Color temperature (CCT) and CRI not in standard IES fields
- Spectral data (IES TM-27) not supported

### 14.2 Future Enhancements

- Support for TILT=INCLUDE
- Spectral data parsing (TM-27)
- Extended metadata fields

---

## 15. Example Files

### 15.1 Road Luminaire

```
IESNA:LM-63-2002
[TEST] P2DG220923056-10_IESNA2002
[TESTLAB] Bay Area Compliance Labs Corp.
[MANUFAC] Sylvania Schreder
[LUMCAT] StreetLED3 17W 4K Aeroscreen Visor
[LUMINAIRE] StreetLED3 17W 4K Aeroscreen Visor
[LAMP] LED Module 4K
TILT=NONE
1 -1 1 181 73 1 2 0.280 0.285 0.000
1.000 1 16.8
<181 vertical angles from 0 to 180>
<73 horizontal angles from 0 to 360>
<13,213 candela values>
```

---

## 16. References

1. ANSI/IES LM-63-19 - IES Standard File Format for Electronic Transfer of Photometric Data
2. IESNA LM-63-2002 - Standard File Format for Electronic Transfer of Photometric Data
3. CIE 102:1993 - Recommended File Format for Electronic Transfer of Luminaire Photometric Data
4. AGI32 Photometric Toolbox Documentation - IES Format

---

## Appendix A: Glossary

| Term | Definition |
|------|------------|
| Candela (cd) | SI unit of luminous intensity |
| Lumen (lm) | SI unit of luminous flux |
| Lux (lx) | SI unit of illuminance (lm/m²) |
| Goniophotometer | Instrument for measuring light distribution |
| Type C Photometry | Horizontal measurement plane system |
| Nadir | Point directly below luminaire (0°) |
| Zenith | Point directly above luminaire (180°) |

---

*Document Version: 1.0*
*Date: 2026-03-09*
*Source: ANSI/IES LM-63-19, IES LM-63-2002*
