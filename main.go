package main

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/mjibson/go-dsp/fft"
)

var (
	numprocs int
	MAX_N int
)

func main() {
	var r testing.BenchmarkResult

	for MAX_N = 1 << 20; MAX_N > 0; MAX_N >>= 1 {
		for n := 1; n <= runtime.NumCPU(); n++ {
			numprocs = n

			for j := 1; j <= n * 2; j++ {
				fft.WORKER_POOL_SIZE = j
				r = testing.Benchmark(BenchmarkFFT)
				fmt.Printf("WPS: %02d N: %12d procs: %2d ns: %12d\n", j, MAX_N, numprocs, r.NsPerOp())
			}
		}
	}
}

func BenchmarkFFT(b *testing.B) {
	b.StopTimer()

	runtime.GOMAXPROCS(numprocs)

	N := MAX_N
	a := make([]complex128, N)
	for i := 0; i < N; i++ {
		a[i] = complex(float64(i)/float64(N), 0)
	}

	fft.EnsureRadix2Factors(N)

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		fft.FFT(a)
	}
}
