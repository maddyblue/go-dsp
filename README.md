# GO-DSP

go-dsp is a digital signal processing package for the [Go programming language](http://golang.org).

## Packages

* **fft** - fast Fourier transform
  * docs at: http://gopkgdoc.appspot.com/pkg/github.com/mjibson/go-dsp/fft

## Installation

```$ goinstall "github.com/mjibson/go-dsp/fft"```

```
package main

import "github.com/mjibson/go-dsp/fft"
import "fmt"

func main() {
        fmt.Println(fft.FFTReal([]float64 {1, 2, 3}))
}
```

## TODO

* fft:
  * N-dimensional FFT functions
  * support float32/complex64 inputs
  * research possible performance gains with goroutines

## Authors

**Matt Jibson**

* http://mattjibson.com
* http://github.com/mjibson

## License

Licensed under the BSD license.
