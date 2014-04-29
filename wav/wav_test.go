package wav

import (
	"math"
	"os"
	"testing"
)

const (
	smallWav = "small.wav"
	floatWav = "float.wav"
)

func readHeaderData(filePath string) (header []byte, amountRead int, err error) {
	testFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer testFile.Close()

	header = make([]byte, ExpectedHeaderSize)
	amountRead, err = testFile.Read(header)

	return
}

func TestCorrectHeaderValidation(t *testing.T) {
	header, amountRead, err := readHeaderData(smallWav)
	if header == nil {
		t.Fatal("Header data should not be nil")
	}
	if amountRead != ExpectedHeaderSize {
		t.Fatalf("Expected read header size (%d) to match expected header size (%d)", amountRead, ExpectedHeaderSize)
	}

	if err = checkHeader(header); err != nil {
		t.Fatalf("Header validation returned an error: %v", err)
	}
}

func TestShortHeaderValidation(t *testing.T) {
	shortHeader := []byte{0x52, 0x49, 0x46, 0x46, 0x72, 0x8C, 0x34, 0x00, 0x57, 0x41, 0x56, 0x45}
	if err := checkHeader(shortHeader); err == nil {
		t.Fatal("Expected short header to fail validation, but validation passed")
	}
	if err := checkHeader(nil); err == nil {
		t.Fatal("Expected nil header to fail validation, but validation passed")
	}
}

func TestInvalidHeaderValidation(t *testing.T) {
	validateErrorForMissingHeaderValue := func(err error, value string) {
		if err == nil {
			t.Fatalf("Invalid header data missing '%s' should fail validation", value)
		}
	}

	// 44 empty bytes
	invalidHeader := make([]byte, 44)
	err := checkHeader(invalidHeader)
	validateErrorForMissingHeaderValue(err, "RIFF")

	// RIFF and 40 empty bytes
	_ = copy(invalidHeader[RIFFMarkerOffset:4], []byte{0x52, 0x49, 0x46, 0x46})
	err = checkHeader(invalidHeader)
	validateErrorForMissingHeaderValue(err, "WAVE")

	// RIFF, WAVE, and 36 empty bytes
	_ = copy(invalidHeader[8:12], []byte{0x57, 0x41, 0x56, 0x45})
	err = checkHeader(invalidHeader)
	validateErrorForMissingHeaderValue(err, "fmt")

	// RIFF, WAVE, fmt, and 32 empty bytes
	_ = copy(invalidHeader[12:16], []byte{0x66, 0x6D, 0x74, 0x20})
	err = checkHeader(invalidHeader)
	validateErrorForMissingHeaderValue(err, "data")
}

var headers = map[string]map[string]int{
	smallWav: {
		"AudioFormat":   1,
		"NumChannels":   1,
		"SampleRate":    44100,
		"ByteRate":      88200,
		"BlockAlign":    2,
		"BitsPerSample": 16,
		"ChunkSize":     83790,
		"NumSamples":    83790 / 2,
	},
	floatWav: {
		"AudioFormat":   3,
		"NumChannels":   1,
		"SampleRate":    44100,
		"ByteRate":      176400,
		"BlockAlign":    4,
		"BitsPerSample": 32,
		"ChunkSize":     1889280,
		"NumSamples":    1889280 / 4,
	},
}

