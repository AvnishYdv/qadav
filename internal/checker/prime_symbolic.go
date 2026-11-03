package checker

// Prime-symbolic knowledge engine helper functions for semantic analysis and compile-time evaluation.

const PrimeTableSize = 10000

var primeTable []int64

func init() {
	primeTable = sievePrimes(PrimeTableSize)
}

func sievePrimes(count int) []int64 {
	limit := 150000
	is := make([]bool, limit+1)
	for i := range is {
		is[i] = true
	}
	is[0], is[1] = false, false
	for p := 2; p*p <= limit; p++ {
		if is[p] {
			for j := p * p; j <= limit; j += p {
				is[j] = false
			}
		}
	}
	var res []int64
	for i := 2; i <= limit && len(res) < count; i++ {
		if is[i] {
			res = append(res, int64(i))
		}
	}
	return res
}

// IsPrime checks whether n is prime.
func IsPrime(n int64) bool {
	if n < 2 {
		return false
	}
	for _, p := range primeTable {
		if p*p > n {
			break
		}
		if n%p == 0 {
			return n == p
		}
	}
	return true
}

// Factorize computes the prime factors of n.
func Factorize(n int64) []int64 {
	if n <= 1 {
		return nil
	}
	var factors []int64
	curr := n
	for _, p := range primeTable {
		if p*p > curr {
			break
		}
		for curr%p == 0 {
			factors = append(factors, p)
			curr /= p
		}
	}
	if curr > 1 {
		factors = append(factors, curr)
	}
	return factors
}

// EncodeConcept composes an array of prime factors by multiplying them.
func EncodeConcept(primes []int64) int64 {
	if len(primes) == 0 {
		return 1
	}
	var prod int64 = 1
	for _, p := range primes {
		prod *= p
	}
	return prod
}

// DecodeConcept decomposes a concept integer into its prime factors.
func DecodeConcept(n int64) []int64 {
	return Factorize(n)
}
