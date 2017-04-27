package filter

func lfilter(b []float64, a []float64, x []float64, y []float64) []float64 {

	K := len(b)

	var xn float64

	pz := make([]float64, K)
	z := make([]float64, K)

	for n, yn := range y {
		xn = x[n]
		y[n] = b[0]*xn + pz[0]
		for k := 0; k < K-1; k++ {
			z[k] = b[k+1]*xn + pz[k+1] - a[k+1]*yn
		}
		z[K-1] = b[K]*xn - a[K]*yn
		pz, z = z, pz
	}

	return pz
}

func revlfilter(b []float64, a []float64, x []float64, y []float64) []float64 {

	K := len(b)

	var xn float64
	var yn float64

	pz := make([]float64, K)
	z := make([]float64, K)

	for n := len(y) - 1; n >= 0; n-- {
		yn = y[n]
		xn = x[n]
		y[n] = b[0]*xn + pz[0]
		for k := 0; k < K-1; k++ {
			z[k] = b[k+1]*xn + pz[k+1] - a[k+1]*yn
		}
		z[K-1] = b[K]*xn - a[K]*yn

		pz, z = z, pz
	}

	return pz
}
