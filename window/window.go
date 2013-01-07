/*
 * Copyright (c) 2012 Matt Jibson <matt.jibson@gmail.com>
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

// Package window provides window functions for digital signal processing.
package window

import (
	"math"
)

// Apply applies the window windowFunction to x.
func Apply(x []float64, windowFunction func(int) []float64) {
	for i, w := range windowFunction(len(x)) {
		x[i] *= w
	}
}

// Rectangular returns an L-point rectangular window (all values are 1).
func Rectangular(L int) []float64 {
	r := make([]float64, L)

	for i := range r {
		r[i] = 1
	}

	return r
}

// Hamming returns an L-point symmetric Hamming window.
// Reference: http://www.mathworks.com/help/signal/ref/hamming.html
func Hamming(L int) []float64 {
	r := make([]float64, L)

	if L == 1 {
		r[0] = 1
	} else {
		N := L - 1
		coef := math.Pi * 2 / float64(N)
		for n := 0; n <= N; n++ {
			r[n] = 0.54 - 0.46*math.Cos(coef*float64(n))
		}
	}

	return r
}

// Hann returns an L-point Hann window.
// Reference: http://www.mathworks.com/help/signal/ref/hann.html
func Hann(L int) []float64 {
	r := make([]float64, L)

	if L == 1 {
		r[0] = 1
	} else {
		N := L - 1
		coef := 2 * math.Pi / float64(N)
		for n := 0; n <= N; n++ {
			r[n] = 0.5 * (1 - math.Cos(coef*float64(n)))
		}
	}

	return r
}

// Bartlett returns an L-point Bartlett window.
// Reference: http://www.mathworks.com/help/signal/ref/bartlett.html
func Bartlett(L int) []float64 {
	r := make([]float64, L)

	if L == 1 {
		r[0] = 1
	} else {
		N := L - 1
		coef := 2 / float64(N)
		n := 0
		for ; n <= N/2; n++ {
			r[n] = coef * float64(n)
		}
		for ; n <= N; n++ {
			r[n] = 2 - coef*float64(n)
		}
	}

	return r
}
