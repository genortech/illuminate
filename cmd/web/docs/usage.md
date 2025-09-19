# Usage

This document describes how to use the illuminate application.

## REST API

The REST API provides endpoints for converting and validating lighting files.

### Convert File

**Endpoint:** `POST /api/v1/convert`

**Request:**

```json
{
  "input_format": "ies",
  "output_format": "ldt",
  "file_content": "..."
}
```

**Response:**

```json
{
  "file_content": "..."
}
```

### Validate File

**Endpoint:** `POST /api/v1/validate`

**Request:**

```json
{
  "format": "ies",
  "file_content": "..."
}
```

**Response:**

```json
{
  "valid": true,
  "errors": []
}
```

## Command-Line Interface

The command-line interface provides commands for converting and validating lighting files.

### Convert File

```bash
illuminate convert --input <input_file> --output <output_file>
```

### Validate File

```bash
illuminate validate --file <file>
```
