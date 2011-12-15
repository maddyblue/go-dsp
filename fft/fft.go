/*
 * Copyright (c) 2011 Matt Jibson <matt.jibson@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

// Package fft provides forward and inverse fast Fourier transform functions.
package fft

import (
	"math"
	"os"
	"sync"
)

var factors_lock sync.RWMutex

// radix-2 factors
var factors = map[int][]complex128{}

// bluestein factors
var n2_factors = map[int][]complex128{}
var n2_inv_factors = map[int][]complex128{}

// Ensures the complex multiplication factors exist for an input array of length input_len.
func ensureFactors(input_len int) {
	var cos, sin float64

	factors_lock.Lock()

	for i := 4; i <= input_len; i <<= 1 {
		if factors[i] == nil {
			factors[i] = make([]complex128, i)
			for n := 0; n < i; n++ {
				if n == 0 {
					sin, cos = 0, 1
				} else if n*4 == i {
					sin, cos = -1, 0
				} else {
					sin, cos = math.Sincos(-2 * math.Pi / float64(i) * float64(n))
				}
				factors[i][n] = complex(cos, sin)
			}
		}
	}

	if n2_factors[input_len] == nil {
		n2_factors[input_len] = make([]complex128, input_len)
		n2_inv_factors[input_len] = make([]complex128, input_len)

		for i := 0; i < input_len; i++ {
			if i == 0 {
				sin, cos = 0, 1
			} else {
				sin, cos = math.Sincos(math.Pi / float64(input_len) * float64(i*i))
			}
			n2_factors[input_len][i] = complex(cos, sin)
			n2_inv_factors[input_len][i] = complex(cos, -sin)
		}
	}

	factors_lock.Unlock()
}

// FFTReal returns the forward FFT of the real-valued slice.
func FFTReal(x []float64) []complex128 {
	return FFT(toComplex(x))
}

// IFFTReal returns the inverse FFT of the real-valued slice.
func IFFTReal(x []float64) []complex128 {
	return IFFT(toComplex(x))
}

// toComplex returns the complex equivalent of the real-valued slice.
func toComplex(x []float64) []complex128 {
	y := make([]complex128, len(x))
	for n, v := range x {
		y[n] = complex(v, 0)
	}
	return y
}

// IFFT returns the inverse FFT of the complex-valued slice.
func IFFT(x []complex128) []complex128 {
	lx := len(x)
	r := make([]complex128, lx)

	// Reverse inputs, which is calculated with modulo N, hence x[0] as an outlier
	r[0] = x[0]
	for i := 1; i < lx; i++ {
		r[i] = x[lx-i]
	}

	r = FFT(r)

	N := complex(float64(lx), 0)
	for n, _ := range r {
		r[n] /= N
	}
	return r
}

// Convolve returns the convolution of x * y.
func Convolve(x, y []complex128) ([]complex128, os.Error) {
	if len(x) != len(y) {
		return []complex128{}, os.NewError("fft: input arrays are not of equal length")
	}

	fft_x := FFT(x)
	fft_y := FFT(y)

	r := make([]complex128, len(x))
	for i := 0; i < len(r); i++ {
		r[i] = fft_x[i] * fft_y[i]
	}

	return IFFT(r), nil
}

// FFT returns the forward FFT of the complex-valued slice.
func FFT(x []complex128) []complex128 {
	lx := len(x)

	// todo: non-hack handling length <= 1 cases
	if lx <= 1 {
		r := make([]complex128, lx)
		copy(r, x)
		return r
	}

	if isPowerOf2(lx) {
		return radix2FFT(x)
	}

	return bluesteinFFT(x)
}

// radix2FFT returns the FFT calculated using the radix-2 DIT Cooley-Tukey algorithm.
func radix2FFT(x []complex128) []complex128 {
	lx := len(x)
	ensureFactors(lx)

	lx_2 := lx / 2
	r := make([]complex128, lx) // result
	t := make([]complex128, lx) // temp
	copy(r, x)

	// split into even and odd parts for each stage
	for block_sz := lx; block_sz > 1; block_sz >>= 1 {
		i := 0
		bs_2 := block_sz / 2
		for block := 0; block < lx/block_sz; block++ {
			for n := 0; n < bs_2; n++ {
				bn := block_sz*block + n
				t[bn] = r[i]
				i++
				t[bn+bs_2] = r[i]
				i++
			}
		}
		copy(r, t)
	}

	for stage := 2; stage <= lx; stage <<= 1 {
		if stage == 2 { // 2-point transforms
			for n := 0; n < lx_2; n++ {
				t[n*2] = r[n*2] + r[n*2+1]
				t[n*2+1] = r[n*2] - r[n*2+1]
			}
		} else { // >2-point transforms
			blocks := lx / stage
			s_2 := stage / 2

			for n := 0; n < blocks; n++ {
				nb := n * stage
				for j := 0; j < s_2; j++ {
					w_n := r[j+nb+s_2] * factors[stage][j]
					t[j+nb] = r[j+nb] + w_n
					t[j+nb+s_2] = r[j+nb] - w_n
				}
			}
		}

		copy(r, t)
	}

	return r
}

// bluesteinFFT returns the FFT calculated using the Bluestein algorithm.
func bluesteinFFT(x []complex128) []complex128 {
	lx := len(x)
	a := zeroPad(x, nextPowerOf2(lx*2-1))
	la := len(a)
	ensureFactors(lx)

	for n, v := range x {
		a[n] = v * n2_inv_factors[lx][n]
	}

	b := make([]complex128, la)
	for i := 0; i < lx; i++ {
		b[i] = n2_factors[lx][i]

		if i != 0 {
			b[la-i] = n2_factors[lx][i]
		}
	}

	r, _ := Convolve(a, b)

	for i := 0; i < lx; i++ {
		r[i] *= n2_inv_factors[lx][i]
	}

	return r[:lx]
}

// isPowerOf2 returns true if x is a power of 2, else false.
func isPowerOf2(x int) bool {
	return x&(x-1) == 0
}

// nextPowerOf2 returns the next power of 2 >= x.
func nextPowerOf2(x int) int {
	if isPowerOf2(x) {
		return x
	}

	return int(math.Pow(2, math.Ceil(math.Log2(float64(x)))))
}

// zeroPad returns x with zeros appended to the end to the specified length.
// If len(x) == length, x is returned.
func zeroPad(x []complex128, length int) []complex128 {
	if len(x) == length {
		return x
	}

	r := make([]complex128, length)
	copy(r, x)
	return r
}

// zeroPad2 returns zeroPad of x, with the length as next power of 2 >= len(x).
func zeroPad2(x []complex128) []complex128 {
	return zeroPad(x, nextPowerOf2(len(x)))
}

// toComplex2 returns the complex equivalent of the real-valued matrix.
func toComplex2(x [][]float64) [][]complex128 {
	y := make([][]complex128, len(x))
	for n, v := range x {
		y[n] = toComplex(v)
	}
	return y
}

// FFT2Real returns the 2-dimensional, forward FFT of the real-valued matrix.
func FFT2Real(x [][]float64) ([][]complex128, os.Error) {
	return FFT2(toComplex2(x))
}

// FFT2 returns the 2-dimensional, forward FFT of the complex-valued matrix.
func FFT2(x [][]complex128) ([][]complex128, os.Error) {
	return computeFFT2(x, FFT)
}

// IFFT2Real returns the 2-dimensional, inverse FFT of the real-valued matrix.
func IFFT2Real(x [][]float64) ([][]complex128, os.Error) {
	return IFFT2(toComplex2(x))
}

// IFFT2 returns the 2-dimensional, inverse FFT of the complex-valued matrix.
func IFFT2(x [][]complex128) ([][]complex128, os.Error) {
	return computeFFT2(x, IFFT)
}

func computeFFT2(x [][]complex128, fftFunc func([]complex128) []complex128) ([][]complex128, os.Error) {
	rows := len(x)
	if rows == 0 {
		return nil, os.NewError("fft: empty input array")
	}

	cols := len(x[0])
	r := make([][]complex128, rows)
	for i := 0; i < rows; i++ {
		if len(x[i]) != cols {
			return nil, os.NewError("fft: input matrix must have identical row lengths")
		}
		r[i] = make([]complex128, cols)
	}

	for i := 0; i < cols; i++ {
		t := make([]complex128, rows)
		for j := 0; j < rows; j++ {
			t[j] = x[j][i]
		}

		for n, v := range fftFunc(t) {
			r[n][i] = v
		}
	}

	for n, v := range r {
		r[n] = fftFunc(v)
	}

	return r, nil
}
