//package test
//
//import (
//	"crypto/rand"
//	randMath "math/rand"
//	"runtime"
//	"testing"
//	"time"
//
//	"github.com/NickOvt/go-chain-trees/avlhashtree"
//	"github.com/NickOvt/go-chain-trees/utils"
//)
//
//type BenchmarkResult struct {
//	ElementCount      int
//	InsertionTime     time.Duration
//	AvgPerBlock       time.Duration
//	ProofGenTime      time.Duration
//	ProofSize         int
//	ProofVerifyTime   time.Duration
//	MemoryAllocatedMB float64
//	TotalAllocatedMB  float64
//	HeapObjects       uint64
//}
//
//func TestAVLHashTreeBenchmarkSuite(t *testing.T) {
//	testSizes := []int{10_000, 50_000, 100_000, 200_000, 300_000, 500_000, 700_000, 900_000, 1_000_000}
//
//	results := make([]BenchmarkResult, 0, len(testSizes))
//
//	for _, count := range testSizes {
//		t.Logf("\n========================================")
//		t.Logf("Starting test with %d elements", count)
//		t.Logf("========================================")
//
//		result := runBenchmark(t, count)
//		results = append(results, result)
//
//		// Force GC between tests to get cleaner memory measurements
//		runtime.GC()
//		time.Sleep(100 * time.Millisecond)
//	}
//
//	// Log all results in a structured format
//	t.Logf("\n\n========================================")
//	t.Logf("BENCHMARK SUITE RESULTS")
//	t.Logf("========================================\n")
//
//	t.Logf("%-12s %-15s %-15s %-15s %-12s %-15s %-15s %-15s %-12s",
//		"Elements", "Insert Time", "Avg/Block", "Proof Gen", "Proof Size", "Proof Verify", "Mem Alloc MB", "Total Alloc MB", "Heap Objs")
//	t.Logf("%-12s %-15s %-15s %-15s %-12s %-15s %-15s %-15s %-12s",
//		"------------", "---------------", "---------------", "---------------", "------------", "---------------", "---------------", "---------------", "------------")
//
//	for _, r := range results {
//		t.Logf("%-12d %-15v %-15v %-15v %-12d %-15v %-15.2f %-15.2f %-12d",
//			r.ElementCount,
//			r.InsertionTime,
//			r.AvgPerBlock,
//			r.ProofGenTime,
//			r.ProofSize,
//			r.ProofVerifyTime,
//			r.MemoryAllocatedMB,
//			r.TotalAllocatedMB,
//			r.HeapObjects)
//	}
//
//	// Also log in a format easy to copy to CSV/spreadsheet
//	t.Logf("\n\nCSV Format:")
//	t.Logf("Elements,InsertTime(ns),AvgPerBlock(ns),ProofGen(ns),ProofSize(bytes),ProofVerify(ns),MemAllocMB,TotalAllocMB,HeapObjects")
//	for _, r := range results {
//		t.Logf("%d,%d,%d,%d,%d,%d,%.2f,%.2f,%d",
//			r.ElementCount,
//			r.InsertionTime.Nanoseconds(),
//			r.AvgPerBlock.Nanoseconds(),
//			r.ProofGenTime.Nanoseconds(),
//			r.ProofSize,
//			r.ProofVerifyTime.Nanoseconds(),
//			r.MemoryAllocatedMB,
//			r.TotalAllocatedMB,
//			r.HeapObjects)
//	}
//}
//
//type ProofResult struct {
//	proofSizeBytes   int
//	proofTime        time.Duration
//	verificationTime time.Duration
//}
//
//func runBenchmark(t *testing.T, count int) BenchmarkResult {
//	avl := avlhashtree.NewAVLHashTree()
//	blockSize := 1024
//
//	sampleSize := count / 100
//	sampleIndices := map[int]bool{}
//	for j := 0; j < sampleSize; j++ {
//		sampleIndices[randMath.Intn(count)] = true
//	}
//
//	// Memory stats before
//	var m1 runtime.MemStats
//	runtime.ReadMemStats(&m1)
//
//	// Start timing
//	start := time.Now()
//
//	proofGenerationKeySample := make([]utils.Hash, 0, sampleSize)
//
//	for i := 0; i < count; i++ {
//		block := make([]byte, blockSize)
//		rand.Read(block)
//		data := make([]byte, 8)
//		rand.Read(data)
//
//		hashOfBlock := utils.GenerateHash(block)
//
//		if _, ok := sampleIndices[i]; ok {
//			proofGenerationKeySample = append(proofGenerationKeySample, hashOfBlock)
//		}
//
//		err := avl.Insert(hashOfBlock, data)
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
//	// End timing
//	elapsed := time.Since(start)
//
//	// Memory stats after
//	var m2 runtime.MemStats
//	runtime.ReadMemStats(&m2)
//
//	proofResults := map[int]ProofResult{}
//
//	for idx, proofKey := range proofGenerationKeySample {
//		startProof := time.Now()
//		randKeyCBOR, _ := utils.EncodeCBOR(proofKey)
//
//		proof, _ := avl.GenerateInclusionExclusionProof(randKeyCBOR)
//		elapsedProof := time.Since(startProof)
//
//		publicProof := proof.ToPublicProof()
//		proofCbor, _ := utils.EncodeCBOR(publicProof)
//
//		startVerif := time.Now()
//
//		res, err := avl.VerifyPublicProof(publicProof)
//
//		if err != nil {
//			t.Fatal(err)
//		}
//
//		if res != true {
//			t.Errorf("Verification failed!")
//		}
//
//		elapsedVerif := time.Since(startVerif)
//
//		proofResults[idx] = ProofResult{
//			proofSizeBytes:   len(proofCbor),
//			proofTime:        elapsedProof,
//			verificationTime: elapsedVerif,
//		}
//	}
//
//	// Calculate memory used
//	allocatedMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
//	totalAllocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
//
//	// Proofs
//	var totalProofTime time.Duration
//	var totalProofSize int
//	var totalVerifyTime time.Duration
//
//	for _, pr := range proofResults {
//		totalProofTime += pr.proofTime
//		totalProofSize += pr.proofSizeBytes
//		totalVerifyTime += pr.verificationTime
//	}
//
//	avgProofTime := totalProofTime / time.Duration(len(proofResults))
//	avgProofSize := totalProofSize / len(proofResults)
//	avgVerifyTime := totalVerifyTime / time.Duration(len(proofResults))
//
//	// Log individual test results
//	t.Logf("Inserted %d blocks", count)
//	t.Logf("Time taken: %v", elapsed)
//	t.Logf("Overall avg per block: %v", elapsed/time.Duration(count))
//	//t.Logf("Time taken for Proof generation: %v", elapsedProof)
//	//t.Logf("Proof size in bytes: %d", len(proofCbor))
//	//t.Logf("Time taken for Proof verification: %v", elapsedVerif)
//	t.Logf("Average time for Proof generation: %v", avgProofTime)
//	t.Logf("Average proof size in bytes: %d", avgProofSize)
//	t.Logf("Average time for Proof verification: %v", avgVerifyTime)
//	t.Logf("Memory allocated: %.2f MB", allocatedMB)
//	t.Logf("Total allocated (including GC'd): %.2f MB", totalAllocMB)
//	t.Logf("Heap objects: %d", m2.HeapObjects)
//
//	return BenchmarkResult{
//		ElementCount:  count,
//		InsertionTime: elapsed,
//		AvgPerBlock:   elapsed / time.Duration(count),
//		//ProofGenTime:      elapsedProof,
//		//ProofSize:         len(proofCbor),
//		//ProofVerifyTime:   elapsedVerif,
//		ProofGenTime:      avgProofTime,
//		ProofSize:         avgProofSize,
//		ProofVerifyTime:   avgVerifyTime,
//		MemoryAllocatedMB: allocatedMB,
//		TotalAllocatedMB:  totalAllocMB,
//		HeapObjects:       m2.HeapObjects,
//	}
//}

