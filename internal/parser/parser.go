package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"illuminate/internal/database"
)

type Parser interface {
	Parse(filepath string) (*database.ParsedLuminaire, error)
	Write(lum *database.ParsedLuminaire, filepath string) error
}

func GetParser(filename string) (Parser, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".ies":
		return NewIESParser(), nil
	case ".cie":
		return NewCIEParser(), nil
	case ".ldt":
		return NewLDTParser(), nil
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func GetSupportedExtensions() []string {
	return []string{".ies", ".cie", ".ldt"}
}

func DetectFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".ies":
		return "IES (IESNA LM-63)"
	case ".cie":
		return "CIE (CIE 102)"
	case ".ldt":
		return "LDT (Eulumdat)"
	default:
		return "Unknown"
	}
}
