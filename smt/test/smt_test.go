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
		TreeType:       test_interfaces.SMT,
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
		TreeType:                test_interfaces.SMT,
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
	options := test_interfaces.NewProofOnlyBenchmarkOptions(test_interfaces.SMT, buildCount, sampleSize)
	options.ScenarioName = scenarioName
	return options
}

func newOrderedPrehashedProofOnlyBenchmarkOptions(buildCount int, sampleSize float32, scenarioName string) *test_interfaces.BenchmarkOptions {
	options := test_interfaces.NewProofOnlyBenchmarkOptions(test_interfaces.SMT, buildCount, sampleSize)
	options.ScenarioName = scenarioName
	options.UseOrderedPrehashedKeys = true
	return options
}

func newOrderedPrehashedExistingKeyInsertBenchmarkOptions(prebuildCount int, insertCount int, scenarioName string) *test_interfaces.BenchmarkOptions {
	options := test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, prebuildCount, insertCount)
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
// func TestSMT_10(t *testing.T) {
// 	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 10, SampleSize: 1, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
// }

func TestSMT_10k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 10_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_50k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 50_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_100k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 100_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_200k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 200_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_300k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 300_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_500k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 500_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_700k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 700_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_900k(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 900_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_1M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 1_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_2M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 2_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_5M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 5_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_10M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 10_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_25M(t *testing.T) {
	appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 25_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
}

func TestSMT_PrehashedOrdered_10k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(10_000, "prehashed_ordered_build_10k"))
}

func TestSMT_PrehashedOrdered_50k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(50_000, "prehashed_ordered_build_50k"))
}

func TestSMT_PrehashedOrdered_100k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(100_000, "prehashed_ordered_build_100k"))
}

func TestSMT_PrehashedOrdered_200k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(200_000, "prehashed_ordered_build_200k"))
}

func TestSMT_PrehashedOrdered_300k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(300_000, "prehashed_ordered_build_300k"))
}

func TestSMT_PrehashedOrdered_500k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(500_000, "prehashed_ordered_build_500k"))
}

func TestSMT_PrehashedOrdered_700k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(700_000, "prehashed_ordered_build_700k"))
}

func TestSMT_PrehashedOrdered_900k(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(900_000, "prehashed_ordered_build_900k"))
}

func TestSMT_PrehashedOrdered_1M(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(1_000_000, "prehashed_ordered_build_1m"))
}

func TestSMT_PrehashedOrdered_2M(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(2_000_000, "prehashed_ordered_build_2m"))
}

func TestSMT_PrehashedOrdered_5M(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(5_000_000, "prehashed_ordered_build_5m"))
}

func TestSMT_PrehashedOrdered_10M(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(10_000_000, "prehashed_ordered_build_10m"))
}

func TestSMT_PrehashedOrdered_25M(t *testing.T) {
	appendResult(t, newOrderedPrehashedBenchmarkOptions(25_000_000, "prehashed_ordered_build_25m"))
}

//	func TestSMT_50M(t *testing.T) {
//		appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 50_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
//	}
//
//	func TestSMT_100M(t *testing.T) {
//		appendResult(t, &test_interfaces.BenchmarkOptions{TreeType: test_interfaces.SMT, CPUProfile: true, ElementCount: 100_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 32, DataSizeBytes: 8})
//	}

func TestSMT_InsertOnly_10k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(10_000, "insert_only_build_10k"))
}

func TestSMT_InsertOnly_50k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(50_000, "insert_only_build_50k"))
}

func TestSMT_InsertOnly_100k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(100_000, "insert_only_build_100k"))
}

func TestSMT_InsertOnly_200k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(200_000, "insert_only_build_200k"))
}

func TestSMT_InsertOnly_300k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(300_000, "insert_only_build_300k"))
}

func TestSMT_InsertOnly_500k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(500_000, "insert_only_build_500k"))
}

func TestSMT_InsertOnly_700k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(700_000, "insert_only_build_700k"))
}

func TestSMT_InsertOnly_900k(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(900_000, "insert_only_build_900k"))
}

func TestSMT_InsertOnly_1M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(1_000_000, "insert_only_build_1m"))
}

func TestSMT_InsertOnly_2M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(2_000_000, "insert_only_build_2m"))
}

func TestSMT_InsertOnly_5M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(5_000_000, "insert_only_build_5m"))
}

func TestSMT_InsertOnly_10M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(10_000_000, "insert_only_build_10m"))
}

func TestSMT_InsertOnly_25M(t *testing.T) {
	appendResult(t, newInsertOnlyBenchmarkOptions(25_000_000, "insert_only_build_25m"))
}

func TestSMT_ProofOnly_10k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(10_000, 0.05, "proof_only_after_10k_build_sample_5pct"))
}

func TestSMT_ProofOnly_50k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(50_000, 0.05, "proof_only_after_50k_build_sample_5pct"))
}

func TestSMT_ProofOnly_100k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(100_000, 0.05, "proof_only_after_100k_build_sample_5pct"))
}

