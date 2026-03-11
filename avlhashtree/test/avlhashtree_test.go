package test

import (
	"os"
	"testing"
	"time"

	"github.com/NickOvt/go-chain-trees/test_interfaces"
)

var allResults []test_interfaces.BenchmarkResult

var now time.Time

func appendResult(t *testing.T, options *test_interfaces.BenchmarkOptions) {
	allResults = append(allResults, test_interfaces.TestWithProfile(t, options, now))
}

func TestMain(m *testing.M) {
	// Run all tests
	now = time.Now()
	exitCode := m.Run()

	// After all tests, print combined results
	if len(allResults) > 0 {
		test_interfaces.PrintCombinedResults(allResults)
		test_interfaces.SaveResultsToCSV(now, allResults)
	}

	os.Exit(exitCode)
}

// 1.
// ------------------------------

func TestAVL_10k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 10_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_50k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 50_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_100k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 100_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_200k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 200_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_300k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 300_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_500k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 500_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_700k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 700_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_900k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 900_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_1M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 1_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_2M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 2_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_5M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 5_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_10M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 10_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_50M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 50_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_100M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 100_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_ProofOnly_5M(t *testing.T) {
	appendResult(t, test_interfaces.NewProofOnlyBenchmarkOptions(test_interfaces.AVLHASHTREE, 5_000_000, 0.01))
}

func TestAVL_5MThenAdd1M_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 5_000_000, 1_000_000))
}

func TestAVL_5MThenReinsert1M_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 5_000_000, 1_000_000))
}

func TestAVL_10MThenReinsert2M_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 10_000_000, 2_000_000))
}
