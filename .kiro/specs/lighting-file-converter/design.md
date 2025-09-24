# Design Document

## Overview

The lighting file converter follows a modular, hexagonal architecture pattern with clear separation of concerns between parsing, conversion logic, and output generation. The system is designed to handle three major photometric file formats (IES, LDT, CIE) with high accuracy and performance.

## Architecture

### High-Level Architecture

The system uses a pipeline architecture with the following flow:
```
Input File → Format Detection → Parser → Validation → 
Common Model → Transformation → Target Writer → Output File
```

### Core Components

1. **Format Detection**: Automatically identifies input file format using magic numbers and structure analysis
2. **Parser Layer**: Format-specific parsers that convert files to a common internal representation
3. **Common Data Model**: Unified photometric data structure that preserves all essential lighting parameters
4. **Validation Layer**: Cross-format validation ensuring data consistency and compliance
5. **Writer Layer**: Format-specific writers that generate output files from the common model
6. **Conversion Manager**: Orchestrates the entire conversion pipeline

### Project Structure

Based on go-blueprint's standard layout:

```
lighting-converter/
├── cmd/
│   ├── api/main.go              # REST API server entry point
│   └── cli/main.go              # CLI application entry point
├── internal/
│   ├── config/config.go         # Application configuration
│   ├── server/                  # HTTP server components
│   ├── converter/               # Conversion orchestration
│   ├── parsers/                 # Format-specific parsers
│   │   ├── ies/                 # IES format handling
│   │   ├── ldt/                 # LDT/EULUMDAT handling
│   │   └── cie/                 # CIE format handling
│   ├── writers/                 # Format-specific writers
│   ├── models/common.go         # Shared data models
│   └── utils/                   # Utilities and helpers
├── pkg/converter/               # Public API
└── test/                        # Test files and fixtures
```

## Components and Interfaces

### Parser Interface
```go
type Parser interface {
    Parse(data []byte) (*models.PhotometricData, error)
    Validate(data *models.PhotometricData) error
}
```

### Writer Interface
```go
type Writer interface {
    Write(data *models.PhotometricData) ([]byte, error)
    SetOptions(opts WriterOptions) error
}
```

### Conversion Manager
The conversion manager orchestrates the entire pipeline:
- Format detection and parser selection
- Data validation and transformation
- Writer selection and output generation
- Error handling and reporting

## Data Models

### Core Photometric Data Structure
```go
type PhotometricData struct {
    Metadata     LuminaireMetadata
    Geometry     LuminaireGeometry
    Photometry   PhotometricMeasurements
    Electrical   ElectricalData
}
```

### Format-Specific Models
Each format (IES, LDT, CIE) has its own data structures that map to and from the common model:
- IES models support LM-63-1995 and LM-63-2002 standards
- LDT models support EULUMDAT 1.0 format
- CIE models support CIE LTL format

## Error Handling

### Error Categories
1. **Syntax Errors**: Invalid file format, corrupted data
2. **Semantic Errors**: Invalid photometric values, incompatible parameters
3. **Conversion Errors**: Format limitations, data loss warnings
4. **System Errors**: File I/O, memory limitations

### Error Response Format
```go
type ConversionError struct {
    Code     string                 `json:"code"`
    Message  string                 `json:"message"`
    Context  map[string]interface{} `json:"context,omitempty"`
    Warnings []string               `json:"warnings,omitempty"`
}
```

## Testing Strategy

### Unit Testing
- Individual parser testing with format-specific test files
- Writer testing with round-trip validation
- Conversion manager logic testing
- Error handling scenario testing

### Integration Testing
- End-to-end conversion testing
- Round-trip conversion validation (IES→LDT→IES)
- API endpoint testing
- Batch processing validation

### Performance Testing
- Large file processing benchmarks
- Concurrent conversion testing
- Memory usage profiling
- Response time validation

### Test Data Strategy
- Curated collection of real-world photometric files
- Edge case and malformed file testing
- Format variation testing
- Cross-platform compatibility testing