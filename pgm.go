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
		Data:        data,
		Width:       width,
		Height:      height,
		MagicNumber: magicNumber,
		Max:         maxInt,
	}, nil
}

func (pgm *PGM) Size() (int, int) {
	return pgm.Width, pgm.Height
}

func (pgm *PGM) At(x, y int) uint8 {
	return pgm.Data[x][y]
}

func (pgm *PGM) Set(x, y int, value uint8) {
	pgm.Data[x][y] = value
}

func (pgm *PGM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Write PBM information to the file
	fmt.Fprintf(file, "%s\n", pgm.MagicNumber)
	fmt.Fprintf(file, "# saved file\n")
	fmt.Fprintf(file, "%d %d\n", pgm.Width, pgm.Height)
	fmt.Fprintf(file, "%d\n", pgm.Max)

	if pgm.MagicNumber == "P2" {
		for _, row := range pgm.Data {
			for _, pixel := range row {
				fmt.Fprintf(file, "%d ", pixel)
			}
			fmt.Fprintln(file)
		}
	}

	if pgm.MagicNumber == "P5" {
		for _, row := range pgm.Data {
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
	for i, row := range pgm.Data {
		for j, value := range row {
			pgm.Data[i][j] = uint8(pgm.Max) - value
		}
	}
}

func (pgm *PGM) Flip() {
	for x := 0; x < pgm.Height; x++ {
		for i, j := 0, pgm.Width-1; i < j; i, j = i+1, j-1 {
			pgm.Data[x][i], pgm.Data[x][j] = pgm.Data[x][j], pgm.Data[x][i]
		}
	}
}

func (pgm *PGM) Flop() {
	for y := 0; y < pgm.Width; y++ {
		for i, j := 0, pgm.Height-1; i < j; i, j = i+1, j-1 {
			pgm.Data[i][y], pgm.Data[j][y] = pgm.Data[j][y], pgm.Data[i][y]
		}
	}
}

func (pgm *PGM) SetMagicNumber(magicNumber string) {
	if magicNumber == pgm.MagicNumber {
		fmt.Printf("Magic Number already set to %s\n", pgm.MagicNumber)
	} else if magicNumber == "P2" && pgm.MagicNumber == "P5" {
		pgm.MagicNumber = "P2"
	} else if magicNumber == "P5" && pgm.MagicNumber == "P2" {
		pgm.MagicNumber = "P5"
	} else {
		fmt.Printf("Please select a valid magic number (P1 or P4) your curent file is set to %s\n", pgm.MagicNumber)
	}
}

func (pgm *PGM) SetMaxValue(maxValue uint8) {
	pgm.Max = int(maxValue)
}

func (pgm *PGM) Rotate90CW() {
	rotatedData := make([][]uint8, pgm.Width)
	for i := range rotatedData {
		rotatedData[i] = make([]uint8, pgm.Height)
	}

	for i := 0; i < pgm.Height; i++ {
		for j := 0; j < pgm.Width; j++ {
			rotatedData[j][pgm.Height-i-1] = pgm.Data[i][j]
		}
	}

	pgm.Data = rotatedData
	Width := pgm.Width
	pgm.Width = pgm.Height
	pgm.Height = Width
}

// func (pgm *PGM) ToPBM() *PBM{
//     // ...
// }