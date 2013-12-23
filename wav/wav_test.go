package wav

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

const (
	SmallWavFileName = "test_files/small.wav"
)

func openWavFile(filePath string) (wavFile *os.File) {
	wavFile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	return
}

func readHeaderData(filePath string) (header []byte, amountRead int, err error) {
	testFile := openWavFile(filePath)
	defer testFile.Close()

	header = make([]byte, ExpectedHeaderSize)
	amountRead, err = testFile.Read(header)

	return
}

func TestHeaderValidation(t *testing.T) {
	Convey("Correct WAV should pass header validation", t, func() {
		header, amountRead, err := readHeaderData(SmallWavFileName)

		Convey("Should be able to read header data from test file", func() {
			So(err, ShouldBeNil)
			So(amountRead, ShouldEqual, ExpectedHeaderSize)
		})

		So(checkHeader(header), ShouldBeNil)
	})

	Convey("Short header data should fail validation", t, func() {
		shortHeader := []byte{0x52, 0x49, 0x46, 0x46, 0x72, 0x8C, 0x34, 0x00, 0x57, 0x41, 0x56, 0x45}
		So(checkHeader(shortHeader), ShouldNotBeNil)
		So(checkHeader(nil), ShouldNotBeNil)
	})

	// 44 empty bytes
	invalidHeader := make([]byte, 44)
	Convey("Invalid header data missing 'RIFF' should fail validation", t, func() {
		err := checkHeader(invalidHeader)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "RIFF")
	})

	// RIFF and 40 empty bytes
	_ = copy(invalidHeader[0:4], []byte{0x52, 0x49, 0x46, 0x46})
	Convey("Invalid header data missing 'WAVE' should fail validation", t, func() {
		err := checkHeader(invalidHeader)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "WAVE")
	})

	// RIFF, WAVE, and 36 empty bytes
	_ = copy(invalidHeader[8:12], []byte{0x57, 0x41, 0x56, 0x45})
	Convey("Invalid header data missing 'fmt' should fail validation", t, func() {
		err := checkHeader(invalidHeader)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "fmt")
	})

	// RIFF, WAVE, fmt, and 32 empty bytes
	_ = copy(invalidHeader[12:16], []byte{0x66, 0x6D, 0x74, 0x20})
	Convey("Invalid header data missing 'data' should fail validation", t, func() {
		err := checkHeader(invalidHeader)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "data")
	})
}

func expectedHeaderDataForTestFile(filePath string) map[string]int {
	switch filePath {
	case SmallWavFileName:
		return map[string]int{
			"AudioFormat":   1,
			"NumChannels":   1,
			"SampleRate":    44100,
			"ByteRate":      88200,
			"BlockAlign":    2,
			"BitsPerSample": 16,
			"ChunkSize":     83790,
			"NumSamples":    83790 / 2,
		}
	}

	return map[string]int{}
}

func performTestOfHeaderInitialization(header WavHeader, expectedValues map[string]int) {
	So(header.AudioFormat, ShouldEqual, expectedValues["AudioFormat"])
	So(header.NumChannels, ShouldEqual, expectedValues["NumChannels"])
	So(header.SampleRate, ShouldEqual, expectedValues["SampleRate"])
	So(header.ByteRate, ShouldEqual, expectedValues["ByteRate"])
	So(header.BlockAlign, ShouldEqual, expectedValues["BlockAlign"])
	So(header.BitsPerSample, ShouldEqual, expectedValues["BitsPerSample"])
	So(header.ChunkSize, ShouldEqual, expectedValues["ChunkSize"])
	So(header.NumSamples, ShouldEqual, expectedValues["NumSamples"])
}

func TestHeaderInitialization(t *testing.T) {
	Convey("Header should be initialized with correct values from valid WAV file", t, func() {
		testFilePath := SmallWavFileName
		header, amountRead, err := readHeaderData(testFilePath)

		Convey("Should be able to read header data from test file", func() {
			So(err, ShouldBeNil)
			So(amountRead, ShouldEqual, ExpectedHeaderSize)
		})

		wav := new(Wav)
		So(wav, ShouldNotBeNil)

		err = wav.WavHeader.setupWithHeaderData(nil)
		So(err, ShouldNotBeNil)

		err = wav.WavHeader.setupWithHeaderData(header)
		So(err, ShouldBeNil)
		performTestOfHeaderInitialization(wav.WavHeader, expectedHeaderDataForTestFile(testFilePath))
	})
}

func TestReadWav(t *testing.T) {
	Convey("Reading from an invalid Reader should fail", t, func() {
		wav, err := ReadWav(nil)
		So(wav, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("Should be able to read a full file", t, func() {
		testFilePath := SmallWavFileName
		testFile := openWavFile(testFilePath)
		defer testFile.Close()
		So(testFile, ShouldNotBeNil)

		wav, err := ReadWav(testFile)
		So(wav, ShouldNotBeNil)
		So(err, ShouldBeNil)
		performTestOfHeaderInitialization(wav.WavHeader, expectedHeaderDataForTestFile(testFilePath))
		So(len(wav.Data8), ShouldEqual, 0)
		So(len(wav.Data16), ShouldEqual, wav.NumSamples)
		So(len(wav.Data), ShouldEqual, wav.NumSamples)
	})
}

func TestStreamWav(t *testing.T) {
	Convey("Streaming from an invalid Reader should fail", t, func() {
		wav, err := ReadWav(nil)
		So(wav, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("Should be able to stream a file", t, func() {
		testFilePath := SmallWavFileName
		testFile := openWavFile(testFilePath)
		defer testFile.Close()
		So(testFile, ShouldNotBeNil)

		wav, err := StreamWav(testFile)
		So(wav, ShouldNotBeNil)
		So(err, ShouldBeNil)
		performTestOfHeaderInitialization(wav.WavHeader, expectedHeaderDataForTestFile(testFilePath))
		samplesRemaining := wav.NumSamples
		for samplesRemaining > 0 {
			numSamples := 10000
			if samplesRemaining < numSamples {
				numSamples = samplesRemaining
			}
			samplesRemaining -= numSamples

			samples, err := wav.ReadSamples(numSamples)
			So(err, ShouldBeNil)
			So(samples, ShouldNotBeNil)
			So(len(samples), ShouldEqual, numSamples)
			So(len(samples[0]), ShouldEqual, wav.NumChannels)
		}
		samples, err := wav.ReadSamples(1)
		So(err, ShouldNotBeNil)
		So(len(samples), ShouldEqual, 0)
	})
}