func testHeader(t *testing.T, header WavHeader, expectedValues map[string]int) {
	expectedAudioFormat := uint16(expectedValues["AudioFormat"])
	if header.AudioFormat != expectedAudioFormat {
		t.Logf("Audio format does not match. Expected: '%v'. Got: '%v'", expectedAudioFormat, header.AudioFormat)
	}

	expectedNumChannels := uint16(expectedValues["NumChannels"])
	if header.NumChannels != expectedNumChannels {
		t.Logf("Number of channles does not match. Expected: '%v'. Got: '%v'", expectedNumChannels, header.NumChannels)
	}

	expectedSampleRate := uint32(expectedValues["SampleRate"])
	if header.SampleRate != expectedSampleRate {
		t.Logf("Sample rate does not match. Expected: '%v'. Got: '%v'", expectedSampleRate, header.SampleRate)
	}

	expectedByteRate := uint32(expectedValues["ByteRate"])
	if header.ByteRate != expectedByteRate {
		t.Logf("Byte rate does not match. Expected: '%v'. Got: '%v'", expectedByteRate, header.ByteRate)
	}

	expectedBlockAlign := uint16(expectedValues["BlockAlign"])
	if header.BlockAlign != expectedBlockAlign {
		t.Logf("Block align does not match. Expected: '%v'. Got: '%v'", expectedBlockAlign, header.BlockAlign)
	}

	expectedBitsPerSample := uint16(expectedValues["BitsPerSample"])
	if header.BitsPerSample != expectedBitsPerSample {
		t.Logf("Bits per sample does not match. Expected: '%v'. Got: '%v'", expectedBitsPerSample, header.BitsPerSample)
	}

	expectedChunkSize := uint32(expectedValues["ChunkSize"])
	if header.ChunkSize != expectedChunkSize {
		t.Logf("Chunk size does not match. Expected: '%v'. Got: '%v'", expectedChunkSize, header.ChunkSize)
	}

	expectedNumSamples := expectedValues["NumSamples"]
	if header.NumSamples != expectedNumSamples {
		t.Logf("Number of samples does not match. Expected: '%v'. Got: '%v'", expectedNumSamples, header.NumSamples)
	}
}

func TestHeaderInitialization(t *testing.T) {
	testFilePath := smallWav
	header, amountRead, err := readHeaderData(testFilePath)
	if err != nil {
		t.Fatalf("Expected header validation to pass, but recevied erro with message '%s'", err.Error())
	}
	if amountRead != ExpectedHeaderSize {
		t.Fatalf("Expected header of valid size but was of size %d", amountRead)
	}

	wav := new(Wav)
	if wav == nil {
		t.Fatal("Unable to create new Wav")
	}

	err = wav.WavHeader.setup(nil)
	if err == nil {
		t.Fatal("Expected error when setting up wav with nil header")
	}

	err = wav.WavHeader.setup(header)
	if err != nil {
		t.Fatalf("Got error when initializing wav with valid header: '%s'", err.Error())
	}
	testHeader(t, wav.WavHeader, headers[testFilePath])
}

func TestReadWavFromNil(t *testing.T) {
	wav, err := ReadWav(nil)
	if wav != nil {
		t.Fatal("Expected ReadWav(nil) to return nil")
	}
	if err == nil {
		t.Fatal("Expected ReadWav(nil) to return an error")
	}
}

func TestReadWavFromFile(t *testing.T) {
	testFilePath := smallWav
	testFile, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("Unable to run test, can't open test file '%s'", testFilePath)
	}
	defer testFile.Close()

	wav, err := ReadWav(testFile)
	if wav == nil {
		t.Fatalf("Error reading wav from file '%s', nil returned", testFilePath)
	}
	if err != nil {
		t.Fatalf("Error reading wav from file '%s': '%s'", testFilePath, err.Error())
	}

	testHeader(t, wav.WavHeader, headers[testFilePath])

	if len(wav.Data8) != 0 {
		t.Fatalf("Expected wav.Data8 to be empty, but has length %d", len(wav.Data8))
	}
	if len(wav.Data16) != wav.NumSamples {
		t.Fatalf("wav.Data16 has incorrect length. Expected %d. Got %d", wav.NumSamples, len(wav.Data16))
	}
	for sampleIndex := 0; sampleIndex < wav.NumSamples; sampleIndex++ {
		if len(wav.Data16[sampleIndex]) != int(wav.NumChannels) {
			t.Fatalf("wav.Data16[%d] has incorrect length. Expected %d. Got %d", sampleIndex, wav.NumChannels, len(wav.Data16[sampleIndex]))
		}
	}
	if len(wav.Data) != wav.NumSamples {
		t.Fatalf("wav.Data has incorrect length. Expected %d. Got %d", wav.NumChannels, len(wav.Data))
	}
	for sampleIndex := 0; sampleIndex < wav.NumSamples; sampleIndex++ {
		if len(wav.Data[sampleIndex]) != int(wav.NumChannels) {
			t.Fatalf("wav.Data[%d] has incorrect length. Expected %d. Got %d", sampleIndex, wav.NumChannels, len(wav.Data[sampleIndex]))
		}
	}
}

