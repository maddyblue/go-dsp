# GO-DSP

go-dsp is a digital signal processing package for the [Go programming language](http://golang.org).

## Packages

* **[dsputils](http://gopkgdoc.appspot.com/pkg/github.com/mjibson/go-dsp/dsputils)** - utilities and data structures for DSP
* **[fft](http://gopkgdoc.appspot.com/pkg/github.com/mjibson/go-dsp/fft)** - fast Fourier transform
* **[window](http://gopkgdoc.appspot.com/pkg/github.com/mjibson/go-dsp/window)** - window functions

## Installation

```$ go get "github.com/mjibson/go-dsp/fft"```

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
  * support float32/complex64 inputs
  * research possible performance gains with goroutines

## License

ISC/BSD-style license.