func TestSMT_ProofOnly_200k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(200_000, 0.05, "proof_only_after_200k_build_sample_5pct"))
}

func TestSMT_ProofOnly_300k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(300_000, 0.05, "proof_only_after_300k_build_sample_5pct"))
}

func TestSMT_ProofOnly_500k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(500_000, 0.05, "proof_only_after_500k_build_sample_5pct"))
}

func TestSMT_ProofOnly_700k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(700_000, 0.05, "proof_only_after_700k_build_sample_5pct"))
}

func TestSMT_ProofOnly_900k_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(900_000, 0.05, "proof_only_after_900k_build_sample_5pct"))
}

func TestSMT_ProofOnly_1M_Sample5Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(1_000_000, 0.05, "proof_only_after_1m_build_sample_5pct"))
}

func TestSMT_PrehashedOrdered_ProofOnly_1M_Sample5Pct(t *testing.T) {
	appendResult(t, newOrderedPrehashedProofOnlyBenchmarkOptions(1_000_000, 0.05, "prehashed_ordered_proof_only_after_1m_build_sample_5pct"))
}

func TestSMT_ProofOnly_2M_Sample3Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(2_000_000, 0.03, "proof_only_after_2m_build_sample_3pct"))
}

func TestSMT_ProofOnly_5M_Sample3Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(5_000_000, 0.03, "proof_only_after_5m_build_sample_3pct"))
}

func TestSMT_PrehashedOrdered_ProofOnly_5M_Sample3Pct(t *testing.T) {
	appendResult(t, newOrderedPrehashedProofOnlyBenchmarkOptions(5_000_000, 0.03, "prehashed_ordered_proof_only_after_5m_build_sample_3pct"))
}

func TestSMT_ProofOnly_10M_Sample2Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(10_000_000, 0.02, "proof_only_after_10m_build_sample_2pct"))
}

func TestSMT_ProofOnly_25M_Sample2Pct(t *testing.T) {
	appendResult(t, newProofOnlyBenchmarkOptions(25_000_000, 0.02, "proof_only_after_25m_build_sample_2pct"))
}

func TestSMT_10kThenAdd2k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 10_000, 2_000))
}

func TestSMT_50kThenAdd10k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 50_000, 10_000))
}

func TestSMT_100kThenAdd20k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 100_000, 20_000))
}

func TestSMT_200kThenAdd40k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 200_000, 40_000))
}

func TestSMT_300kThenAdd60k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 300_000, 60_000))
}

func TestSMT_500kThenAdd100k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 500_000, 100_000))
}

func TestSMT_700kThenAdd140k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 700_000, 140_000))
}

func TestSMT_900kThenAdd180k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 900_000, 180_000))
}

func TestSMT_1MThenAdd200k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 1_000_000, 200_000))
}

func TestSMT_2MThenAdd400k_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 2_000_000, 400_000))
}

func TestSMT_5MThenAdd1M_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 5_000_000, 1_000_000))
}

func TestSMT_10MThenAdd2M_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 10_000_000, 2_000_000))
}

func TestSMT_25MThenAdd5M_NewElements(t *testing.T) {
	appendResult(t, test_interfaces.NewPostBuildInsertBenchmarkOptions(test_interfaces.SMT, 25_000_000, 5_000_000))
}

func TestSMT_10kThenReinsert2k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 10_000, 2_000))
}

func TestSMT_50kThenReinsert10k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 50_000, 10_000))
}

func TestSMT_100kThenReinsert20k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 100_000, 20_000))
}

func TestSMT_200kThenReinsert40k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 200_000, 40_000))
}

func TestSMT_300kThenReinsert60k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 300_000, 60_000))
}

func TestSMT_500kThenReinsert100k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 500_000, 100_000))
}

func TestSMT_700kThenReinsert140k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 700_000, 140_000))
}

func TestSMT_900kThenReinsert180k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 900_000, 180_000))
}

func TestSMT_1MThenReinsert200k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 1_000_000, 200_000))
}

func TestSMT_PrehashedOrdered_1MThenReinsert200k_ExistingElements(t *testing.T) {
	appendResult(t, newOrderedPrehashedExistingKeyInsertBenchmarkOptions(1_000_000, 200_000, "prehashed_ordered_build_1m_then_reinsert_200k_existing"))
}

func TestSMT_2MThenReinsert400k_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 2_000_000, 400_000))
}

func TestSMT_5MThenReinsert1M_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 5_000_000, 1_000_000))
}

func TestSMT_PrehashedOrdered_5MThenReinsert1M_ExistingElements(t *testing.T) {
	appendResult(t, newOrderedPrehashedExistingKeyInsertBenchmarkOptions(5_000_000, 1_000_000, "prehashed_ordered_build_5m_then_reinsert_1m_existing"))
}

func TestSMT_10MThenReinsert2M_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 10_000_000, 2_000_000))
}

func TestSMT_25MThenReinsert5M_ExistingElements(t *testing.T) {
	appendResult(t, test_interfaces.NewExistingKeyInsertBenchmarkOptions(test_interfaces.SMT, 25_000_000, 5_000_000))
}
