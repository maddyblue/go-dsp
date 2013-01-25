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

// Package wav provides support for the WAV file format.
package wav

import (
	"errors"
	"io"
	"io/ioutil"
)

type Wav struct {
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	ChunkSize     uint32
	NumSamples    int

	// The Data corresponding to BitsPerSample is populated, indexed by channel.
	Data8  [][]uint8
	Data16 [][]int16

	// Data is always populated, indexed by channel. It is a copy of DataXX.
	Data [][]int
}

// ReadWav reads a wav file.
func ReadWav(r io.Reader) (*Wav, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(b) < 44 ||
		string(b[0:4]) != "RIFF" ||
		string(b[8:12]) != "WAVE" ||
		bLEtoUint32(b, 4) != uint32(len(b)) ||
		string(b[12:16]) != "fmt " ||
		string(b[36:40]) != "data" ||
		bLEtoUint32(b, 40) != uint32(len(b)-44) {
		return nil, errors.New("wav: not a WAV")
	}

	w := Wav{
		AudioFormat:   bLEtoUint16(b, 20),
		NumChannels:   bLEtoUint16(b, 22),
		SampleRate:    bLEtoUint32(b, 24),
		ByteRate:      bLEtoUint32(b, 28),
		BlockAlign:    bLEtoUint16(b, 32),
		BitsPerSample: bLEtoUint16(b, 34),
		ChunkSize:     bLEtoUint32(b, 40),
	}
	w.NumSamples = int(w.ChunkSize) / int(w.BlockAlign)

	data := b[44 : w.ChunkSize+44]

	w.Data = make([][]int, w.NumChannels)

	if w.BitsPerSample == 8 {
		w.Data8 = make([][]uint8, w.NumChannels)
		for ch := 0; ch < int(w.NumChannels); ch++ {
			w.Data8[ch] = make([]uint8, w.NumSamples)
			w.Data[ch] = make([]int, w.NumSamples)
		}

		for i := 0; i < int(w.ChunkSize); i++ {
			for ch := 0; ch < int(w.NumChannels); ch++ {
				w.Data8[ch][i] = uint8(data[i*int(w.NumChannels)+ch])
				w.Data[ch][i] = int(w.Data8[ch][i])
			}
		}
	} else if w.BitsPerSample == 16 {
		w.Data16 = make([][]int16, w.NumChannels)
		for ch := 0; ch < int(w.NumChannels); ch++ {
			w.Data16[ch] = make([]int16, w.NumSamples)
			w.Data[ch] = make([]int, w.NumSamples)
		}

		for i := 0; i < int(w.ChunkSize)/int(w.BlockAlign); i++ {
			for ch := 0; ch < int(w.NumChannels); ch++ {
				w.Data16[ch][i] = int16(data[i*int(w.NumChannels)+ch])
				w.Data[ch][i] = int(w.Data16[ch][i])
			}
		}
	}

	return &w, nil
}

// little-endian [4]byte to uint32 conversion
func bLEtoUint32(b []byte, idx int) uint32 {
	return uint32(b[idx+3])<<24 +
		uint32(b[idx+2])<<16 +
		uint32(b[idx+1])<<8 +
		uint32(b[idx])
}

// little-endian [2]byte to uint16 conversion
func bLEtoUint16(b []byte, idx int) uint16 {
	return uint16(b[idx+1])<<8 + uint16(b[idx])
}

func bLEtoInt16(b []byte, idx int) int16 {
	return int16(b[idx+1])<<8 + int16(b[idx])
}
