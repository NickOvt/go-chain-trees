package test_interfaces

import (
	"bytes"
	"testing"
	"time"

	"github.com/NickOvt/go-chain-trees/utils"
)

func TestCalculateTimeBucketsFromDurations_ExactDeciles(t *testing.T) {
	durations := make([]time.Duration, timeBucketCount)
	for idx := range durations {
		durations[idx] = time.Duration(idx+1) * time.Millisecond
	}

	buckets := calculateTimeBucketsFromDurations(durations)
	if len(buckets) != timeBucketCount {
		t.Fatalf("expected %d buckets, got %d", timeBucketCount, len(buckets))
	}

	for idx, bucket := range buckets {
		if bucket.SampleCount != 1 {
			t.Fatalf("bucket %d expected sample count 1, got %d", idx, bucket.SampleCount)
		}
		expected := time.Duration(idx+1) * time.Millisecond
		if bucket.AvgDuration != expected {
			t.Fatalf("bucket %d expected avg %v, got %v", idx, expected, bucket.AvgDuration)
		}
	}
}

func TestCalculateTimeBucketsFromDurations_UnevenDistribution(t *testing.T) {
	durations := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
		6 * time.Millisecond,
		7 * time.Millisecond,
		8 * time.Millisecond,
		9 * time.Millisecond,
		10 * time.Millisecond,
		11 * time.Millisecond,
	}

	buckets := calculateTimeBucketsFromDurations(durations)

	if buckets[0].SampleCount != 2 {
		t.Fatalf("expected first bucket to contain 2 samples, got %d", buckets[0].SampleCount)
	}
	if buckets[0].AvgDuration != 1500*time.Microsecond {
		t.Fatalf("expected first bucket avg 1.5ms, got %v", buckets[0].AvgDuration)
	}

	if buckets[9].SampleCount != 1 {
		t.Fatalf("expected last bucket to contain 1 sample, got %d", buckets[9].SampleCount)
	}
	if buckets[9].AvgDuration != 11*time.Millisecond {
		t.Fatalf("expected last bucket avg 11ms, got %v", buckets[9].AvgDuration)
	}
}

func TestSortPrehashedKeysForTree_AVLUsesByteComparison(t *testing.T) {
	keys := []utils.Hash{
		{0x02, 0x00},
		{0x01, 0xff},
		{0x01, 0x00},
	}

	if err := sortPrehashedKeysForTree(AVLHASHTREE, keys); err != nil {
		t.Fatalf("sortPrehashedKeysForTree returned error: %v", err)
	}

	want := []utils.Hash{
		{0x01, 0x00},
		{0x01, 0xff},
		{0x02, 0x00},
	}

	for idx := range want {
		if !bytes.Equal(keys[idx], want[idx]) {
			t.Fatalf("unexpected AVL order at %d: got %x want %x", idx, keys[idx], want[idx])
		}
	}
}

func TestSortPrehashedKeysForTree_SMTUsesLSBPathOrder(t *testing.T) {
	keys := []utils.Hash{
		{0x01},
		{0x80},
		{0x02},
		{0x40},
	}

	if err := sortPrehashedKeysForTree(SMT, keys); err != nil {
		t.Fatalf("sortPrehashedKeysForTree returned error: %v", err)
	}

	want := []utils.Hash{
		{0x80},
		{0x40},
		{0x02},
		{0x01},
	}

	for idx := range want {
		if !bytes.Equal(keys[idx], want[idx]) {
			t.Fatalf("unexpected SMT order at %d: got %x want %x", idx, keys[idx], want[idx])
		}
	}
}
