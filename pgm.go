package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type PGM struct {
	data          [][]uint8
	width, height int
	magicNumber   string
	max           uint8
}

func ReadPGM(filename string) (*PGM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read magic number
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber)
	if magicNumber != "P2" && magicNumber != "P5" {
		return nil, fmt.Errorf("invalid magic number: %s", magicNumber)
	}

	// Read dimensions
	dimensions, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading dimensions: %v", err)
	}
	var width, height int
	_, err = fmt.Sscanf(strings.TrimSpace(dimensions), "%d %d", &width, &height)
	if err != nil {
		return nil, fmt.Errorf("invalid dimensions: %v", err)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width and height must be positive")
	}

	// Read max value
	maxValue, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading max value: %v", err)
	}
	maxValue = strings.TrimSpace(maxValue)
	var max int
	_, err = fmt.Sscanf(maxValue, "%d", &max)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %v", err)
	}

	// Read data
	data := make([][]uint8, height)
	expectedBytesPerPixel := 1

	if magicNumber == "P2" {
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			rowData := make([]uint8, width)
			for x, field := range fields {
				if x >= width {
					return nil, fmt.Errorf("index out of range at row %d", y)
				}
				var pixelValue uint8
				_, err := fmt.Sscanf(field, "%d", &pixelValue)
				if err != nil {
					return nil, fmt.Errorf("error parsing pixel value at row %d, column %d: %v", y, x, err)
				}
				rowData[x] = pixelValue
			}
			data[y] = rowData
		}
	} else if magicNumber == "P5" {
		for y := 0; y < height; y++ {
			row := make([]byte, width*expectedBytesPerPixel)
			n, err := reader.Read(row)
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected end of file at row %d", y)
				}
				return nil, fmt.Errorf("error reading pixel data at row %d: %v", y, err)
			}
			if n < width*expectedBytesPerPixel {
				return nil, fmt.Errorf("unexpected end of file at row %d, expected %d bytes, got %d", y, width*expectedBytesPerPixel, n)
			}

			rowData := make([]uint8, width)
			for x := 0; x < width; x++ {
				pixelValue := uint8(row[x*expectedBytesPerPixel])
				rowData[x] = pixelValue
			}
			data[y] = rowData
		}
	}

	return &PGM{data, width, height, magicNumber, uint8(max)}, nil
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
	pgm.max = maxValue
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

func (pgm *PGM) ToPBM() *PBM {
	var dataResult [][]bool
	for x := 0; x < pgm.height; x++ {
		for y := 0; y < pgm.width; y++ {
			if pgm.data[x][y] >= (pgm.max / 2) {
				dataResult[x][y] = true
			} else {
				dataResult[x][y] = false
			}
		}
	}
	return &PBM{
		data: dataResult,
	}
}
