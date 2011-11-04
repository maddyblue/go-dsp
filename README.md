# GO-DSP

go-dsp is a digital signal processing package for the [Go programming language](http://golang.org).

## In-Progress Packages and Functions

These methods are working, but have known bugs.

* **fft** - fast Fourier transform
  * **fft.Fft([]float64) []complex128** - forward FFT
    * todo:
      * use goroutines
      * only returns correct results sometimes
      * support lengths that are not a power of 2

## Planned Packages and Functions

* **fft.Ifft** - inverse FFT

## Authors

**Matt Jibson**

* http://mattjibson.com
* http://github.com/mjibson

## License

Licensed under the BSD license.
