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
)

func Fft(x []float64) []complex128 {
	lx := len(x)

	// expand a length of a power of 2
	if lx & (lx - 1) != 0 { // not a power of 2
		nl := int(math.Pow(2, math.Ceil(math.Log2(float64(lx)))))
		nx := make([]float64, nl)
		copy(nx, x)
		for i := 0; i < nl - lx; i++ {
			nx[i + lx] = 0
		}
		x = nx[:]
		lx = nl
	}

	lx_2 := lx / 2

	r := make([]complex128, lx) // result
	t := make([]complex128, lx) // temp
	w := make([]complex128, lx_2)
	w[0] = complex(1, 0)

	// split into event and odd parts
	for n := 0; n < lx_2; n++ {
		r[n] = complex(x[n * 2], 0)
		r[n + lx_2] = complex(x[n * 2 + 1], 0)
	}

	for i := 0; i < int(math.Log2(float64(lx))); i++ {
		stage := 1 << uint(i + 1)

		// 2-point transforms
		if i == 0 {
			for n := 0; n < lx_2; n++ {
				t[n * 2] = r[n * 2] + r[n * 2 + 1]
				t[n * 2 + 1] = r[n * 2] - r[n * 2 + 1]
			}
		// >2-point transforms
		} else {
			blocks := lx / stage
			s_2 := stage / 2

			for j := 1; j < s_2; j++ {
				if j * 4 == stage {
					w[j] = complex(0, -1)
				} else {
					sin, cos := math.Sincos(-2 * math.Pi / float64(stage) * float64(j))
					w[j] = complex(cos, sin)
				}
			}

			for n := 0; n < blocks; n++ {
				nb := n * stage
				for j := 0; j < s_2; j++ {
					w_n := r[j + nb + s_2] * w[j]
					t[j + nb      ] = r[j + nb] + w_n
					t[j + nb + s_2] = r[j + nb] - w_n
				}
			}
		}

		copy(r, t)
	}

	return r
}
