# AS/NZS 1158 P Category Lighting Spacing Chart Generation Requirements Specification

## Executive Summary

This specification outlines the software requirements for generating spacing charts and tables for Category P (Pedestrian Area) lighting in accordance with AS/NZS 1158 Australian/New Zealand lighting standards. The requirements are derived from AS/NZS 1158.3.1:2020 performance standards and AS/NZS 1158.2:2020 computer calculation procedures.

## 1. Regulatory Framework

### 1.1 Primary Standards
- **AS/NZS 1158.3.1:2020**: Pedestrian area (Category P) lighting - Performance and design requirements
- **AS/NZS 1158.2:2020**: Computer procedures for the calculation of light technical parameters for Category V and Category P lighting
- **AS/NZS 1158.0**: General principles and recommendations

### 1.2 Compliance Requirements
- Software must comply with Section 4.2 "Fundamental Software Requirements" of AS/NZS 1158.2:2020
- Calculations must meet accuracy requirements specified in Section 4.5 of AS/NZS 1158.2:2020
- Output must demonstrate compliance with applicable Category P lighting subcategories (P1-P6, PR1-PR6, PC1-PC3)

## 2. Software Architecture Requirements

### 2.1 Core Calculation Engine
- **Illuminance-based calculations** for horizontal light distribution
- **Point-by-point grid analysis** across designated calculation fields
- **Photometric data processing** using standard I-table formats
- **Optimization algorithms** to determine maximum compliant spacings

### 2.2 Input Data Management
- **Photometric I-table support** in CIE/SAASTAN standard format
- **Luminaire database integration** with manufacturer data
- **Parameter validation** for all input values
- **Default value management** for common configurations

### 2.3 Calculation Field Specifications

#### 2.3.1 Grid Definition Requirements
- **Calculation field** must cover the area under consideration
- **Calculation points** must form a grid of uniformly and equally spaced test points in two dimensions
- **Grid density** must be sufficient to accurately represent illuminance distribution
- **Field boundaries** must extend appropriately beyond roadway edges

#### 2.3.2 Spatial Parameters
- **Road width range**: 2m to 40m road reserve calculations
- **Mounting height range**: 0.5m to 20m above roadway surface
- **Spacing analysis range**: Up to 120m maximum luminaire spacing
- **Calculation precision**: Appropriate rounding as per Section 4.5

## 3. Input Parameter Requirements

### 3.1 Luminaire Data
- **Photometric I-table file** (CIE/SAASTAN format)
- **Luminaire description** and manufacturer details
- **Initial luminous flux** (100-hour lamp rating)
- **Lamp type and wattage** specifications
- **Maintenance factor** (default 0.7 minimum maintained output)

### 3.2 Installation Parameters
- **Mounting height** (0.5m to 20m)
- **Luminaire arrangement**: Single-sided (0), Staggered (1), Opposite (2)
- **Upcast angle**: 0° to 60° above horizontal (default 5°)
- **Offset distance**: -12m to +30m from property line (* = over notional kerb at 1/4 road reserve width)
- **Road geometry**: Carriageway width, verge details

### 3.3 Lighting Category Selection
- **Standard categories**: P1, P2, P3, P4, P5, P6 (local roads)
- **Residential categories**: PR1, PR2, PR3, PR4, PR5, PR6 (residential areas)
- **Car park categories**: PC1, PC2, PC3 (outdoor car parks)
- **User-defined category**: Custom LTP values

## 4. Light Technical Parameters (LTP) Calculations

### 4.1 Primary Parameters
- **Eh (Average Horizontal Illuminance)**: Average of all calculation points in the area
- **Eph (Point Horizontal Illuminance)**: Minimum illuminance at any calculation point
- **Ue2 (Horizontal Illuminance Uniformity)**: Maximum/Average illuminance ratio
- **Epv (Point Vertical Illuminance)**: Vertical illuminance at 1.5m height (PC1, PC2 categories only)

### 4.2 Additional Requirements
- **Spill light calculations**: 50% of Eph at specified distances into abutting properties
- **Glare assessment**: Where specified by applicable subcategory
- **Upward waste light ratio (UWLR)**: Environmental impact considerations

### 4.3 Calculation Methodology
- **Grid-based analysis**: Systematic evaluation across uniform calculation grid
- **Interpolation algorithms**: For points between luminaire positions
- **Cumulative illuminance**: From all luminaires contributing to each calculation point
- **Worst-case analysis**: Identification of minimum performance locations

## 5. Spacing Chart Generation Requirements

