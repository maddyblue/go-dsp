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
	"dsputils"
	"math"
	"sync"
)

var (
	radix2Lock sync.RWMutex
	radix2Factors = map[int][]complex128{
		4: {complex(1, 0), complex(0, -1), complex(-1, 0), complex(0, 1)},
	}
)

var (
	bluesteinLock sync.RWMutex
	bluesteinFactors = map[int][]complex128{}
	bluesteinInvFactors = map[int][]complex128{}
)

func getRadix2Factors(input_len int) []complex128 {
	radix2Lock.RLock()

	if hasRadix2Factors(input_len) {
		defer radix2Lock.RUnlock()
		return radix2Factors[input_len]
	}

	radix2Lock.RUnlock()
	radix2Lock.Lock()
	defer radix2Lock.Unlock()

	if !hasRadix2Factors(input_len) {
		for i, p := 8, 4; i <= input_len; i, p = i<<1, i {
			if radix2Factors[i] == nil {
				radix2Factors[i] = make([]complex128, i)

				for n, j := 0, 0; n < i; n, j = n+2, j+1 {
					radix2Factors[i][n] = radix2Factors[p][j]
				}

				for n := 1; n < i; n += 2 {
					sin, cos := math.Sincos(-2 * math.Pi / float64(i) * float64(n))
					radix2Factors[i][n] = complex(cos, sin)
				}
			}
		}
	}

	return radix2Factors[input_len]
}

func hasRadix2Factors(idx int) bool {
	return radix2Factors[idx] != nil
}

func getBluesteinFactors(input_len int) ([]complex128, []complex128) {
	bluesteinLock.RLock()

	if hasBluesteinFactors(input_len) {
		defer bluesteinLock.RUnlock()
		return bluesteinFactors[input_len], bluesteinInvFactors[input_len]
	}

	bluesteinLock.RUnlock()
	bluesteinLock.Lock()
	defer bluesteinLock.Unlock()

	if !hasBluesteinFactors(input_len) {
		bluesteinFactors[input_len] = make([]complex128, input_len)
		bluesteinInvFactors[input_len] = make([]complex128, input_len)

		var sin, cos float64
		for i := 0; i < input_len; i++ {
			if i == 0 {
				sin, cos = 0, 1
			} else {
				sin, cos = math.Sincos(math.Pi / float64(input_len) * float64(i*i))
			}
			bluesteinFactors[input_len][i] = complex(cos, sin)
			bluesteinInvFactors[input_len][i] = complex(cos, -sin)
		}
	}

	return bluesteinFactors[input_len], bluesteinInvFactors[input_len]
}

func hasBluesteinFactors(idx int) bool {
	return bluesteinFactors[idx] != nil
}

// FFTReal returns the forward FFT of the real-valued slice.
func FFTReal(x []float64) []complex128 {
	return FFT(dsputils.ToComplex(x))
}

// IFFTReal returns the inverse FFT of the real-valued slice.
func IFFTReal(x []float64) []complex128 {
	return IFFT(dsputils.ToComplex(x))
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
func Convolve(x, y []complex128) []complex128 {
	if len(x) != len(y) {
		panic("arrays not of equal size")
	}

	fft_x := FFT(x)
	fft_y := FFT(y)

	r := make([]complex128, len(x))
	for i := 0; i < len(r); i++ {
		r[i] = fft_x[i] * fft_y[i]
	}

	return IFFT(r)
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

	if dsputils.IsPowerOf2(lx) {
		return radix2FFT(x)
	}

	return bluesteinFFT(x)
}

// radix2FFT returns the FFT calculated using the radix-2 DIT Cooley-Tukey algorithm.
func radix2FFT(x []complex128) []complex128 {
	lx := len(x)
	factors := getRadix2Factors(lx)

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
					w_n := r[j+nb+s_2] * factors[blocks*j]
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
	a := dsputils.ZeroPad(x, dsputils.NextPowerOf2(lx*2-1))
	la := len(a)
	factors, invFactors := getBluesteinFactors(lx)

	for n, v := range x {
		a[n] = v * invFactors[n]
	}

	b := make([]complex128, la)
	for i := 0; i < lx; i++ {
		b[i] = factors[i]

		if i != 0 {
			b[la-i] = factors[i]
		}
	}

	r := Convolve(a, b)

	for i := 0; i < lx; i++ {
		r[i] *= invFactors[i]
	}

	return r[:lx]
}

// FFT2Real returns the 2-dimensional, forward FFT of the real-valued matrix.
func FFT2Real(x [][]float64) [][]complex128 {
	return FFT2(dsputils.ToComplex2(x))
}

// FFT2 returns the 2-dimensional, forward FFT of the complex-valued matrix.
func FFT2(x [][]complex128) [][]complex128 {
	return computeFFT2(x, FFT)
}

// IFFT2Real returns the 2-dimensional, inverse FFT of the real-valued matrix.
func IFFT2Real(x [][]float64) [][]complex128 {
	return IFFT2(dsputils.ToComplex2(x))
}

// IFFT2 returns the 2-dimensional, inverse FFT of the complex-valued matrix.
func IFFT2(x [][]complex128) [][]complex128 {
	return computeFFT2(x, IFFT)
}

func computeFFT2(x [][]complex128, fftFunc func([]complex128) []complex128) [][]complex128 {
	rows := len(x)
	if rows == 0 {
		panic("empty input array")
	}

	cols := len(x[0])
	r := make([][]complex128, rows)
	for i := 0; i < rows; i++ {
		if len(x[i]) != cols {
			panic("ragged input array")
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

	return r
}
