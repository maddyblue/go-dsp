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
	"testing"

	"github.com/mjibson/go-dsp/dsputils"
)

type windowTest struct {
	in       int
	hamming  []float64
	hann     []float64
	bartlett []float64
}

var windowTests = []windowTest{
	{
		1,
		[]float64{1},
		[]float64{1},
		[]float64{1},
	},
	{
		5,
		[]float64{0.08, 0.54, 1, 0.54, 0.08},
		[]float64{0, 0.5, 1, 0.5, 0},
		[]float64{0, 0.5, 1, 0.5, 0},
	},
	{
		10,
		[]float64{0.08, 0.18761956, 0.46012184, 0.77, 0.97225861, 0.97225861, 0.77, 0.46012184, 0.18761956, 0.08},
		[]float64{0, 0.116977778440511, 0.413175911166535, 0.75, 0.969846310392954, 0.969846310392954, 0.75, 0.413175911166535, 0.116977778440511, 0},
		[]float64{0, 0.222222222222222, 0.444444444444444, 0.666666666666667, 0.888888888888889, 0.888888888888889, 0.666666666666667, 0.444444444444444, 0.222222222222222, 0},
	},
}

func TestWindowFunctions(t *testing.T) {
	for _, v := range windowTests {
		o := Hamming(v.in)
		if !dsputils.PrettyClose(o, v.hamming) {
			t.Error("hamming error\ninput:", v.in, "\noutput:", o, "\nexpected:", v.hamming)
		}

		o = Hann(v.in)
		if !dsputils.PrettyClose(o, v.hann) {
			t.Error("hann error\ninput:", v.in, "\noutput:", o, "\nexpected:", v.hann)
		}

		o = Bartlett(v.in)
		if !dsputils.PrettyClose(o, v.bartlett) {
			t.Error("bartlett error\ninput:", v.in, "\noutput:", o, "\nexpected:", v.bartlett)
		}

		o = Rectangular(v.in)
		Apply(o, Hamming)
		if !dsputils.PrettyClose(o, v.hamming) {
			t.Error("apply error\noutput:", o, "\nexpected:", v.hamming)
		}
	}
}
