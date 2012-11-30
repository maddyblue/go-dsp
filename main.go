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

			for i := MAX_N; i > 0; i >>= 1 {
				fft.MP_MIN_BLOCKSIZE = i

				fft.MP_METHOD = fft.MP_METHOD_NORMAL
				r = testing.Benchmark(BenchmarkFFT)
				fmt.Printf("%20s %12d %2d %10d %12d\n", "normal", MAX_N, numprocs, i, r.NsPerOp())

				fft.MP_METHOD = fft.MP_METHOD_WAIT_GROUP
				r = testing.Benchmark(BenchmarkFFT)
				fmt.Printf("%20s %12d %2d %10d %12d\n", "waitgroups", MAX_N, numprocs, i, r.NsPerOp())

				fft.MP_METHOD = fft.MP_METHOD_WORKER_POOLS
				for j := 1; j <= n * 2; j++ {
					fft.WORKER_POOLS_COUNT = j
					r = testing.Benchmark(BenchmarkFFT)
					fmt.Printf("%17s-%02d %12d %2d %10d %12d\n", "workerpools", j, MAX_N, numprocs, i, r.NsPerOp())
				}
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
