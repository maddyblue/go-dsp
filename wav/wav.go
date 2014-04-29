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
	"fmt"
	"io"
	"io/ioutil"
	"math"
)

const (
	RIFFMarkerOffset = 0
	WAVEMarkerOffset = 8
	FMTMarkerOffset  = 12
	DataMarkerOffset = 36

	AudioFormatOffset   = 20
	NumChannelsOffset   = 22
	SampleRateOffset    = 24
	ByteRateOffset      = 28
	BlockAlignOffset    = 32
	BitsPerSampleOffset = 34
	ChunkSizeOffset     = 40
	ExpectedHeaderSize  = 44

	wavFormatPCM       = 1
	wavFormatIEEEFloat = 3
)

type WavHeader struct {
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	ChunkSize     uint32
	NumSamples    int
}

type Wav struct {
	WavHeader

	// DataXX for the corresponding BitsPerSample is populated, indexed by sample then channel.
	Data8  [][]uint8
	Data16 [][]int16

	// Data is populated for 8- and 16-bit samples. It is a copy of DataXX.
	Data [][]int

	// Float32 is populated for 32-bit samples.
	Float32 [][]float32
}

type StreamedWav struct {
	WavHeader
	io.Reader
}

func checkHeader(header []byte) error {
	if len(header) < ExpectedHeaderSize {
		return errors.New("wav: Invalid header size")
	}
	if string(header[0:4]) != "RIFF" {
		return errors.New("wav: Header does not conatin 'RIFF'")
	}
	if string(header[8:12]) != "WAVE" {
		return errors.New("wav: Header does not contain 'WAVE'")
	}
	if string(header[12:16]) != "fmt " {
		return errors.New("wav: Header does not contain 'fmt'")
	}
	if string(header[36:40]) != "data" {
		return errors.New("wav: Header does not contain 'data'")
	}
	return nil
}

func (wavHeader *WavHeader) setup(header []byte) error {
	if err := checkHeader(header); err != nil {
		return err
	}
	wavHeader.AudioFormat = bLEtoUint16(header, AudioFormatOffset)
	wavHeader.NumChannels = bLEtoUint16(header, NumChannelsOffset)
	wavHeader.SampleRate = bLEtoUint32(header, SampleRateOffset)
	wavHeader.ByteRate = bLEtoUint32(header, ByteRateOffset)
	wavHeader.BlockAlign = bLEtoUint16(header, BlockAlignOffset)
	wavHeader.BitsPerSample = bLEtoUint16(header, BitsPerSampleOffset)
	wavHeader.ChunkSize = bLEtoUint32(header, ChunkSizeOffset)
	wavHeader.NumSamples = int(wavHeader.ChunkSize) / int(wavHeader.BlockAlign)
	switch wavHeader.AudioFormat {
	case wavFormatPCM:
	case wavFormatIEEEFloat:
	default:
		return fmt.Errorf("wav: unknown audio format; %02x", wavHeader.AudioFormat)
	}
	return nil
}

// Returns a single sample laid out by channel e.g. [ch0, ch1, ...]
func readSample(data []byte, sampleIndex int, header WavHeader) (n []int, f []float32) {
	n = make([]int, header.NumChannels)
	f = make([]float32, header.NumChannels)
	for ch := uint16(0); ch < header.NumChannels; ch++ {
		si := uint16(sampleIndex)*header.NumChannels + ch
		switch header.AudioFormat {
		case wavFormatPCM:
			switch header.BitsPerSample {
			case 8:
				n[ch] = int(data[si])
			case 16:
				n[ch] = int(bLEtoInt16(data, 2*si))
			}
		case wavFormatIEEEFloat:
			switch header.BitsPerSample {
			case 32:
				f[ch] = bLEtoFloat32(data, 4*si)
			}
		}
	}
	return
}

