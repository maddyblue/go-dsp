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

const (
	AudioFormatOffset   int = 20
	NumChannelsOffset   int = 22
	SampleRateOffset    int = 24
	ByteRateOffset      int = 28
	BlockAlignOffset    int = 32
	BitsPerSampleOffset int = 34
	ChunkSizeOffset     int = 40
	ExpectedHeaderSize  int = 44
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

	// The Data corresponding to BitsPerSample is populated, indexed by sample.
	Data8  [][]uint8
	Data16 [][]int16

	// Data is always populated, indexed by sample. It is a copy of DataXX.
	Data [][]int
}

type StreamedWav struct {
	WavHeader
	Reader io.Reader
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
	// if bLEtoUint32(header, 40) != uint32(len(header) - ExpectedHeaderSize) {
	// 	return errors.New("wav: Header does not specify correct data size")
	// }

	return nil
}

func (wavHeader *WavHeader) setupWithHeaderData(header []byte) (err error) {
	if err = checkHeader(header); err != nil {
		return
	}

	wavHeader.AudioFormat = bLEtoUint16(header, AudioFormatOffset)
	wavHeader.NumChannels = bLEtoUint16(header, NumChannelsOffset)
	wavHeader.SampleRate = bLEtoUint32(header, SampleRateOffset)
	wavHeader.ByteRate = bLEtoUint32(header, ByteRateOffset)
	wavHeader.BlockAlign = bLEtoUint16(header, BlockAlignOffset)
	wavHeader.BitsPerSample = bLEtoUint16(header, BitsPerSampleOffset)
	wavHeader.ChunkSize = bLEtoUint32(header, ChunkSizeOffset)
	wavHeader.NumSamples = int(wavHeader.ChunkSize) / int(wavHeader.BlockAlign)

	return
}

func readSampleFromData(data []byte, sampleIndex int, header WavHeader) (sample []int) {
	sample = make([]int, header.NumChannels)

	for channelIdx := 0; channelIdx < int(header.NumChannels); channelIdx++ {
		if header.BitsPerSample == 8 {
			sample[channelIdx] = int(data[sampleIndex*int(header.NumChannels)+channelIdx])
		} else if header.BitsPerSample == 16 {
			sample[channelIdx] = int(bLEtoInt16(data, sampleIndex*int(header.NumChannels)+channelIdx))
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
	if len(bytes) < ExpectedHeaderSize {
		return nil, errors.New("wav: Unable to read enough data")
	}

	wav = new(Wav)
	err = wav.WavHeader.setupWithHeaderData(bytes)

	data := bytes[ExpectedHeaderSize : int(wav.ChunkSize)+ExpectedHeaderSize]

	wav.Data = make([][]int, wav.NumSamples)

	if wav.BitsPerSample == 8 {
		wav.Data8 = make([][]uint8, wav.NumSamples)

		for i := 0; i < wav.NumSamples; i++ {
			wav.Data8[i] = make([]uint8, wav.NumChannels)
			wav.Data[i] = make([]int, wav.NumChannels)

			sample := readSampleFromData(data, i, wav.WavHeader)

			for ch := 0; ch < int(wav.NumChannels); ch++ {
				wav.Data8[i][ch] = uint8(sample[ch])
				wav.Data[i][ch] = sample[ch]
			}
		}
	} else if wav.BitsPerSample == 16 {
		wav.Data16 = make([][]int16, wav.NumSamples)

		for i := 0; i < wav.NumSamples; i++ {
			wav.Data16[i] = make([]int16, wav.NumChannels)
			wav.Data[i] = make([]int, wav.NumChannels)

			sample := readSampleFromData(data, i, wav.WavHeader)

			for ch := 0; ch < int(wav.NumChannels); ch++ {
				wav.Data16[i][ch] = int16(sample[ch])
				wav.Data[i][ch] = sample[ch]
			}
		}
	}

	return
}

func StreamWav(reader io.Reader) (wav *StreamedWav, err error) {
	if reader == nil {
		return nil, errors.New("wav: Invalid Reader")
	}

	header := make([]byte, ExpectedHeaderSize)
	amountRead, err := reader.Read(header)
	if err != nil {
		return nil, err
	}
	if amountRead < ExpectedHeaderSize {
		return nil, errors.New("wav: Unable to read enough data")
	}

	wav = new(StreamedWav)
	err = wav.setupWithHeaderData(header)
	if err != nil {
		return nil, err
	}
	wav.Reader = reader

	return
}

// Returns an array of [sampleIndex][channelIndex]
func (wav *StreamedWav) ReadSamples(numSamples int) (samples [][]int, err error) {
	data := make([]byte, numSamples*int(wav.BlockAlign))
	amountRead, err := wav.Reader.Read(data)
	if err != nil {
		return
	}
	if amountRead < len(data) {
		err = errors.New("wav: Unable to read enough data")
		return
	}

	samples = make([][]int, numSamples)
	for sampleIndex := 0; sampleIndex < numSamples; sampleIndex++ {
		samples[sampleIndex] = readSampleFromData(data, sampleIndex, wav.WavHeader)
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
func bLEtoUint16(b []byte, idx int) uint16 {
	return uint16(b[idx+1])<<8 + uint16(b[idx])
}

func bLEtoInt16(b []byte, idx int) int16 {
	return int16(b[idx+1])<<8 + int16(b[idx])
}
