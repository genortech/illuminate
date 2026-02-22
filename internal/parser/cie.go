package parser

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"illuminate/internal/database"
	"illuminate/internal/logger"
)

var cieHeaderRegex = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+(\d+)\s+(.+)$`)

type CIEParser struct{}

func NewCIEParser() *CIEParser {
	return &CIEParser{}
}

func (p *CIEParser) Parse(filepath string) (*database.ParsedLuminaire, error) {
	logger.Default.Debugf("parsing CIE file: %s", filepath)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	reader := io.TeeReader(file, hash)

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	metadata := database.Luminaire{
		OriginalFilename: filepath,
		FormatType:       "CIE",
	}

	var candelaLines []string
	firstLine := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if firstLine {
			if match := cieHeaderRegex.FindStringSubmatch(line); match != nil {
				metadata.SymmetryFlag, _ = strconv.Atoi(match[1])
				formatType, _ := strconv.Atoi(match[2])
				_ = formatType
				metadata.FormatType = "CIE"

				nameAndFlux := strings.TrimSpace(match[4])
				metadata.LuminaireDesc = nameAndFlux

				if idx := strings.LastIndex(nameAndFlux, " "); idx > 0 {
					fluxPart := strings.Trim(nameAndFlux[idx:], "lmsLMS")
					if flux, err := strconv.ParseFloat(strings.TrimSpace(fluxPart), 64); err == nil {
						metadata.LuminousFlux = flux
					}
				}

				if dashIdx := strings.LastIndex(nameAndFlux, "-"); dashIdx > 0 {
					metadata.Model = strings.TrimSpace(nameAndFlux[:dashIdx])
				}
			}
			firstLine = false
			continue
		}

		candelaLines = append(candelaLines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	var candelaMatrix [][]float64
	for _, line := range candelaLines {
		row := parseFloatLine(line)
		if len(row) > 0 {
			candelaMatrix = append(candelaMatrix, row)
		}
	}

	verticalAngles := make([]float64, len(candelaMatrix))
	for i := 0; i < len(candelaMatrix); i++ {
		verticalAngles[i] = float64(i * 10)
	}

	if len(candelaMatrix) > 0 {
		horizontalAngles := make([]float64, len(candelaMatrix[0]))
		for i := range horizontalAngles {
			horizontalAngles[i] = float64(i * 10)
		}
		metadata.Symmetry = metadata.SymmetryFlag
		_ = horizontalAngles
	}

	fileHash := fmt.Sprintf("%x", hash.Sum(nil))
	metadata.FileHash = fileHash

	logger.Default.Debugf("CIE parse complete: file_hash=%s, vertical_angles=%d, horizontal_angles=%d",
		fileHash, len(verticalAngles), len(candelaMatrix))

	return &database.ParsedLuminaire{
		Metadata:       metadata,
		VerticalAngles: verticalAngles,
		CandelaMatrix:  candelaMatrix,
	}, nil
}

func (p *CIEParser) Write(lum *database.ParsedLuminaire, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	symmetryFlag := lum.Metadata.SymmetryFlag
	if symmetryFlag == 0 {
		symmetryFlag = 1
	}

	name := lum.Metadata.LuminaireDesc
	if name == "" {
		name = lum.Metadata.Model
	}
	if name == "" {
		name = "Luminaire"
	}

	lumenStr := ""
	if lum.Metadata.LuminousFlux > 0 {
		lumenStr = fmt.Sprintf(" %.0f lms", lum.Metadata.LuminousFlux)
	}

	writer.WriteString(fmt.Sprintf("   %d   0   0        %s%s\n", symmetryFlag, name, lumenStr))

	for _, row := range lum.CandelaMatrix {
		for i, v := range row {
			if i > 0 {
				writer.WriteString(" ")
			}
			writer.WriteString(fmt.Sprintf("%3d", int(v)))
		}
		writer.WriteString("\n")
	}

	return nil
}
