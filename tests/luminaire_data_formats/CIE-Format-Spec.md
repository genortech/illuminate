# CIE Photometric File Format Specification

## Overview

CIE formats are global standards set by the International Commission on Illumination, extensively used in Europe and Australia. CIE files convey luminaire performance via structured, often matrix-based text data adaptable for symmetric and asymmetric distributions.

## CIE 102-1993 Format

### Structure
ASCII file, typically `.cie` extension, containing explicit header fields and block data layout.

| Element | Description |
|---------|-------------|
| CIEF | Format identification (e.g. CIEF=CIE File Format, Version 1.0) |
| IDNM | Identification number |
| LUMN | Luminaire name |
| LAMP | Lamp description |
| NLPS | Number of lamp sets |
| TOLU | Total luminous flux |
| LLGE | Luminaire geometry code |
| BLID | Ballast identification |
| INPW | Input power (W) |
| INVO | Input voltage |
| INVA | Input current |
| TLME | Test method code |
| LSHP | Luminaire shape code |
| NLAV | Number of lamps |
| PHOT | Photometric data reference (INCLUDE/filename) |
| PTYP | Photometric type (A/B/C) |
| APOS | Axis position code |
| LUBA | Bare flux (lumens) |
| MULT | Intensity multiplier |
| BAFA | Ballast factor |
| NCON | Number of C-planes |
| NPLA | Number of gamma angles |
| CONA | C-plane angle list |
| GAMA | Gamma angle list |
| Data | Intensity values (matrix)

## CIE i-table Format (CIE 132)
Mandated for Australian/New Zealand road lighting standards: structured, fixed-column and row width matrix.

| Line | Description |
|------|-------------|
| 1 | Header Control Line (four integers) |
| 2+ | Data Matrix (intensity, 17 values/row, 4 chars/col.) |

**Units:** Intensity in cd/1000 lumens, azimuth/gamma angles specified per standard.

### Example CIE i-table File
```
1 1 0 8201
 139 139 139 139 139 139 139 139 139 ... (17 values)
 154 154 154 154 154 154 154 154 154 ...
 ...
```

## Key Guidelines
- Include all mandatory header fields
- Validate geometric and photometric codes
- Ensure intensity matrix matches angular references
- Comply strictly with CIE standard requirements

## Software Compatibility
- DIALux, Relux, Perfect Lite, various European and Australian lighting software

## Best Practices
- Provide detailed metadata
- Use correct matrix and angle formats
- Validate against CIE specification

