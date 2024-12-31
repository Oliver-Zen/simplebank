package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Define a package-level rand.Rand instance
// `rand.NewSource()` creates a new random number generator.
// `time.Now().UnixNano()` creates a seed by current Unix timestamp in nanoseconds, ensuring randomness for each run.
// `rand.New()` creates a new *rand.Rand instance with the specified source.
// `seededRand` is a pointer that can call methods of [rand.Rand].
// Because Go automatically dereferences the pointer when calling a method.
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomInt generates a random integer between min and max.
func RandomInt(min, max int64) int64 {
	return min + seededRand.Int63n(max-min+1)
}

// RandomString generates a random string of length n.
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[seededRand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner generates a random owner name.
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney generates a random amount of money.
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency generates a random currency code.
func RandomCurrency() string {
	currencies := []string{EUR, USD, CAD}
	n := len(currencies)
	return currencies[seededRand.Intn(n)]
}

// RandomEmail generates a random email.
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
