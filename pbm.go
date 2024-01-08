package netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PBM struct {
	Data          [][]bool
	Width, Height int
	MagicNumber   string
}

func ReadPBM(filename string) (*PBM, error) {
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

	width, err := strconv.Atoi(dimensions[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse width: %v", err)
	}

	height, err := strconv.Atoi(dimensions[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse height: %v", err)
	}

	// Read data
	var data [][]bool
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		row := make([]bool, width)
		if magicNumber == "P1" {
			for i, token := range tokens {
				if i >= width {
					break
				}
				if token == "1" {
					row[i] = true
				} else if token == "0" {
					row[i] = false
				} else {
					return nil, fmt.Errorf("invalid character in data: %s", token)
				}
			}
		}
		if magicNumber == "P4" {
			for _, hexValue := range tokens {
				val, _ := strconv.ParseInt(hexValue, 16, 8)
				binaryValue := fmt.Sprintf("%08b", val)
				for _, bit := range binaryValue {
					if bit == '1' {
						row = append(row, true)
					} else {
						row = append(row, false)
					}
				}
			}
		}
		data = append(data, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return &PBM{
		Data:        data,
		Width:       width,
		Height:      height,
		MagicNumber: magicNumber,
	}, nil
}

func HexToBinary(hexString string) (string, error) {
	hexValue, err := strconv.ParseInt(hexString, 16, 8)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08b", hexValue), nil
}

func (pbm *PBM) Size() (int, int) {
	return pbm.Width, pbm.Height
}

func (pbm *PBM) At(x, y int) bool {
	return pbm.Data[x][y]
}

func (pbm *PBM) Set(x, y int, value bool) {
	pbm.Data[x][y] = value
}

func (pbm *PBM) Save(filename string) error {
	fileName := "save.pbm"
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Write PBM information to the file
	fmt.Fprintf(file, "%s\n", pbm.MagicNumber)
	fmt.Fprintf(file, "# saved file\n")
	fmt.Fprintf(file, "%d %d\n", pbm.Width, pbm.Height)
	for _, row := range pbm.Data {
		for _, pixel := range row {
			if pixel {
				fmt.Fprint(file, "1")
			} else {
				fmt.Fprint(file, "0")
			}
		}
		fmt.Fprintln(file)
	}

	fmt.Printf("File created: %s\n", fileName)
	return nil
}
