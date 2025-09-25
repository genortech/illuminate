# EULUMDAT (LDT) Luminaire Format Specification

## Overview

The EULUMDAT format (also known as the LDT format) is a standardized ASCII text file format used for exchanging photometric data of luminaires. It was proposed by Axel Stockmar (Light Consult Inc., Berlin) in 1990 and serves as the European equivalent to the IES file format. The file extension is `.ldt`.

## Purpose and Applications

The LDT format is primarily used for:
- Exchanging photometric data between lighting manufacturers and designers
- Importing luminaire data into lighting design software (DIALux, Relux, AGI32)
- Storing essential luminaire characteristics including light distribution, intensity, and physical dimensions
- Enabling accurate lighting simulations and calculations in professional lighting applications

## File Structure

The EULUMDAT file is a plain ASCII text format where each line represents a specific data element. The format consists of 30+ lines, each containing specific information about the luminaire.

### Complete Line-by-Line Specification

| Line | Description | Max Characters |
|------|-------------|----------------|
| 1 | Company identification/database/version/format identification | Max 78 |
| 2 | **Ityp** - Type indicator:<br>1 = Point source with symmetry about vertical axis<br>2 = Linear luminaire<br>3 = Point source with any other symmetry | 1 |
| 3 | **Isym** - Symmetry indicator:<br>0 = No symmetry<br>1 = Symmetry about vertical axis<br>2 = Symmetry to plane C0-C180<br>3 = Symmetry to plane C90-C270<br>4 = Symmetry to both C0-C180 and C90-C270 planes | 1 |
| 4 | **Mc** - Number of C-planes between 0° and 360°<br>(Usually 24 for interior luminaires, 36 for road lighting) | 2 |
| 5 | **Dc** - Distance between C-planes in degrees<br>(Dc = 0 for non-equidistant C-planes) | 5 |
| 6 | **Ng** - Number of luminous intensities in each C-plane<br>(Usually 19, 37, or 73) | 2 |
| 7 | **Dg** - Distance between luminous intensities per C-plane in degrees<br>(Dg = 0 for non-equidistant luminous intensities) | 5 |
| 8 | Measurement report number/test certificate | Max 78 |
| 9 | Luminaire name | Max 78 |
| 10 | Luminaire number | Max 78 |
| 11 | File name | 8 |
| 12 | Date/user/person responsible | Max 78 |
| 13 | Length/diameter of luminaire (mm) | 4 |
| 14 | Width of luminaire (mm)<br>(0 for circular luminaires) | 4 |
| 15 | Height of luminaire (mm) | 4 |
| 16 | Length/diameter of luminous area (mm) | 4 |
| 17 | Width of luminous area (mm)<br>(0 for circular luminous area) | 4 |
| 18 | Height of luminous area C0-plane (mm) | 4 |
| 19 | Height of luminous area C90-plane (mm) | 4 |
| 20 | Height of luminous area C180-plane (mm) | 4 |
| 21 | Height of luminous area C270-plane (mm) | 4 |
| 22 | **DFF** - Downward flux fraction (%) | 4 |
| 23 | **LORL** - Light output ratio luminaire (%) | 4 |
| 24 | Conversion factor for luminous intensities | 6 |
| 25 | Tilt of luminaire during measurement (degrees)<br>(Important for road lighting luminaires) | 6 |
| 26 | **n** - Number of standard sets of lamps | 4 |
| 26a | Number of lamps (for each set) | n × 4 |
| 26b | Type of lamps | n × 24 |
| 26c | Total luminous flux of lamps (lm) | n × 12 |
| 26d | Color temperature of lamps (K) | n × 16 |
| 26e | Color rendering index (CRI) | n × 6 |
| 26f | Total system power/wattage including ballast (W) | n × 8 |
| 27 | Direct ratios for room indices k = 0.6 to 5.0<br>(For utilization factor method) | 10 × 7 |
| 28 | Angles C (starting with 0°) | Mc × 6 |
| 29 | Angles G (starting with 0°) | Ng × 6 |
| 30 | Luminous intensity distribution (cd/1000 lumens) | Variable |

## Key Data Elements

### Physical Dimensions
- **Lines 13-15**: Overall luminaire dimensions (length/diameter, width, height)
- **Lines 16-21**: Luminous area dimensions for different C-planes
- All dimensions are specified in millimeters

### Photometric Data
- **Line 22**: Downward flux fraction as percentage
- **Line 23**: Light output ratio of luminaire as percentage
- **Line 24**: Conversion factor for luminous intensity values
- **Line 30**: Complete luminous intensity distribution matrix

### Lamp Information
- **Lines 26a-26f**: Comprehensive lamp data including:
  - Number and type of lamps
  - Total luminous flux
  - Color temperature
  - Color rendering index
  - Power consumption

### Angular Data
- **Line 28**: C-plane angles (horizontal angles around vertical axis)
- **Line 29**: G-plane angles (vertical angles from nadir to zenith)
- **Line 30**: Intensity values for each C-plane and G-angle combination

## Symmetry Handling

The format supports various symmetry types to optimize file size:

1. **No symmetry (Isym = 0)**: Full 360° data required
2. **Vertical axis symmetry (Isym = 1)**: Only one C-plane needed
3. **C0-C180 plane symmetry (Isym = 2)**: Half the C-planes required
4. **C90-C270 plane symmetry (Isym = 3)**: Half the C-planes required
5. **Both planes symmetry (Isym = 4)**: Quarter of the C-planes required

## Absolute Photometry Support

For LED luminaires and absolute photometry measurements:
- Line 26: Set number of lamp sets to 1
- Line 26a: Set number of lamps to negative value
- Line 26c: Contains total luminous flux of luminaire instead of individual lamps

## Data Format Requirements

- All data elements are saved as ASCII strings
- Each line is terminated with an end-of-string symbol
- Numerical values follow specified character limits
- Text fields support maximum character counts as specified

## Applications in Lighting Design

The LDT format enables:
- Accurate lighting simulations in professional software
- Calculation of illuminance levels and uniformity
- Assessment of light distribution patterns
- Energy efficiency analysis
- Compliance verification with lighting standards

## Software Compatibility

Common software supporting LDT format:
- **DIALux/DIALux Evo**: Primary lighting design software
- **Relux**: Professional lighting calculation software
- **AGI32**: Architectural and roadway lighting software
- **LDT Editor**: Free editing tool by DIAL
- **Photometric Toolbox**: Analysis and conversion tool

## File Creation and Editing

LDT files can be:
- Generated by goniophotometer measurement equipment
- Created using specialized software tools
- Edited with text editors (though specialized tools recommended)
- Converted from IES format using various conversion utilities

## Best Practices

1. **Accurate Dimensions**: Ensure physical and luminous dimensions are correct
2. **Complete Data**: Include all required lamp and photometric information
3. **Proper Naming**: Use descriptive luminaire and file names
4. **Symmetry Optimization**: Use appropriate symmetry settings to reduce file size
5. **Color Information**: Include accurate color temperature and CRI data
6. **Documentation**: Provide clear identification and measurement details

This specification provides the complete structure for EULUMDAT (LDT) files, enabling proper creation, interpretation, and use of photometric data in professional lighting applications.