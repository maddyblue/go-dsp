package main

import (
	"./fft"
	"fmt"
)

func main() {
	//x := [...]float64{0, 1, 0, 0, 0, 0, 0, 0}
	x := [...]float64{1, 2, 3, 4, 5, 6, 7, 8}

	nx := fft.Fft(x[:])

	fmt.Println("input:", len(x), x)
	fmt.Println("output:", len(nx), nx)
}