package test

import (
	"crypto/rand"
	"fmt"
	randMath "math/rand"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/utils"
)

type BenchmarkResult struct {
	ElementCount      int
	InsertionTime     time.Duration
	AvgPerBlock       time.Duration
	ProofGenTime      time.Duration
	ProofSize         int
	ProofVerifyTime   time.Duration
	MemoryAllocatedMB float64
	TotalAllocatedMB  float64
	HeapObjects       uint64
}

type BenchmarkOptions struct {
	IncludeInclusionProof bool
	IncludeExclusionProof bool
	SampleSize            float32 // percentage
	BlockSizeBytes        int
	DataSizeBytes         int
	MeasureInserts        bool
	MeasureDeletes        bool
	ExecuteDeletes        bool
	ElementCount          int
}

type ProofResult struct {
	proofSizeBytes   int
	proofTime        time.Duration
	verificationTime time.Duration
}

var allResults []BenchmarkResult

func TestMain(m *testing.M) {
	// Run all tests
	exitCode := m.Run()

	// After all tests, print combined results
	if len(allResults) > 0 {
		printCombinedResults()
		saveResultsToCSV()
	}

	os.Exit(exitCode)
}

// 1.
// ------------------------------

func TestAVL_10k(t *testing.T) {
	result := runBenchmark(t, 10_000)
	allResults = append(allResults, result)
}

