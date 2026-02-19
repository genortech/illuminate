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

var keywordRegex = regexp.MustCompile(`^\[(\w+)\]\s*(.*)$`)

type IESParser struct{}

func NewIESParser() *IESParser {
	return &IESParser{}
}

func (p *IESParser) Parse(filepath string) (*database.ParsedLuminaire, error) {
	logger.Default.Debugf("parsing IES file: %s", filepath)

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
	}

	keywords := make(map[string]string)
	var tiltLine string
	var mainDataLine string
	var verticalAnglesLine string
	var horizontalAnglesLine string
	var candelaLines []string

	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if match := keywordRegex.FindStringSubmatch(line); match != nil {
				keywords[strings.ToUpper(match[1])] = strings.TrimSpace(match[2])
			}
			continue
		}

		if strings.HasPrefix(line, "TILT=") {
			tiltLine = line
			continue
		}

		if tiltLine != "" && mainDataLine == "" {
			mainDataLine = line
			continue
		}

		if mainDataLine != "" && verticalAnglesLine == "" {
			verticalAnglesLine = line
			continue
		}

		if verticalAnglesLine != "" && horizontalAnglesLine == "" {
			horizontalAnglesLine = line
			continue
		}

		if horizontalAnglesLine != "" {
			candelaLines = append(candelaLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	metadata.TestNumber = keywords["TEST"]
	metadata.TestLab = keywords["TESTLAB"]
	metadata.Manufacturer = keywords["MANUFAC"]
	metadata.IssueDate = keywords["ISSUEDATE"]
	metadata.Model = keywords["LUMCAT"]
	metadata.LuminaireDesc = keywords["LUMINAIRE"]
	metadata.LampCatalog = keywords["LAMPCAT"]
	metadata.LampType = keywords["LAMP"]
	metadata.Ballast = keywords["BALLAST"]
	metadata.LampPosition = keywords["LAMPPOSITION"]

	mainData := strings.Fields(mainDataLine)
	if len(mainData) >= 10 {
		if n, err := strconv.Atoi(mainData[3]); err == nil {
			metadata.PhotometricType = database.PhotometricType(n)
		}
		if n, err := strconv.Atoi(mainData[7]); err == nil {
			metadata.InputWatts = float64(n)
		}
	}

	verticalAngles := parseFloatLine(verticalAnglesLine)
	horizontalAngles := parseFloatLine(horizontalAnglesLine)

	var candelaMatrix [][]float64
	for _, line := range candelaLines {
		row := parseFloatLine(line)
		if len(row) > 0 {
			candelaMatrix = append(candelaMatrix, row)
		}
	}

	fileHash := fmt.Sprintf("%x", hash.Sum(nil))
	metadata.FileHash = fileHash

	logger.Default.Debugf("IES parse complete: file_hash=%s, vertical_angles=%d, horizontal_angles=%d",
		fileHash, len(verticalAngles), len(horizontalAngles))

	return &database.ParsedLuminaire{
		Metadata:         metadata,
		VerticalAngles:   verticalAngles,
		HorizontalAngles: horizontalAngles,
		CandelaMatrix:    candelaMatrix,
	}, nil
}

func parseFloatLine(line string) []float64 {
	fields := strings.Fields(line)
	result := make([]float64, 0, len(fields))
	for _, f := range fields {
		if v, err := strconv.ParseFloat(f, 64); err == nil {
			result = append(result, v)
		}
	}
	return result
}

func (p *IESParser) Write(lum *database.ParsedLuminaire, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	writer.WriteString("IESNA:LM-63-2002\n")

	if lum.Metadata.TestNumber != "" {
		writer.WriteString(fmt.Sprintf("[TEST] %s\n", lum.Metadata.TestNumber))
	}
	if lum.Metadata.TestLab != "" {
		writer.WriteString(fmt.Sprintf("[TESTLAB] %s\n", lum.Metadata.TestLab))
	}
	if lum.Metadata.Manufacturer != "" {
		writer.WriteString(fmt.Sprintf("[MANUFAC] %s\n", lum.Metadata.Manufacturer))
	}
	if lum.Metadata.IssueDate != "" {
		writer.WriteString(fmt.Sprintf("[ISSUEDATE] %s\n", lum.Metadata.IssueDate))
	}
	if lum.Metadata.Model != "" {
		writer.WriteString(fmt.Sprintf("[LUMCAT] %s\n", lum.Metadata.Model))
	}
	if lum.Metadata.LuminaireDesc != "" {
		writer.WriteString(fmt.Sprintf("[LUMINAIRE] %s\n", lum.Metadata.LuminaireDesc))
	}
	if lum.Metadata.LampCatalog != "" {
		writer.WriteString(fmt.Sprintf("[LAMPCAT] %s\n", lum.Metadata.LampCatalog))
	}
	if lum.Metadata.LampType != "" {
		writer.WriteString(fmt.Sprintf("[LAMP] %s\n", lum.Metadata.LampType))
	}
	if lum.Metadata.Ballast != "" {
		writer.WriteString(fmt.Sprintf("[BALLAST] %s\n", lum.Metadata.Ballast))
	}
	if lum.Metadata.LampPosition != "" {
		writer.WriteString(fmt.Sprintf("[LAMPPOSITION] %s\n", lum.Metadata.LampPosition))
	}

	writer.WriteString("TILT=NONE\n")

	numVert := len(lum.VerticalAngles)
	numHorz := len(lum.HorizontalAngles)
	photometricType := int(lum.Metadata.PhotometricType)
	if photometricType == 0 {
		photometricType = 1
	}

	writer.WriteString(fmt.Sprintf("1 -1 1 %d %d 1 2 0.2 0.2 %.0f\n",
		numVert, numHorz, lum.Metadata.InputWatts))

	writer.WriteString(fmt.Sprintf("1 1 %.2f\n", lum.Metadata.InputWatts))

	writer.WriteString(floatSliceToString(lum.VerticalAngles))
	writer.WriteString("\n")

	writer.WriteString(floatSliceToString(lum.HorizontalAngles))
	writer.WriteString("\n")

	for _, row := range lum.CandelaMatrix {
		writer.WriteString(floatSliceToString(row))
		writer.WriteString("\n")
	}

	logger.Default.Infof("Wrote IES file to %s", filepath)
	return nil
}

func floatSliceToString(vals []float64) string {
	var sb strings.Builder
	for i, v := range vals {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%.1f", v))
	}
	return sb.String()
}
