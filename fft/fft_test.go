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
	sqrt2_2 = math.Sqrt2 / 2
	closeFactor = 1e-14 // todo: test on a 32-bit machine
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
		[]complex128{complex(1, 0), complex(1, 0), complex(1, 0), complex(1, 0), complex(1, 0), complex(1, 0), complex(1, 0), complex(1, 0)},
	},

	// impulse shifted response
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
		[]complex128{complex(1, 0), complex(sqrt2_2, -sqrt2_2), complex(0, -1), complex(-sqrt2_2, -sqrt2_2), complex(-1, 0), complex(-sqrt2_2, sqrt2_2), complex(0, 1), complex(sqrt2_2, sqrt2_2)},
	},
}

func prettyClose(a, b []complex128) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		a_r := real(c)
		a_i := imag(c)
		b_r := real(b[i])
		b_i := imag(b[i])

		if math.Fabs(1 - a_r / b_r) > closeFactor ||
		   math.Fabs(1 - a_i / b_i) > closeFactor {
			return false
		}

	}
	return true
}

func TestFft(t *testing.T) {
	for _, ft := range fftTests {
		v := Fft(ft.in)
		if !prettyClose(v, ft.out) {
			t.Errorf("input: %s\noutput: %s\nexcepted: %s", ft.in, v, ft.out)
		}
	}
}
