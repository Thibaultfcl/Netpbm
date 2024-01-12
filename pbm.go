package Netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PBM struct {
	data          [][]bool
	width, height int
	magicNumber   string
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
		var binaryValue []string
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
					binaryValue = append(binaryValue, binaryDigits...)
				}
			}
			for _, value := range binaryValue {
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
		data = append(data, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return &PBM{
		data:        data,
		width:       width,
		height:      height,
		magicNumber: magicNumber,
	}, nil
}

func (pbm *PBM) Size() (int, int) {
	return pbm.width, pbm.height
}

func (pbm *PBM) At(x, y int) bool {
	return pbm.data[x][y]
}

func (pbm *PBM) Set(x, y int, value bool) {
	pbm.data[x][y] = value
}

func (pbm *PBM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	Width := pbm.width
	if pbm.magicNumber == "P4" {
		Width /= 8
		if Width <= 0 {
			Width = 1
		}
	}

	// Write PBM information to the file
	fmt.Fprintf(file, "%s\n", pbm.magicNumber)
	fmt.Fprintf(file, "# saved file\n")
	fmt.Fprintf(file, "%d %d\n", Width, pbm.height)
	if pbm.magicNumber == "P1" {
		for _, row := range pbm.data {
			for _, pixel := range row {
				if pixel {
					fmt.Fprint(file, "1")
				} else {
					fmt.Fprint(file, "0")
				}
				fmt.Fprint(file, " ")
			}
			fmt.Fprintln(file)
		}
	}

	if pbm.magicNumber == "P4" {
		for _, row := range pbm.data {
			var packedByte byte
			for i, pixel := range row {
				if pixel {
					packedByte |= 1 << (7 - i%8)
				}
				if i%8 == 7 || i == len(row)-1 {
					fmt.Fprintf(file, "0x%02X ", packedByte)
					packedByte = 0
				}
			}
			fmt.Fprintln(file)
		}
	}

	fmt.Printf("File created: %s\n", filename)
	return nil
}

func (pbm *PBM) Invert() {
	fmt.Println(pbm.height)
	fmt.Println(pbm.width)
	for x := 0; x < pbm.height; x++ {
		for y := 0; y < pbm.width; y++ {
			if pbm.data[x][y] {
				pbm.data[x][y] = false
			} else {
				pbm.data[x][y] = true
			}
		}
	}
}

func (pbm *PBM) Flip() {
	for x := 0; x < pbm.height; x++ {
		for i, j := 0, pbm.width-1; i < j; i, j = i+1, j-1 {
			pbm.data[x][i], pbm.data[x][j] = pbm.data[x][j], pbm.data[x][i]
		}
	}
}

func (pbm *PBM) Flop() {
	for y := 0; y < pbm.width; y++ {
		for i, j := 0, pbm.height-1; i < j; i, j = i+1, j-1 {
			pbm.data[i][y], pbm.data[j][y] = pbm.data[j][y], pbm.data[i][y]
		}
	}
}

func (pbm *PBM) SetMagicNumber(magicNumber string) {
	if magicNumber == pbm.magicNumber {
		fmt.Printf("Magic Number already set to %s\n", pbm.magicNumber)
	} else if magicNumber == "P4" && pbm.magicNumber == "P1" {
		pbm.magicNumber = "P4"
	} else if magicNumber == "P1" && pbm.magicNumber == "P4" {
		pbm.magicNumber = "P1"
	} else {
		fmt.Printf("Please select a valid magic number (P1 or P4) your curent file is set to %s\n", pbm.magicNumber)
	}
}
