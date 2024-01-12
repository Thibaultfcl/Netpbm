package main

import (
	"fmt"

	netpbm "github.com/Thibaultfcl/Netpbm"
)

func main() {
	filename := "testP5.pgm"
	pgm, err := netpbm.ReadPGM(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Magic Number: %s\n", pgm.MagicNumber)
	fmt.Println("Width: ", pgm.Width)
	fmt.Println("Height: ", pgm.Height)
	fmt.Println("Max Int: ", pgm.Max)
	fmt.Println("Data:")
	for _, row := range pgm.Data {
		for _, pixel := range row {
			fmt.Print(pixel)
			fmt.Print(" ")
		}
		fmt.Println()
	}
}
