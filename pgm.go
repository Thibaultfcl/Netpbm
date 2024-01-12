package Netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PGM struct {
	data          [][]uint8
	width, height int
	magicNumber   string
	max           int
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

	if magicNumber != "P2" && magicNumber != "P5" {
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
		var decimalValues []uint8
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
			for _, token := range tokens {
				decimalValue, err := strconv.ParseUint(token, 0, 8)
				if err != nil {
					return nil, fmt.Errorf("failed to convert in decimal: %v", err)
				}
				decimalValues = append(decimalValues, uint8(decimalValue))
			}

			for i, token := range decimalValues {
				if i >= width {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("invalid character in data: %v", token)
				}
				if int(token) <= maxInt {
					row[i] = token
				} else {
					return nil, fmt.Errorf("value out of range: %d", token)
				}
			}
		}
		data = append(data, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return &PGM{
		data:        data,
		width:       width,
		height:      height,
		magicNumber: magicNumber,
		max:         maxInt,
	}, nil
}

func (pgm *PGM) Size() (int, int) {
	return pgm.width, pgm.height
}

func (pgm *PGM) At(x, y int) uint8 {
	return pgm.data[x][y]
}

func (pgm *PGM) Set(x, y int, value uint8) {
	pgm.data[x][y] = value
}

func (pgm *PGM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Write PBM information to the file
	fmt.Fprintf(file, "%s\n", pgm.magicNumber)
	fmt.Fprintf(file, "# saved file\n")
	fmt.Fprintf(file, "%d %d\n", pgm.width, pgm.height)
	fmt.Fprintf(file, "%d\n", pgm.max)

	if pgm.magicNumber == "P2" {
		for _, row := range pgm.data {
			for _, pixel := range row {
				fmt.Fprintf(file, "%d ", pixel)
			}
			fmt.Fprintln(file)
		}
	}

	if pgm.magicNumber == "P5" {
		for _, row := range pgm.data {
			for _, value := range row {
				str := fmt.Sprintf("0x%02x", value)
				fmt.Fprintf(file, "%s ", str)
			}
			fmt.Fprintln(file)
		}
	}

	fmt.Printf("File created: %s\n", filename)
	return nil
}

func (pgm *PGM) Invert() {
	for i, row := range pgm.data {
		for j, value := range row {
			pgm.data[i][j] = uint8(pgm.max) - value
		}
	}
}

func (pgm *PGM) Flip() {
	for x := 0; x < pgm.height; x++ {
		for i, j := 0, pgm.width-1; i < j; i, j = i+1, j-1 {
			pgm.data[x][i], pgm.data[x][j] = pgm.data[x][j], pgm.data[x][i]
		}
	}
}

func (pgm *PGM) Flop() {
	for y := 0; y < pgm.width; y++ {
		for i, j := 0, pgm.height-1; i < j; i, j = i+1, j-1 {
			pgm.data[i][y], pgm.data[j][y] = pgm.data[j][y], pgm.data[i][y]
		}
	}
}

func (pgm *PGM) SetMagicNumber(magicNumber string) {
	if magicNumber == pgm.magicNumber {
		fmt.Printf("Magic Number already set to %s\n", pgm.magicNumber)
	} else if magicNumber == "P2" && pgm.magicNumber == "P5" {
		pgm.magicNumber = "P2"
	} else if magicNumber == "P5" && pgm.magicNumber == "P2" {
		pgm.magicNumber = "P5"
	} else {
		fmt.Printf("Please select a valid magic number (P1 or P4) your curent file is set to %s\n", pgm.magicNumber)
	}
}

func (pgm *PGM) SetMaxValue(maxValue uint8) {
	pgm.max = int(maxValue)
}

func (pgm *PGM) Rotate90CW() {
	rotatedData := make([][]uint8, pgm.width)
	for i := range rotatedData {
		rotatedData[i] = make([]uint8, pgm.height)
	}

	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			rotatedData[j][pgm.height-i-1] = pgm.data[i][j]
		}
	}

	pgm.data = rotatedData
	Width := pgm.width
	pgm.width = pgm.height
	pgm.height = Width
}

// func (pgm *PGM) ToPBM() *PBM{
//     // ...
// }