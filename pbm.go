package Netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
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

	widthTab := width
	if magicNumber == "P4" {
		for widthTab%8 != 0 {
			widthTab++
		}
	}

	// Read data
	var data [][]bool
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Fields(line)
		row := make([]bool, widthTab)

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
			data = append(data, row)
		}
		if magicNumber == "P4" {
			i := 0
			for _, value := range line {
				asciiNb, err := ConvertToASCII(string(value))
				if err != nil {
					return nil, fmt.Errorf("failed to converse to Ascii: %v", err)
				}
				binaryStr := fmt.Sprintf("%08b", asciiNb[:])
				binaryStr = strings.TrimPrefix(binaryStr, "[")
				binaryStr = strings.TrimSuffix(binaryStr, "]")

				if i%2 == 0 {
					row = make([]bool, 0)
				}

				for _, token := range binaryStr {
					tokenStr := string(token)

					if tokenStr == "1" {
						row = append(row, true)
					} else if tokenStr == "0" {
						row = append(row, false)
					} else {
						return nil, fmt.Errorf("invalid character in data: %v", token)
					}
				}

				i++
				if i%2 == 0 {
					data = append(data, row)
				}
			}
		}
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

func ConvertToASCII(input string) ([]int, error) {
	var result []int

	encoder := charmap.Windows1252.NewEncoder()

	for _, char := range input {
		encodedChar, err := encoder.String(string(char))
		if err != nil {
			return nil, err
		}

		if len(encodedChar) == 1 {
			result = append(result, int(encodedChar[0]))
		}
	}

	return result, nil
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

	// Write PBM information to the file
	fmt.Fprintf(file, "%s\n", pbm.magicNumber)
	fmt.Fprintf(file, "# saved file\n")
	fmt.Fprintf(file, "%d %d\n", pbm.width, pbm.height)
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
		var intArray []int
		for _, row := range pbm.data {
			for _, value := range row {
				if value {
					intArray = append(intArray, 1)
				} else {
					intArray = append(intArray, 0)
				}
			}
		}

		var byteArrays [][]int
		for i := 0; i < len(intArray); i += 8 {
			byteArrays = append(byteArrays, intArray[i:i+8])
		}

		for _, byteArray := range byteArrays {
			result, err := BinaryToWindows1252(byteArray)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Fprint(file, result)
			}
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

func BinaryToWindows1252(binaryArray []int) (string, error) {
	if len(binaryArray)%8 != 0 {
		return "", fmt.Errorf("the length of the binary array must be a multiple of 8")
	}

	var byteArray []byte
	for i := 0; i < len(binaryArray); i += 8 {
		byteValue, err := BinaryToDecimal(binaryArray[i : i+8])
		if err != nil {
			fmt.Println(err)
		}
		byteArray = append(byteArray, byte(byteValue))
	}

	return string(byteArray), nil
}

func BinaryToDecimal(binary []int) (int, error) {
	binaryStr := ""
	for _, bit := range binary {
		binaryStr += strconv.Itoa(bit)
	}

	decimal, err := strconv.ParseInt(binaryStr, 2, 64)
	if err != nil {
		return 0, err
	}

	return int(decimal), nil
}
