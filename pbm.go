package Netpbm

import (
	"bufio"
	"fmt"
	"io"
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

	reader := bufio.NewReader(file)

	// Read magic number
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber)
	if magicNumber != "P1" && magicNumber != "P4" {
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

	// Read data
	data := make([][]bool, height)

	for i := range data {
		data[i] = make([]bool, width)
	}

	if magicNumber == "P1" {
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			for x, field := range fields {
				if x >= width {
					return nil, fmt.Errorf("index out of range at row %d", y)
				}
				data[y][x] = field == "1"
			}
		}

	} else if magicNumber == "P4" {
		expectedBytesPerRow := (width + 7) / 8
		for y := 0; y < height; y++ {
			row := make([]byte, expectedBytesPerRow)
			n, err := reader.Read(row)
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected end of file at row %d", y)
				}
				return nil, fmt.Errorf("error reading pixel data at row %d: %v", y, err)
			}
			if n < expectedBytesPerRow {
				return nil, fmt.Errorf("unexpected end of file at row %d, expected %d bytes, got %d", y, expectedBytesPerRow, n)
			}

			for x := 0; x < width; x++ {
				byteIndex := x / 8
				bitIndex := 7 - (x % 8)

				decimalValue := int(row[byteIndex])
				bitValue := (decimalValue >> bitIndex) & 1

				data[y][x] = bitValue != 0
			}
		}
	}

	return &PBM{data, width, height, magicNumber}, nil
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
