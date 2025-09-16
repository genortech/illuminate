# System Design: Go Lighting File Converter

## Architecture Overview
The lighting file converter follows a modular, hexagonal architecture pattern with clear separation of concerns between parsing, conversion logic, and output generation.

## Project Structure (Using go-blueprint)

Based on go-blueprint's standard layout for a server project:

```
lighting-converter/
├── .github/
│   └── workflows/
│       ├── go-test.yml
│       └── release.yml
├── cmd/
│   ├── api/
│   │   └── main.go              # REST API server entry point
│   └── cli/
│       └── main.go              # CLI application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Application configuration
│   ├── server/
│   │   ├── routes.go            # HTTP route definitions
│   │   ├── handlers.go          # HTTP request handlers
│   │   └── middleware.go        # Authentication, logging middleware
│   ├── converter/
│   │   ├── interfaces.go        # Parser and writer interfaces
│   │   ├── manager.go           # Conversion orchestration
│   │   └── validation.go        # Cross-format validation
│   ├── parsers/
│   │   ├── ies/
│   │   │   ├── parser.go        # IES file parsing logic
│   │   │   ├── models.go        # IES data structures
│   │   │   └── validator.go     # IES-specific validation
│   │   ├── ldt/
│   │   │   ├── parser.go        # LDT/EULUMDAT parsing
│   │   │   ├── models.go        # LDT data structures
│   │   │   └── validator.go     # LDT-specific validation
│   │   └── cie/
│   │       ├── parser.go        # CIE file parsing
│   │       ├── models.go        # CIE data structures
│   │       └── validator.go     # CIE-specific validation
│   ├── writers/
│   │   ├── ies/
│   │   │   └── writer.go        # IES file generation
│   │   ├── ldt/
│   │   │   └── writer.go        # LDT file generation
│   │   └── cie/
│   │       └── writer.go        # CIE file generation
│   ├── models/
│   │   └── common.go            # Shared data models
│   └── utils/
│       ├── logger.go            # Structured logging
│       ├── errors.go            # Custom error types
│       └── fileutils.go         # File handling utilities
├── pkg/
│   └── converter/
│       ├── converter.go         # Public API for library usage
│       └── types.go             # Public type definitions
├── test/
│   ├── fixtures/
│   │   ├── sample.ies           # Test IES files
│   │   ├── sample.ldt           # Test LDT files
│   │   └── sample.cie           # Test CIE files
│   ├── integration/
│   │   └── converter_test.go    # End-to-end tests
│   └── unit/
│       ├── parser_test.go       # Parser unit tests
│       └── writer_test.go       # Writer unit tests
├── .air.toml                    # Hot reload configuration
├── .env                         # Environment variables
├── .gitignore
├── docker-compose.yml           # Development environment
├── Dockerfile                   # Container configuration
├── go.mod
├── go.sum
├── Makefile                     # Build and development commands
└── README.md
```

## Core Components

### 1. Parser Interface
```go
type Parser interface {
    Parse(data []byte) (*models.PhotometricData, error)
    Validate(data *models.PhotometricData) error
}
```

### 2. Writer Interface
```go
type Writer interface {
    Write(data *models.PhotometricData) ([]byte, error)
    SetOptions(opts WriterOptions) error
}
```

### 3. Conversion Manager
- Orchestrates the parsing → validation → conversion → writing pipeline
- Handles format detection and routing to appropriate parsers/writers
- Manages conversion context and metadata

## Data Flow Architecture

```
Input File → Format Detection → Parser → Validation → 
Common Model → Transformation → Target Writer → Output File
```

### 1. Input Processing
- File type detection (magic numbers, structure analysis)
- Format-specific parser selection
- Initial syntax validation

### 2. Intermediate Representation
- Common photometric data model
- Preserves all essential lighting parameters
- Handles different coordinate systems and units

### 3. Output Generation
- Target format writer selection
- Data transformation and serialization
- Final validation and compliance checking

## API Design

### CLI Interface
```bash
# Basic conversion
lighting-converter convert --input sample.ies --output sample.ldt

# Batch processing
lighting-converter batch --input-dir ./input --output-dir ./output --format ldt

# Validation only
lighting-converter validate --file sample.ies
```

### REST API Endpoints
```
POST   /api/v1/convert      # Single file conversion
POST   /api/v1/batch        # Batch conversion
GET    /api/v1/validate     # File validation
GET    /api/v1/formats      # Supported formats info
```

## Error Handling Strategy

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

## Performance Considerations

### Memory Management
- Streaming parsers for large files
- Buffer pooling for concurrent operations
- Garbage collection optimization

### Concurrency
- Worker pool pattern for batch processing
- Context-aware cancellation
- Rate limiting for API endpoints

## Extensibility

### Adding New Formats
1. Implement Parser interface
2. Implement Writer interface
3. Register with conversion manager
4. Add validation rules
5. Update API documentation

### Plugin Architecture (Future)
- Dynamic format loading
- Custom validation rules
- Third-party parser integration