func TestStreamWav(t *testing.T) {
	wav, err := StreamWav(nil)
	if wav != nil {
		t.Fatal("Streaming from a nil reader should return a nil wav interface")
	}
	if err == nil {
		t.Fatal("Streaming from a nil reader should return an error")
	}

	testFilePath := smallWav
	testFile, err := os.Open(testFilePath)
	if err != nil {
		t.Fatalf("Unable to run test, can't open test file '%s'", testFilePath)
	}
	defer testFile.Close()

	wav, err = StreamWav(testFile)
	if wav == nil {
		t.Fatal("Streaming from a valid reader returned a nil wav interface")
	}
	if err != nil {
		t.Fatalf("Streaming from a valid reader returned an error", err.Error())
	}

	testHeader(t, wav.WavHeader, headers[testFilePath])

	numberOfSamplesToRead := 1008
	samplesRemaining := wav.NumSamples
	expectedNumberOfSamples := numberOfSamplesToRead
	samples, err := wav.ReadSamples(numberOfSamplesToRead)
	for len(samples) > 0 {
		if err != nil {
			t.Fatalf("ReadSamples returned an unexpected error: %s", err.Error())
		}

		if samples == nil {
			t.Fatal("ReadSamples returned nil")
		}

		if len(samples) != expectedNumberOfSamples {
			t.Fatalf("Number of samples is not as expected. Expected %d. Got %d. %d remaining", expectedNumberOfSamples, len(samples), samplesRemaining)
		}
		for sampleIndex := 0; sampleIndex < len(samples); sampleIndex++ {
			if len(samples[sampleIndex]) != int(wav.NumChannels) {
				t.Fatalf("Number of channels not as expected. Expected %d. Got %d", wav.NumChannels, len(samples[sampleIndex]))
			}
		}
		samplesRemaining -= len(samples)
		expectedNumberOfSamples = int(math.Min(float64(samplesRemaining), float64(numberOfSamplesToRead)))
		samples, err = wav.ReadSamples(numberOfSamplesToRead)
	}

	samples, err = wav.ReadSamples(1)
	if err == nil {
		t.Fatal("Expected error when reading past end of reader")
	}
	if len(samples) != 0 {
		t.Fatal("Expected zero samples returned when reading past end of reader")
	}
}

func TestFloat32(t *testing.T) {
	testFilePath := floatWav
	testFile, err := os.Open(testFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer testFile.Close()

	wav, err := ReadWav(testFile)
	if err != nil {
		t.Fatal(err)
	}

	testHeader(t, wav.WavHeader, headers[testFilePath])

	if len(wav.Data8) != 0 {
		t.Fatalf("Expected wav.Data8 to be empty, but has length %d", len(wav.Data8))
	}
	if len(wav.Data16) != 0 {
		t.Fatalf("Expected wav.Data16 to be empty, but has length %d", len(wav.Data16))
	}
	if len(wav.Float32) != wav.NumSamples {
		t.Fatalf("wav.Float32 has incorrect length. Expected %d. Got %d", wav.NumSamples, len(wav.Float32))
	}
	for sampleIndex := 0; sampleIndex < wav.NumSamples; sampleIndex++ {
		if len(wav.Float32[sampleIndex]) != int(wav.NumChannels) {
			t.Fatalf("wav.Float32[%d] has incorrect length. Expected %d. Got %d", sampleIndex, wav.NumChannels, len(wav.Float32[sampleIndex]))
		}
	}
	if len(wav.Data) != 0 {
		t.Fatalf("wav.Data has incorrect length. Expected 0. Got %d", len(wav.Data))
	}
}
