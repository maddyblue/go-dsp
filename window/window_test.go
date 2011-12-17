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

package window

import (
	"dsputils"
	"testing"
)

type windowTest struct {
	in int
	hamming []float64
}

var windowTests = []windowTest{
	windowTest{
		5,
		[]float64 {0.08, 0.54, 1, 0.54, 0.08},
	},

	windowTest{
		10,
		[]float64 {0.08, 0.18761956, 0.46012184, 0.77, 0.97225861, 0.97225861, 0.77, 0.46012184, 0.18761956, 0.08},
	},
}

func TestHamming(t *testing.T) {
	for _, v := range windowTests {
		o := Hamming(v.in)
		if !dsputils.PrettyClose(o, v.hamming) {
			t.Error("hamming error\ninput:", v.in, "\noutput:", o, "\nexpected:", v.hamming)
		}
	}
}
