# IES Photometric File Format Specification

## Overview

The IES format is the North American standard for photometric data exchange, created by the Illuminating Engineering Society of North America (IESNA). Commonly used in lighting simulation software, it enables photometric information to be reliably shared and validated across the industry.

## File Structure

The IES format uses plain ASCII text files with the extension `.ies` and follows the IESNA LM-63 standard. The most current specification is IESNA LM-63-2002, with earlier versions still in use.

### Complete Line-by-Line Specification

| Line | Description | Required |
|------|-------------|----------|
| 1 | Format Identifier (e.g., IESNA:LM-63-2002) | Yes |
| 2-n | Metadata Keywords ([TEST], [LAMP], [LUMINAIRE], etc.) | Optional |
| --- | --- | --- |
| TILT | Tilt Information (e.g., TILT=NONE, INCLUDE, or filename.TLT) | Yes |
| --- | --- | --- |
| Main Data | Photometric and geometric data (see below) | Yes |

#### Main Data Elements
- Number of lamps
- Lumens per lamp
- Candela multiplier
- Number of vertical angles
- Number of horizontal angles
- Photometric type (1=Type C, 2=Type B, 3=Type A)
- Units (1=Feet, 2=Meters)
- Luminaire width, length, height
- Ballast factor, ballast-lamp photometric factor
- Input watts
- Vertical angles (typically 0-180°)
- Horizontal angles (typically 0-360°)
- Candela values matrix (intensity at each angle pair)

#### Example IES File Structure
```
IESNA:LM-63-2002
[TEST] Test report details
[MANUFAC] Manufacturer ABC
[LUMCAT] LM-500LED
[LUMINAIRE] LED Area Light
TILT=NONE
1 9000 1.0 37 73 1 2 1.5 1.5 0.6
1.000 1.000 120
0.0 2.5 5.0 7.5 ... 90.0
0.0 5.0 10.0 ... 355.0
1234 1235 1236 ...
1240 1241 1242 ...
...
```

## Key Guidelines
- **Format Compliance:** Include correct version identifier
- **Metadata:** Provide relevant optional information
- **Angle and Intensity Data:** Ensure values and sequence are correct
- **Units:** Maintain consistency (feet or meters)
- **Validation:** Use software tools for automatic check

## Software Compatibility
- AGI32, DIALux, Relux, AutoCAD, Revit, 3ds Max, others

## Best Practices
- Include descriptive metadata
- Validate for software compatibility
- Follow IESNA LM-63 standard

