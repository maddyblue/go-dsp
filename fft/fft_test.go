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

package fft

import (
	"math"
	"testing"
)

const (
	sqrt2_2     = math.Sqrt2 / 2
	closeFactor = 1e-8 // todo: test on a 32-bit machine
)

type fftTest struct {
	in  []float64
	out []complex128
}

var fftTests = []fftTest{
	// impulse responses
	fftTest{
		[]float64{1},
		[]complex128{complex(1, 0)},
	},
	fftTest{
		[]float64{1, 0},
		[]complex128{complex(1, 0), complex(1, 0)},
	},
	fftTest{
		[]float64{1, 0, 0, 0},
		[]complex128{complex(1, 0), complex(1, 0), complex(1, 0), complex(1, 0)},
	},
	fftTest{
		[]float64{1, 0, 0, 0, 0, 0, 0, 0},
		[]complex128{
			complex(1, 0),
			complex(1, 0),
			complex(1, 0),
			complex(1, 0),
			complex(1, 0),
			complex(1, 0),
			complex(1, 0),
			complex(1, 0)},
	},

	// shifted impulse response
	fftTest{
		[]float64{0, 1},
		[]complex128{complex(1, 0), complex(-1, 0)},
	},
	fftTest{
		[]float64{0, 1, 0, 0},
		[]complex128{complex(1, 0), complex(0, -1), complex(-1, 0), complex(0, 1)},
	},
	fftTest{
		[]float64{0, 1, 0, 0, 0, 0, 0, 0},
		[]complex128{
			complex(1, 0),
			complex(sqrt2_2, -sqrt2_2),
			complex(0, -1),
			complex(-sqrt2_2, -sqrt2_2),
			complex(-1, 0),
			complex(-sqrt2_2, sqrt2_2),
			complex(0, 1),
			complex(sqrt2_2, sqrt2_2)},
	},

	// other
	fftTest{
		[]float64{1, 2, 3, 4},
		[]complex128{
			complex(10, 0),
			complex(-2, 2),
			complex(-2, 0),
			complex(-2, -2)},
	},
	fftTest{
		[]float64{1, 3, 5, 7},
		[]complex128{
			complex(16, 0),
			complex(-4, 4),
			complex(-4, 0),
			complex(-4, -4)},
	},
	fftTest{
		[]float64{1, 2, 3, 4, 5, 6, 7, 8},
		[]complex128{
			complex(36, 0),
			complex(-4, 9.65685425),
			complex(-4, 4),
			complex(-4, 1.65685425),
			complex(-4, 0),
			complex(-4, -1.65685425),
			complex(-4, -4),
			complex(-4, -9.65685425)},
	},

	// non power of 2 lengths
	fftTest{
		[]float64{1, 0, 0, 0, 0},
		[]complex128{
			complex(1, 0),
			complex(1, 0),
			complex(1, 0),
			complex(1, 0),
			complex(1, 0)},
	},
	fftTest{
		[]float64{1, 2, 3},
		[]complex128{
			complex(6, 0),
			complex(-1.5, 0.8660254),
			complex(-1.5, -0.8660254)},
	},
	fftTest{
		[]float64{1, 1, 1},
		[]complex128{
			complex(3, 0),
			complex(0, 0),
			complex(0, 0)},
	},
}

type fft2Test struct {
	in  [][]float64
	out [][]complex128
}

var fft2Tests = []fft2Test{
	fft2Test{
		[][]float64{{1, 2, 3}, {3, 4, 5}},
		[][]complex128{
			{complex(18, 0), complex(-3, 1.73205081), complex(-3, -1.73205081)},
			{complex(-6, 0), complex(0, 0), complex(0, 0)}},
	},
	fft2Test{
		[][]float64{{0.1, 0.2, 0.3, 0.4, 0.5}, {1, 2, 3, 4, 5}, {3, 2, 1, 0, -1}},
		[][]complex128{
			{complex(21.5, 0), complex(-0.25, 0.34409548), complex(-0.25, 0.08122992), complex(-0.25, -0.08122992), complex(-0.25, -0.34409548)},
			{complex(-8.5, -8.66025404), complex(5.70990854, 4.6742225), complex(1.15694356, 4.41135694), complex(-1.65694356, 4.24889709), complex(-6.20990854, 3.98603154)},
			{complex(-8.5, 8.66025404), complex(-6.20990854, -3.98603154), complex(-1.65694356, -4.24889709), complex(1.15694356, -4.41135694), complex(5.70990854, -4.6742225)}},
	},
}

func prettyClose(a, b []complex128) bool {
	if len(a) != len(b) {
		return false
	}

	for i, c := range a {
		if !ComplexEqual(c, b[i]) {
			return false
		}
	}
	return true
}

func prettyClose2(a, b [][]complex128) bool {
	if len(a) != len(b) {
		return false
	}

	for i, c := range a {
		if !prettyClose(c, b[i]) {
			return false
		}
	}
	return true
}

// returns true if a and b are very close, else false
func ComplexEqual(a, b complex128) bool {
	r_a := real(a)
	r_b := real(b)
	i_a := imag(a)
	i_b := imag(b)

	return ((math.Fabs(r_a-r_b) <= closeFactor || math.Fabs(1-r_a/r_b) <= closeFactor) &&
		(math.Fabs(i_a-i_b) <= closeFactor || math.Fabs(1-i_a/i_b) <= closeFactor))
}

func TestFFT(t *testing.T) {
	for _, ft := range fftTests {
		v := FFTReal(ft.in)
		if !prettyClose(v, ft.out) {
			t.Error("FFT error\ninput:", ft.in, "\noutput:", v, "\nexpected:", ft.out)
		}

		vi := IFFT(ft.out)
		if !prettyClose(vi, toComplex(ft.in)) {
			t.Error("IFFT error\ninput:", ft.out, "\noutput:", vi, "\nexpected:", toComplex(ft.in))
		}
	}
}

func TestFFT2(t *testing.T) {
	for _, ft := range fft2Tests {
		v, _ := FFT2Real(ft.in)
		if !prettyClose2(v, ft.out) {
			t.Error("FFT2 error\ninput:", ft.in, "\noutput:", v, "\nexpected:", ft.out)
		}

		vi, _ := IFFT2(ft.out)
		if !prettyClose2(vi, toComplex2(ft.in)) {
			t.Error("IFFT2 error\ninput:", ft.out, "\noutput:", vi, "\nexpected:", toComplex2(ft.in))
		}
	}
}