func TestAVL_50k(t *testing.T) {
	result := runBenchmark(t, 50_000)
	allResults = append(allResults, result)
}

func TestAVL_100k(t *testing.T) {
	result := runBenchmark(t, 100_000)
	allResults = append(allResults, result)
}

func TestAVL_200k(t *testing.T) {
	result := runBenchmark(t, 200_000)
	allResults = append(allResults, result)
}

func TestAVL_300k(t *testing.T) {
	result := runBenchmark(t, 300_000)
	allResults = append(allResults, result)
}

func TestAVL_500k(t *testing.T) {
	result := runBenchmark(t, 500_000)
	allResults = append(allResults, result)
}

func TestAVL_700k(t *testing.T) {
	result := runBenchmark(t, 700_000)
	allResults = append(allResults, result)
}

func TestAVL_900k(t *testing.T) {
	result := runBenchmark(t, 900_000)
	allResults = append(allResults, result)
}

func TestAVL_1M(t *testing.T) {
	result := runBenchmark(t, 1_000_000)
	allResults = append(allResults, result)
}

func TestAVL_2M(t *testing.T) {
	result := runBenchmark(t, 2_000_000)
	allResults = append(allResults, result)
}

func TestAVL_5M(t *testing.T) {
	result := runBenchmark(t, 5_000_000)
	allResults = append(allResults, result)
}

func TestAVL_10M(t *testing.T) {
	result := runBenchmark(t, 10_000_000)
	allResults = append(allResults, result)
}

// ------------------------------

func printCombinedResults() {
	fmt.Println("\n\n========================================")
	fmt.Println("BENCHMARK SUITE RESULTS")
	fmt.Println("========================================")

	fmt.Printf("%-12s %-15s %-15s %-15s %-12s %-15s %-15s %-15s %-12s\n",
		"Elements", "Insert Time", "Avg/Block", "Proof Gen", "Proof Size", "Proof Verify", "Mem Alloc MB", "Total Alloc MB", "Heap Objs")
	fmt.Printf("%-12s %-15s %-15s %-15s %-12s %-15s %-15s %-15s %-12s\n",
		"------------", "---------------", "---------------", "---------------", "------------", "---------------", "---------------", "---------------", "------------")

	for _, r := range allResults {
		fmt.Printf("%-12d %-15v %-15v %-15v %-12d %-15v %-15.2f %-15.2f %-12d\n",
			r.ElementCount,
			r.InsertionTime,
			r.AvgPerBlock,
			r.ProofGenTime,
			r.ProofSize,
			r.ProofVerifyTime,
			r.MemoryAllocatedMB,
			r.TotalAllocatedMB,
			r.HeapObjects)
	}
}

func saveResultsToCSV() {
	now := time.Now()
	filename := fmt.Sprintf("%02d-%02d-%d-%02d-%02d-%02d.csv",
		now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second())

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer file.Close()

	file.WriteString("Elements,InsertTime(ns),AvgPerBlock(ns),ProofGen(ns),ProofSize(bytes),ProofVerify(ns),MemAllocMB,TotalAllocMB,HeapObjects\n")

	for _, r := range allResults {
		line := fmt.Sprintf("%d,%d,%d,%d,%d,%d,%.2f,%.2f,%d\n",
			r.ElementCount,
			r.InsertionTime.Nanoseconds(),
			r.AvgPerBlock.Nanoseconds(),
			r.ProofGenTime.Nanoseconds(),
			r.ProofSize,
			r.ProofVerifyTime.Nanoseconds(),
			r.MemoryAllocatedMB,
			r.TotalAllocatedMB,
			r.HeapObjects)
		file.WriteString(line)
	}

	fmt.Printf("\nResults saved to: %s\n", filename)
}

