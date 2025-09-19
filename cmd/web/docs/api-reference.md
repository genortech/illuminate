# API Reference

This document provides a detailed reference for the REST API.

## Endpoints

### POST /api/v1/convert

Converts a lighting file from one format to another.

**Request Body:**

*   `input_format` (string, required): The format of the input file. Can be `ies`, `ldt`, or `cie`.
*   `output_format` (string, required): The format of the output file. Can be `ies`, `ldt`, or `cie`.
*   `file_content` (string, required): The content of the input file.

**Response Body:**

*   `file_content` (string): The content of the output file.

**Example:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "input_format": "ies",
  "output_format": "ldt",
  "file_content": "..."
}' http://localhost:8080/api/v1/convert
```

### POST /api/v1/validate

Validates a lighting file.

**Request Body:**

*   `format` (string, required): The format of the file. Can be `ies`, `ldt`, or `cie`.
*   `file_content` (string, required): The content of the file.

**Response Body:**

*   `valid` (boolean): Whether the file is valid.
*   `errors` (array): A list of errors, if any.

**Example:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "format": "ies",
  "file_content": "..."
}' http://localhost:8080/api/v1/validate
```
