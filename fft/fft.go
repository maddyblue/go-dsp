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

var factors = map[int][]complex128{}
var inv_factors = map[int][]complex128{}

// Ensures the complex multiplication factors exist for an input array of length input_len.
func ensureFactors(input_len int) {
	var cos, sin float64
	for i := 4; i <= input_len; i <<= 1 {
		if _, present := factors[i]; !present {
			factors[i] = make([]complex128, i)
			inv_factors[i] = make([]complex128, i)
			for n := 0; n < i; n++ {
				if n == 0 {
					sin, cos = 0, 1
				} else if n*4 == i {
					sin, cos = -1, 0
				} else {
					sin, cos = math.Sincos(-2 * math.Pi / float64(i) * float64(n))
				}
				factors[i][n] = complex(cos, sin)
				inv_factors[i][n] = complex(cos, -sin)
			}
		}
	}
}

func Fft(x []float64) []complex128 {
	lx := len(x)
	ensureFactors(lx)

	// todo: non-hack handling length <= 1 cases
	if lx <= 1 {
		r := make([]complex128, lx)
		for n, v := range x {
			r[n] = complex(v, 0)
		}
		return r
	}

	// expand a length of a power of 2
	if lx&(lx-1) != 0 { // not a power of 2
		nl := int(math.Pow(2, math.Ceil(math.Log2(float64(lx)))))
		nx := make([]float64, nl)
		copy(nx, x)
		for i := 0; i < nl-lx; i++ {
			nx[i+lx] = 0
		}
		x = nx[:]
		lx = nl
	}

	lx_2 := lx / 2

	r := make([]complex128, lx) // result
	t := make([]complex128, lx) // temp

	for n, v := range x {
		r[n] = complex(v, 0)
	}

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