// ReadWav reads a wav file.
func ReadWav(r io.Reader) (wav *Wav, err error) {
	if r == nil {
		return nil, errors.New("wav: Invalid Reader")
	}
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	wav = new(Wav)
	err = wav.WavHeader.setup(bytes)
	if err != nil {
		return nil, err
	}
	data := bytes[ExpectedHeaderSize : int(wav.ChunkSize)+ExpectedHeaderSize]
	if wav.BitsPerSample == 8 {
		wav.Data = make([][]int, wav.NumSamples)
		wav.Data8 = make([][]uint8, wav.NumSamples)
		for i := 0; i < wav.NumSamples; i++ {
			wav.Data8[i] = make([]uint8, wav.NumChannels)
			sample, _ := readSample(data, i, wav.WavHeader)
			wav.Data[i] = sample
			for ch := uint16(0); ch < wav.NumChannels; ch++ {
				wav.Data8[i][ch] = uint8(sample[ch])
			}
		}
	} else if wav.BitsPerSample == 16 {
		wav.Data = make([][]int, wav.NumSamples)
		wav.Data16 = make([][]int16, wav.NumSamples)
		for i := 0; i < wav.NumSamples; i++ {
			wav.Data16[i] = make([]int16, wav.NumChannels)
			sample, _ := readSample(data, i, wav.WavHeader)
			wav.Data[i] = sample
			for ch := uint16(0); ch < wav.NumChannels; ch++ {
				wav.Data16[i][ch] = int16(sample[ch])
			}
		}
	} else if wav.BitsPerSample == 32 {
		wav.Float32 = make([][]float32, wav.NumSamples)
		for i := 0; i < wav.NumSamples; i++ {
			_, wav.Float32[i] = readSample(data, i, wav.WavHeader)
		}
	} else {
		return nil, fmt.Errorf("wav: unknown bits per sample: %v", wav.BitsPerSample)
	}
	return
}

// StreamedWav returns a wav for streamed reading.
func StreamWav(reader io.Reader) (wav *StreamedWav, err error) {
	if reader == nil {
		return nil, errors.New("wav: Invalid Reader")
	}

	header := make([]byte, ExpectedHeaderSize)
	_, err = reader.Read(header)
	if err != nil {
		return nil, err
	}

	wav = new(StreamedWav)
	err = wav.setup(header)
	if err != nil {
		return nil, err
	}

	wav.Reader = reader

	return
}

// ReadSamples returns an array of [channelIndex][sampleIndex] samples. The
// number of samples returned may be less than the amount requested depending
// on the amount of data available.
func (wav *StreamedWav) ReadSamples(numSamples int) (samples [][]int, err error) {
	data := make([]byte, numSamples*int(wav.BlockAlign))
	amountRead, err := wav.Reader.Read(data)
	if err != nil {
		return
	}
	if amountRead%int(wav.BlockAlign) != 0 {
		err = errors.New("wav: Read an invalid amount of data")
		return
	}

	numberOfSamplesRead := amountRead / int(wav.BlockAlign)
	samples = make([][]int, numberOfSamplesRead)

	for sampleIndex := 0; sampleIndex < numberOfSamplesRead; sampleIndex++ {
		samples[sampleIndex], _ = readSample(data, sampleIndex, wav.WavHeader)
	}

	return
}

// little-endian [4]byte to uint32 conversion
func bLEtoUint32(b []byte, idx int) uint32 {
	return uint32(b[idx+3])<<24 +
		uint32(b[idx+2])<<16 +
		uint32(b[idx+1])<<8 +
		uint32(b[idx])
}

// little-endian [2]byte to uint16 conversion
func bLEtoUint16(b []byte, idx uint16) uint16 {
	return uint16(b[idx+1])<<8 + uint16(b[idx])
}

func bLEtoInt16(b []byte, idx uint16) int16 {
	return int16(b[idx+1])<<8 + int16(b[idx])
}

func bLEtoFloat32(b []byte, idx uint16) float32 {
	var u uint32
	u += uint32(b[idx+3]) << 24
	u += uint32(b[idx+2]) << 16
	u += uint32(b[idx+1]) << 8
	u += uint32(b[idx])
	return math.Float32frombits(u)
}
