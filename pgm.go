package netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PGM struct {
	Data          [][]uint8
	Width, Height int
	MagicNumber   string
	Max           int
}

func ReadPGM(filename string) (*PGM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Read magic number
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read magic number")
	}
	magicNumber := scanner.Text()

	if magicNumber != "P1" && magicNumber != "P4" {
		return nil, fmt.Errorf("unsupported PBM format: %s", magicNumber)
	}

	// Skip comments and empty lines
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 && line[0] != '#' {
			break
		}
	}

	// Read width and height
	if scanner.Err() != nil {
		return nil, fmt.Errorf("error reading dimensions line: %v", scanner.Err())
	}
	dimensions := strings.Fields(scanner.Text())
	if len(dimensions) != 2 {
		return nil, fmt.Errorf("invalid dimensions line")
	}

	width, err := strconv.Atoi(dimensions[0]) // largeur
	if err != nil {
		return nil, fmt.Errorf("failed to parse width: %v", err)
	}

	height, err := strconv.Atoi(dimensions[1]) // hauteur
	if err != nil {
		return nil, fmt.Errorf("failed to parse height: %v", err)
	}

	if magicNumber == "P5" {
		width *= 8
	}

	// Read maximum pixel value
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read max value")
	}
	maxInt, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return nil, fmt.Errorf("failed to parse max value: %v", err)
	}

	// Read data
	var data [][]uint8

	for scanner.Scan() {
		var binaryValue []string
		line := scanner.Text()
		tokens := strings.Fields(line)
		row := make([]uint8, width)

		if magicNumber == "P2" {
			for i, token := range tokens {
				if i >= width {
					break
				}
				value, err := strconv.Atoi(token)
				if err != nil {
					return nil, fmt.Errorf("invalid character in data: %s", token)
				}

				if value >= 0 && value <= maxInt {
					row[i] = uint8(value)
				} else {
					return nil, fmt.Errorf("value out of range: %d", value)
				}
			}
		}
		if magicNumber == "P5" {
			i := 0
			for _, token := range tokens {
				token = strings.TrimPrefix(token, "0x")
				for _, digit := range token {
					digitValue, err := strconv.ParseUint(string(digit), 16, 4)
					if err != nil {
						return nil, err
					}
					binaryDigits := strings.Split(fmt.Sprintf("%04b", digitValue), "")
					binaryValue = append(binaryValue, binaryDigits...)
				}
			}
			for _, token := range binaryValue {
				if i >= width {
					break
				}
				value, err := strconv.Atoi(token)
				if err != nil {
					return nil, fmt.Errorf("invalid character in data: %s", token)
				}

				if value >= 0 && value <= maxInt {
					row[i] = uint8(value)
				} else {
					return nil, fmt.Errorf("value out of range: %d", value)
				}
			}
		}
		data = append(data, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return &PGM{
		Data:        data,
		Width:       width,
		Height:      height,
		MagicNumber: magicNumber,
		Max:         maxInt,
	}, nil
}
