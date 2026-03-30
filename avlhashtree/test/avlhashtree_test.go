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

func newInsertOnlyBenchmarkOptions(elementCount int, scenarioName string) *test_interfaces.BenchmarkOptions {
	return &test_interfaces.BenchmarkOptions{
		TreeType:       test_interfaces.AVLHASHTREE,
		ScenarioName:   scenarioName,
		CPUProfile:     true,
		ElementCount:   elementCount,
		SampleSize:     0.01,
		MeasureInserts: true,
		BlockSizeBytes: 32,
		DataSizeBytes:  8,
	}
}

func newOrderedPrehashedBenchmarkOptions(elementCount int, scenarioName string) *test_interfaces.BenchmarkOptions {
	return &test_interfaces.BenchmarkOptions{
		TreeType:                test_interfaces.AVLHASHTREE,
		ScenarioName:            scenarioName,
		CPUProfile:              true,
		ElementCount:            elementCount,
		SampleSize:              0.01,
		MeasureInserts:          true,
		IncludeInclusionProof:   true,
		BlockSizeBytes:          32,
		DataSizeBytes:           8,
		UseOrderedPrehashedKeys: true,
	}
}

func newProofOnlyBenchmarkOptions(buildCount int, sampleSize float32, scenarioName string) *test_interfaces.BenchmarkOptions {
	options := test_interfaces.NewProofOnlyBenchmarkOptions(test_interfaces.AVLHASHTREE, buildCount, sampleSize)
	options.ScenarioName = scenarioName
	return options
}

func newOrderedPrehashedProofOnlyBenchmarkOptions(buildCount int, sampleSize float32, scenarioName string) *test_interfaces.BenchmarkOptions {
	options := test_interfaces.NewProofOnlyBenchmarkOptions(test_interfaces.AVLHASHTREE, buildCount, sampleSize)
	options.ScenarioName = scenarioName
	options.UseOrderedPrehashedKeys = true
	return options
}

func newOrderedPrehashedExistingKeyInsertBenchmarkOptions(prebuildCount int, insertCount int, scenarioName string) *test_interfaces.BenchmarkOptions {
	options := test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, prebuildCount, insertCount)
	options.ScenarioName = scenarioName
	options.UseOrderedPrehashedKeys = true
	return options
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

func TestAVL_PrehashedOrdered_1M(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(1_000_000, "prehashed_ordered_build_1m"))
}

func TestAVL_2M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 2_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_5M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 5_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_PrehashedOrdered_5M(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(5_000_000, "prehashed_ordered_build_5m"))
}

func TestAVL_10M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 10_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestAVL_25M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 25_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

//func TestAVL_50M(t *testing.T) {
//	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 50_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
//}
//
//func TestAVL_100M(t *testing.T) {
//	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.AVLHASHTREE, CPUProfile: true, ElementCount: 100_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
//}

func TestAVL_InsertOnly_10k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(10_000, "insert_only_build_10k"))
}

func TestAVL_InsertOnly_50k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(50_000, "insert_only_build_50k"))
}

func TestAVL_InsertOnly_100k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(100_000, "insert_only_build_100k"))
}

func TestAVL_InsertOnly_200k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(200_000, "insert_only_build_200k"))
}

func TestAVL_InsertOnly_300k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(300_000, "insert_only_build_300k"))
}

func TestAVL_InsertOnly_500k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(500_000, "insert_only_build_500k"))
}

func TestAVL_InsertOnly_700k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(700_000, "insert_only_build_700k"))
}

func TestAVL_InsertOnly_900k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(900_000, "insert_only_build_900k"))
}

func TestAVL_InsertOnly_1M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(1_000_000, "insert_only_build_1m"))
}

func TestAVL_InsertOnly_2M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(2_000_000, "insert_only_build_2m"))
}

func TestAVL_InsertOnly_5M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(5_000_000, "insert_only_build_5m"))
}

func TestAVL_InsertOnly_10M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(10_000_000, "insert_only_build_10m"))
}

func TestAVL_InsertOnly_25M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(25_000_000, "insert_only_build_25m"))
}

func TestAVL_ProofOnly_1M_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(1_000_000, 0.05, "proof_only_after_1m_build_sample_5pct"))
}

func TestAVL_PrehashedOrdered_ProofOnly_1M_Sample5Pct(t *testing.T) {
	appendResult(t, newOrderedPrehashedProofOnlyBenchmarkOptions(1_000_000, 0.05, "prehashed_ordered_proof_only_after_1m_build_sample_5pct"))
}

func TestAVL_ProofOnly_5M_Sample3Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(5_000_000, 0.03, "proof_only_after_5m_build_sample_3pct"))
}

func TestAVL_PrehashedOrdered_ProofOnly_5M_Sample3Pct(t *testing.T) {
	appendResult(t, newOrderedPrehashedProofOnlyBenchmarkOptions(5_000_000, 0.03, "prehashed_ordered_proof_only_after_5m_build_sample_3pct"))
}

func TestAVL_ProofOnly_10M_Sample2Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(10_000_000, 0.02, "proof_only_after_10m_build_sample_2pct"))
}

func TestAVL_1MThenAdd200k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 1_000_000, 200_000))
}

func TestAVL_2MThenAdd400k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 2_000_000, 400_000))
}

func TestAVL_5MThenAdd1M_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 5_000_000, 1_000_000))
}

func TestAVL_10MThenAdd2M_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 10_000_000, 2_000_000))
}

func TestAVL_1MThenReinsert200k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 1_000_000, 200_000))
}

func TestAVL_PrehashedOrdered_1MThenReinsert200k_ExistingElements(t *testing.T) {
	appendResult(t, newOrderedPrehashedExistingKeyInsertBenchmarkOptions(1_000_000, 200_000, "prehashed_ordered_build_1m_then_reinsert_200k_existing"))
}

func TestAVL_2MThenReinsert400k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 2_000_000, 400_000))
}

func TestAVL_5MThenReinsert1M_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 5_000_000, 1_000_000))
}

func TestAVL_PrehashedOrdered_5MThenReinsert1M_ExistingElements(t *testing.T) {
	appendResult(t, newOrderedPrehashedExistingKeyInsertBenchmarkOptions(5_000_000, 1_000_000, "prehashed_ordered_build_5m_then_reinsert_1m_existing"))
}

func TestAVL_10MThenReinsert2M_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.AVLHASHTREE, 10_000_000, 2_000_000))
}
