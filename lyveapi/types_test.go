package lyveapi

import (
	"math"
	"testing"
	"time"

	"github.com/szaydel/lyvecloud/lyveapi/monotime"
)

func TestBucketBytesUsed(t *testing.T) {
	t.Parallel()

	const expected = 1.45656e+12
	bucket := Bucket{
		Name:    "alpha-bucket",
		UsageGB: 1456.56,
	}

	if !almostEqual(expected, bucket.BytesUsed(), 0.01) {
		t.Errorf("Usage expected to be ~ %v; got %v",
			expected, bucket.BytesUsed())
	}
}

func almostEqual(expected, actual, epsilon float64) bool {
	return math.Abs(expected-actual) < epsilon
}

func TestTokenExpiresMonoNanos(t *testing.T) {
	t.Parallel()

	const low time.Duration = 12000000000000
	const high time.Duration = 12000000100000

	token := Token{
		Token:         "mock-token-string",
		ExpirationSec: "12000",
	}

	nowMonotonic := monotime.Monotonic()
	e, _ := token.ExpiresMonoNanos()
	if !approx(e-nowMonotonic, low, high) {
		t.Errorf("Value %v not in the range '%v - %v'",
			e-nowMonotonic, low, high)
	}
}

func approx(n, low, high time.Duration) bool {
	if n < low || n > high {
		return false
	}
	return true
}
