package main

import (
	"fmt"
	"math"
	"time"
)

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; float64(i) <= math.Sqrt(float64(n)); i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

func calculatePrimes(maxNumber int) []int {
	var primes []int
	for number := 2; number < maxNumber; number++ {
		if isPrime(number) {
			primes = append(primes, number)
		}
	}
	return primes
}

func main() {
	start := time.Now()
	primes := calculatePrimes(50000) // Find primes up to 50,000
	elapsed := time.Since(start)

	fmt.Println("Number of primes found:", len(primes))
	fmt.Printf("Time taken in Go: %s\n", elapsed)
}
