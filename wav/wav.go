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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

func (wavHeader *WavHeader) setup(header []byte) (err error) {
	if err = checkHeader(header); err != nil {
		return
	}
	wavHeader.AudioFormat, err = bLEtoUint16(header, AudioFormatOffset)
	if err != nil {
		return
	}

	wavHeader.NumChannels, err = bLEtoUint16(header, NumChannelsOffset)
	if err != nil {
		return
	}

	wavHeader.SampleRate, err = bLEtoUint32(header, SampleRateOffset)
	if err != nil {
		return
	}

	wavHeader.ByteRate, err = bLEtoUint32(header, ByteRateOffset)
	if err != nil {
		return
	}

	wavHeader.BlockAlign, err = bLEtoUint16(header, BlockAlignOffset)
	if err != nil {
		return
	}

	wavHeader.BitsPerSample, err = bLEtoUint16(header, BitsPerSampleOffset)
	if err != nil {
		return
	}

	wavHeader.ChunkSize, err = bLEtoUint32(header, ChunkSizeOffset)
	if err != nil {
		return
	}

	wavHeader.NumSamples = int(wavHeader.ChunkSize) / int(wavHeader.BlockAlign)

	switch wavHeader.AudioFormat {
	case wavFormatPCM:
	case wavFormatIEEEFloat:
	default:
		err = fmt.Errorf("wav: unknown audio format; %02x", wavHeader.AudioFormat)
	}
	return
}

// Returns a single sample laid out by channel e.g. [ch0, ch1, ...]
func readSample(data []byte, sampleIndex int, header WavHeader) (n []int, f []float32, err error) {
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
				var value int16
				value, err = bLEtoInt16(data, 2*si)
				if err != nil {
					return
				}
				n[ch] = int(value)
			}
		case wavFormatIEEEFloat:
			switch header.BitsPerSample {
			case 32:
				f[ch], err = bLEtoFloat32(data, 4*si)
				if err != nil {
					return
				}
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

			sample, _, err := readSample(data, i, wav.WavHeader)
			if err != nil {
				return nil, err
			}

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

			sample, _, err := readSample(data, i, wav.WavHeader)
			if err != nil {
				return nil, err
			}

			wav.Data[i] = sample
			for ch := uint16(0); ch < wav.NumChannels; ch++ {
				wav.Data16[i][ch] = int16(sample[ch])
			}
		}
	} else if wav.BitsPerSample == 32 {
		wav.Float32 = make([][]float32, wav.NumSamples)
		for i := 0; i < wav.NumSamples; i++ {
			_, wav.Float32[i], err = readSample(data, i, wav.WavHeader)
			if err != nil {
				return nil, err
			}
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
		samples[sampleIndex], _, err = readSample(data, sampleIndex, wav.WavHeader)
		if err != nil {
			return
		}
	}

	return
}

// little-endian [4]byte to uint32 conversion
func bLEtoUint32(b []byte, idx uint16) (value uint32, err error) {
	buf := bytes.NewReader([]byte{b[idx], b[idx+1], b[idx+2], b[idx+3]})
	err = binary.Read(buf, binary.LittleEndian, &value)

	return
}

// little-endian [2]byte to uint16 conversion
func bLEtoUint16(b []byte, idx uint16) (value uint16, err error) {
	buf := bytes.NewReader([]byte{b[idx], b[idx+1]})
	err = binary.Read(buf, binary.LittleEndian, &value)

	return
}

func bLEtoInt16(b []byte, idx uint16) (value int16, err error) {
	buf := bytes.NewReader([]byte{b[idx], b[idx+1]})
	err = binary.Read(buf, binary.LittleEndian, &value)

	return
}

func bLEtoFloat32(b []byte, idx uint16) (value float32, err error) {
	buf := bytes.NewReader([]byte{b[idx], b[idx+1], b[idx+2], b[idx+3]})
	err = binary.Read(buf, binary.LittleEndian, &value)

	return
}
