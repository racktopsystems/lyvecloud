package lyveapi

import (
	"math"
	"testing"
)

func TestBucketUsageInBytes(t *testing.T) {
	t.Parallel()

	const expected = 1.45656e+12
	bucket := Bucket{
		Name:    "alpha-bucket",
		UsageGB: 1456.56,
	}

	t.Log(almostEqual(expected, bucket.UsageInBytes(), 0.01))
}

func almostEqual(expected, actual, epsilon float64) bool {
	return math.Abs(expected-actual) < epsilon
}
