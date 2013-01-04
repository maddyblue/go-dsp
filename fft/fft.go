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
	"runtime"
	"sync"

	"github.com/mjibson/go-dsp/dsputils"
)

var (
	radix2Lock    sync.RWMutex
	radix2Factors = map[int][]complex128{
		4: {complex(1, 0), complex(0, -1), complex(-1, 0), complex(0, 1)},
	}
)

var (
	bluesteinLock       sync.RWMutex
	bluesteinFactors    = map[int][]complex128{}
	bluesteinInvFactors = map[int][]complex128{}
)

// EnsureRadix2Factors ensures that all radix 2 factors are computed for inputs
// of length input_len. This is used to precompute needed factors for known
// sizes. Generally should only be used for benchmarks.
func EnsureRadix2Factors(input_len int) {
	getRadix2Factors(input_len)
}

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

// Convolve returns the convolution of x ∗ y.
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

var (
	worker_pool_size = 0
)

// SetWorkerPoolSize sets the number of workers during FFT computation on multicore systems.
// If n is 0 (the default), then GOMAXPROCS workers will be created.
func SetWorkerPoolSize(n int) {
	if n < 0 {
		n = 0
	}

	worker_pool_size = n
}

type fft_work struct {
	start, end int
}

// radix2FFT returns the FFT calculated using the radix-2 DIT Cooley-Tukey algorithm.
func radix2FFT(x []complex128) []complex128 {
	lx := len(x)
	factors := getRadix2Factors(lx)

	t := make([]complex128, lx) // temp
	r := reorderData(x)

	var blocks, stage, s_2 int

	jobs := make(chan *fft_work, lx)
	results := make(chan bool, lx)

	num_workers := worker_pool_size
	if (num_workers) == 0 {
		num_workers = runtime.GOMAXPROCS(0)
	}

	idx_diff := lx / num_workers
	if idx_diff < 2 {
		idx_diff = 2
	}

	worker := func() {
		for work := range jobs {
			for nb := work.start; nb < work.end; nb += stage {
				if stage != 2 {
					for j := 0; j < s_2; j++ {
						idx := j + nb
						idx2 := idx + s_2
						ridx := r[idx]
						w_n := r[idx2] * factors[blocks*j]
						t[idx] = ridx + w_n
						t[idx2] = ridx - w_n
					}
				} else {
					n1 := nb + 1
					rn := r[nb]
					rn1 := r[n1]
					t[nb] = rn + rn1
					t[n1] = rn - rn1
				}
			}

			results <- true
		}
	}

	for i := 0; i < num_workers; i++ {
		go worker()
	}
	defer close(jobs)

	for stage = 2; stage <= lx; stage <<= 1 {
		blocks = lx / stage
		s_2 = stage / 2
		workers_spawned := 0

		for start, end := 0, stage; ; {
			if end-start >= idx_diff || end == lx {
				workers_spawned++
				jobs <- &fft_work{start, end}

				if end == lx {
					break
				}

				start = end
			}

			end += stage
		}

		for n := 0; n < workers_spawned; n++ {
			<-results
		}

		r, t = t, r
	}

	return r
}

// reorderData returns a copy of x reordered for the DFT.
func reorderData(x []complex128) []complex128 {
	lx := uint(len(x))
	r := make([]complex128, lx)
	s := log2(lx)

	var n uint
	for ; n < lx; n++ {
		r[reverseBits(n, s)] = x[n]
	}

	return r
}

// log2 returns the log base 2 of v
// from: http://graphics.stanford.edu/~seander/bithacks.html#IntegerLogObvious
func log2(v uint) uint {
	var r uint

	for v >>= 1; v != 0; v >>= 1 {
		r++
	}

	return r
}

// reverseBits returns the first s bits of v in reverse order
// from: http://graphics.stanford.edu/~seander/bithacks.html#BitReverseObvious
func reverseBits(v, s uint) uint {
	var r uint

	// Since we aren't reversing all the bits in v (just the first s bits),
	// we only need the first bit of v instead of a full copy.
	r = v & 1
	s--

	for v >>= 1; v != 0; v >>= 1 {
		r <<= 1
		r |= v & 1
		s--
	}

	return r << s
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

// FFTN returns the forward FFT of the matrix m, computed in all N dimensions.
func FFTN(m *dsputils.Matrix) *dsputils.Matrix {
	return computeFFTN(m, FFT)
}

// IFFTN returns the forward FFT of the matrix m, computed in all N dimensions.
func IFFTN(m *dsputils.Matrix) *dsputils.Matrix {
	return computeFFTN(m, IFFT)
}

func computeFFTN(m *dsputils.Matrix, fftFunc func([]complex128) []complex128) *dsputils.Matrix {
	dims := m.Dimensions()
	t := m.Copy()
	r := dsputils.MakeEmptyMatrix(dims)

	for n := range dims {
		dims[n] -= 1
	}

	for n := range dims {
		d := make([]int, len(dims))
		copy(d, dims)
		d[n] = -1

		for {
			r.SetDim(fftFunc(t.Dim(d)), d)

			if !decrDim(d, dims) {
				break
			}
		}

		r, t = t, r
	}

	return t
}

// decrDim decrements an element of x by 1, skipping all -1s, and wrapping up to d.
// If a value is 0, it will be reset to its correspending value in d, and will carry one from the next non -1 value to the right.
// Returns true if decremented, else false.
func decrDim(x, d []int) bool {
	for n, v := range x {
		if v == -1 {
			continue
		} else if v == 0 {
			i := n
			// find the next element to decrement
			for ; i < len(x); i++ {
				if x[i] == -1 {
					continue
				} else if x[i] == 0 {
					x[i] = d[i]
				} else {
					x[i] -= 1
					return true
				}
			}

			// no decrement
			return false
		} else {
			x[n] -= 1
			return true
		}
	}

	return false
}