### 5.1 Calculation Process
1. **Grid establishment**: Define calculation field with uniform point spacing
2. **Photometric processing**: Load and validate I-table data
3. **Iterative spacing analysis**: Calculate LTPs for increasing spacing values
4. **Compliance checking**: Verify each spacing against applicable LTP requirements
5. **Maximum determination**: Identify largest compliant spacing for each road width
6. **Table generation**: Compile results into structured spacing table
7. **Graph creation**: Generate visual spacing chart with compliance boundaries

### 5.2 Output Requirements

#### 5.2.1 Tabular Output
- **Spacing table format**: Road width vs. maximum compliant spacing
- **Performance metrics**: Eh, Eph, Ue2 values for each configuration
- **Compliance indicators**: Clear identification of compliant/non-compliant combinations
- **Additional data**: Average illuminance, uniformity ratios, efficiency metrics

#### 5.2.2 Graphical Output
- **Spacing graph**: Road reserve width (horizontal) vs. spacing (vertical)
- **Compliance zone**: Area under graph line indicates compliant combinations
- **Reference lines**: Vertical lines for common road widths
- **Annotations**: Design parameters and luminaire details

### 5.3 Validation and Quality Assurance
- **LTP verification**: Confirm all calculations meet specified requirements
- **Boundary condition checking**: Validate performance at spacing extremes
- **Interpolation accuracy**: Ensure smooth transitions between calculated points
- **Documentation trail**: Maintain calculation audit trail for verification

## 6. Software Implementation Requirements

### 6.1 User Interface
- **Windows-based GUI**: Modern, intuitive interface design
- **Input validation**: Real-time checking of parameter values
- **Progress indication**: Status display during lengthy calculations
- **Help system**: Context-sensitive guidance (F1 key activation)
- **Error handling**: Clear error messages with corrective guidance

### 6.2 Data Management
- **File import/export**: Standard file format support
- **Project management**: Save/load project configurations
- **Library integration**: Luminaire database management
- **Backup functionality**: Automatic data protection

### 6.3 Output Generation
- **Print formatting**: Professional report generation
- **File export**: Multiple format support (PDF, CSV, etc.)
- **Graph customization**: Adjustable scales, annotations, reference lines
- **Data export**: Raw calculation data access

## 7. Performance Requirements

### 7.1 Calculation Speed
- **Processing efficiency**: Reasonable calculation times for typical projects
- **Memory management**: Efficient handling of large photometric datasets
- **Progress feedback**: User notification during lengthy operations
- **Interruption capability**: Allow user cancellation of calculations

### 7.2 Accuracy Standards
- **Numerical precision**: Comply with AS/NZS 1158.2:2020 Section 4.5 rounding requirements
- **Calculation verification**: Results must be reproducible and verifiable
- **Edge case handling**: Proper behavior at parameter limits
- **Quality control**: Built-in checks for calculation integrity

## 8. Compliance and Documentation

### 8.1 Standards Compliance
- **AS/NZS 1158.2:2020**: Full compliance with computer procedure requirements
- **AS/NZS 1158.3.1:2020**: Alignment with Category P performance requirements
- **Calculation methodology**: Traceable to standard requirements
- **Verification capability**: Results must be independently verifiable

### 8.2 Documentation Requirements
- **User manual**: Comprehensive operating instructions
- **Technical specification**: Detailed calculation methodology
- **Validation report**: Proof of standards compliance
- **Example calculations**: Worked examples for verification

## 9. Future Considerations

### 9.1 Standard Updates
- **Version management**: Capability to handle standard revisions
- **Legacy support**: Maintain compatibility with previous standard versions
- **Update mechanisms**: Systematic approach to incorporating changes
- **Migration tools**: Assistance for transitioning between standard versions

### 9.2 Technology Evolution
- **Modern lighting technologies**: LED, adaptive lighting considerations
- **Smart lighting integration**: Support for dynamic lighting systems
- **Cloud connectivity**: Remote access and collaboration capabilities
- **Mobile compatibility**: Tablet/mobile device support where appropriate

## 10. Summary

This specification provides a comprehensive framework for developing software capable of generating AS/NZS 1158 compliant spacing charts and tables for Category P lighting applications. Implementation must ensure full compliance with the fundamental software requirements established in AS/NZS 1158.2:2020 while providing practical tools for lighting professionals to design safe, efficient, and compliant pedestrian area lighting systems.

The specification emphasizes accuracy, usability, and regulatory compliance while maintaining flexibility to accommodate the diverse range of Category P lighting applications covered by the Australian/New Zealand lighting standards.