func runBenchmark(t *testing.T, count int) BenchmarkResult {
	t.Logf("\n========================================")
	t.Logf("Starting test with %d elements", count)
	t.Logf("========================================")

	avl := avlhashtree.NewAVLHashTree()
	blockSize := 1024

	sampleSize := count / 100
	if sampleSize == 0 {
		sampleSize = 1
	}
	sampleIndices := map[int]bool{}
	for j := 0; j < sampleSize; j++ {
		sampleIndices[randMath.Intn(count)] = true
	}

	// Memory stats before
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Start timing
	start := time.Now()

	proofGenerationKeySample := make([]utils.Hash, 0, sampleSize)

	for i := 0; i < count; i++ {
		block := make([]byte, blockSize)
		rand.Read(block)
		data := make([]byte, 8)
		rand.Read(data)

		hashOfBlock := utils.GenerateHash(block)

		if _, ok := sampleIndices[i]; ok {
			proofGenerationKeySample = append(proofGenerationKeySample, hashOfBlock)
		}

		err := avl.Insert(hashOfBlock, data)
		if err != nil {
			t.Fatal(err)
		}
	}

	// End timing
	elapsed := time.Since(start)

	// Memory stats after
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	proofResults := map[int]ProofResult{}

	for idx, proofKey := range proofGenerationKeySample {
		startProof := time.Now()
		randKeyCBOR, _ := utils.EncodeCBOR(proofKey)

		proof, _ := avl.GenerateInclusionExclusionProof(randKeyCBOR)
		elapsedProof := time.Since(startProof)

		publicProof := proof.ToPublicProof()
		proofCbor, _ := utils.EncodeCBOR(publicProof)

		startVerif := time.Now()

		res, err := avl.VerifyPublicProof(publicProof)

		if err != nil {
			t.Fatal(err)
		}

		if res != true {
			t.Errorf("Verification failed!")
		}

		elapsedVerif := time.Since(startVerif)

		proofResults[idx] = ProofResult{
			proofSizeBytes:   len(proofCbor),
			proofTime:        elapsedProof,
			verificationTime: elapsedVerif,
		}
	}

	// Calculate memory used
	allocatedMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
	totalAllocMB := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024

	// Proofs
	var totalProofTime time.Duration
	var totalProofSize int
	var totalVerifyTime time.Duration

	for _, pr := range proofResults {
		totalProofTime += pr.proofTime
		totalProofSize += pr.proofSizeBytes
		totalVerifyTime += pr.verificationTime
	}

	avgProofTime := totalProofTime / time.Duration(len(proofResults))
	avgProofSize := totalProofSize / len(proofResults)
	avgVerifyTime := totalVerifyTime / time.Duration(len(proofResults))

	// Log individual test results
	t.Logf("Inserted %d blocks", count)
	t.Logf("Time taken: %v", elapsed)
	t.Logf("Overall avg per block: %v", elapsed/time.Duration(count))
	t.Logf("Average time for Proof generation: %v", avgProofTime)
	t.Logf("Average proof size in bytes: %d", avgProofSize)
	t.Logf("Average time for Proof verification: %v", avgVerifyTime)
	t.Logf("Memory allocated: %.2f MB", allocatedMB)
	t.Logf("Total allocated (including GC'd): %.2f MB", totalAllocMB)
	t.Logf("Heap objects: %d", m2.HeapObjects)

	// Force GC between tests
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	return BenchmarkResult{
		ElementCount:      count,
		InsertionTime:     elapsed,
		AvgPerBlock:       elapsed / time.Duration(count),
		ProofGenTime:      avgProofTime,
		ProofSize:         avgProofSize,
		ProofVerifyTime:   avgVerifyTime,
		MemoryAllocatedMB: allocatedMB,
		TotalAllocatedMB:  totalAllocMB,
		HeapObjects:       m2.HeapObjects,
	}
}
