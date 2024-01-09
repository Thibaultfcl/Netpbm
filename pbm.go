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

	width, err := strconv.Atoi(dimensions[0]) // largeur
	if err != nil {
		return nil, fmt.Errorf("failed to parse width: %v", err)
	}

	height, err := strconv.Atoi(dimensions[1]) // hauteur
	if err != nil {
		return nil, fmt.Errorf("failed to parse height: %v", err)
	}

	if magicNumber == "P4" {
		width *= 8
	}

	// Read data
	var data [][]bool
	for scanner.Scan() {
		var binaryBits []string
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
			i := 0
			for _, token := range tokens {
				token = strings.TrimPrefix(token, "0x")
				for _, digit := range token {
					digitValue, err := strconv.ParseUint(string(digit), 16, 4)
					if err != nil {
						return nil, err
					}
					binaryDigits := strings.Split(fmt.Sprintf("%04b", digitValue), "")
					binaryBits = append(binaryBits, binaryDigits...)
				}

				if i >= width {
					break
				}
				for _, value := range binaryBits {
					if value == "1" {
						row[i] = true
						i++
					} else if value == "0" {
						row[i] = false
						i++
					} else {
						return nil, fmt.Errorf("invalid character in data: %v", value)
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
