# GO-DSP

go-dsp is a digital signal processing package for the [Go programming language](http://golang.org).

## Packages

* **fft** - fast Fourier transform
  * **fft.FFT([]complex128) []complex128** - forward FFT for complex inputs
  * **fft.IFFT([]complex128) []complex128** - inverse FFT for complex inputs
  * **fft.FFT_real([]float64) []complex128** - forward FFT for real inputs
  * **fft.IFFT_real([]float64) []complex128** - inverse FFT for real inputs

## TODO

* All FFT functions:
  * use goroutines

## Authors

**Matt Jibson**

* http://mattjibson.com
* http://github.com/mjibson

## License

Licensed under the BSD license.
