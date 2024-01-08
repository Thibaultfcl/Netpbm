package main

import (
	"fmt"

	netpbm "github.com/Thibaultfcl/Netpbm"
)

func main() {
	filename := "testP4.pbm"
	pbm, err := netpbm.ReadPBM(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	Width, Height := pbm.Size()
	pbm.Set(0, 0, true)

	fmt.Printf("Magic Number: %s\n", pbm.MagicNumber)
	fmt.Println("Width: ", Width)
	fmt.Println("Height: ", Height)
	fmt.Println("Data:")
	for _, row := range pbm.Data {
		for _, pixel := range row {
			if pixel {
				fmt.Print("■")
			} else {
				fmt.Print("□")
			}
		}
		fmt.Println()
	}
}
