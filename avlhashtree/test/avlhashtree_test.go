package test

import (
	"crypto/rand"
	"fmt"
	"math"
	randMath "math/rand"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/utils"
)

type BenchmarkResult struct {
	ElementCount int

	ProofGenTime    time.Duration
	ProofSize       int
	ProofVerifyTime time.Duration

	InsertionTime     time.Duration
	AvgPerBlock       time.Duration
	MemoryAllocatedMB float64
	TotalAllocatedMB  float64
	HeapObjects       uint64

	DeletionTime             time.Duration
	AvgDeletionPerBlock      time.Duration
	DeletesMemoryAllocatedMB float64
	DeletesTotalAllocatedMB  float64
	DeletesHeapObjects       uint64
}

type BenchmarkOptions struct {
	IncludeInclusionProof bool
	IncludeExclusionProof bool
	SampleSize            float32 // percentage
	BlockSizeBytes        int
	DataSizeBytes         int
	MeasureInserts        bool
	MeasureDeletes        bool
	ElementCount          int
	DeleteCount           int
	HashAlgo              utils.HashAlgo
}

type ProofResult struct {
	proofSizeBytes   int
	proofTime        time.Duration
	verificationTime time.Duration
}

var allResults []BenchmarkResult

type InsertDeleteMetrics struct {
	Elapsed      time.Duration
	AllocatedMB  float64
	TotalAllocMB float64
	HeapObjects  uint64
}

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
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 10_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8, MeasureDeletes: true, DeleteCount: 1000})
	allResults = append(allResults, result)
}

func TestAVL_50k(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 50_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_100k(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 100_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_200k(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 200_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_300k(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 300_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_500k(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 500_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_700k(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 700_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_900k(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 900_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_1M(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 1_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_2M(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 2_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_5M(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 5_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_10M(t *testing.T) {
	result := runBenchmark(t, &BenchmarkOptions{ElementCount: 10_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
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

func runBenchmark(t *testing.T, options *BenchmarkOptions) BenchmarkResult {
	t.Logf("\n========================================")
	t.Logf("Starting test with %d elements", options.ElementCount)
	t.Logf("========================================")

	if options.HashAlgo == "" {
		options.HashAlgo = utils.SHA256
	}

	avl := avlhashtree.NewAVLHashTree(options.HashAlgo)
	
	blockSize := options.BlockSizeBytes
	if blockSize == 0 {
		blockSize = 1024
	}
	dataSize := options.DataSizeBytes
	if dataSize == 0 {
		dataSize = 8
	}

	sampleSize := int(math.Round(float64(float32(options.ElementCount) * options.SampleSize)))
	if sampleSize < 1 {
		sampleSize = 1
	}
	sampleIndices := map[int]bool{}
	for j := 0; j < sampleSize; j++ {
		sampleIndices[randMath.Intn(options.ElementCount)] = true
	}

	proofGenerationKeySample := make([]utils.Hash, 0, sampleSize)

	results := map[string]any{}

	deleteIndices := map[int]bool{}
	deletesKeySample := make([]utils.Hash, 0, options.DeleteCount)

	if options.MeasureDeletes && options.DeleteCount > 0 {
		for j := 0; j < options.DeleteCount; j++ {
			deleteIndices[randMath.Intn(options.DeleteCount)] = true
		}
	}

	if options.MeasureInserts {
		// Memory stats before
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Start timing
		startInserts := time.Now()

		for i := 0; i < options.ElementCount; i++ {
			block := make([]byte, blockSize)
			rand.Read(block)
			data := make([]byte, dataSize)
			rand.Read(data)

			hashOfBlock := utils.GenerateHashSha256(block)

			if _, ok := sampleIndices[i]; ok {
				proofGenerationKeySample = append(proofGenerationKeySample, hashOfBlock)
			}

			if _, ok := deleteIndices[i]; ok {
				deletesKeySample = append(deletesKeySample, hashOfBlock)
			}

			err := avl.Insert(hashOfBlock, data)
			if err != nil {
				t.Fatal(err)
			}
		}

		// End timing
		elapsedInserts := time.Since(startInserts)

		// Memory stats after
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Calculate memory used
		allocatedMBInserts := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
		totalAllocMBInserts := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
		heapObjects := m2.HeapObjects

		results["inserts"] = InsertDeleteMetrics{
			Elapsed:      elapsedInserts,
			AllocatedMB:  allocatedMBInserts,
			TotalAllocMB: totalAllocMBInserts,
			HeapObjects:  heapObjects,
		}
	}

	proofResults := map[int]ProofResult{}

	if options.IncludeInclusionProof {
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
	}

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

	if options.MeasureDeletes && options.DeleteCount > 0 {
		// Memory stats before
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		// Start timing
		startDeletes := time.Now()

		for _, deleteKey := range deletesKeySample {
			err := avl.Delete(deleteKey)
			if err != nil {
				t.Fatal(err)
			}
		}

		// End timing
		elapsedDeletes := time.Since(startDeletes)

		// Memory stats after
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)

		// Calculate memory used
		allocatedMBDeletes := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
		totalAllocMBDeletes := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024
		heapObjects := m2.HeapObjects

		results["deletes"] = InsertDeleteMetrics{
			Elapsed:      elapsedDeletes,
			AllocatedMB:  allocatedMBDeletes,
			TotalAllocMB: totalAllocMBDeletes,
			HeapObjects:  heapObjects,
		}
	}

	// Log individual test results
	t.Logf("Inserted %d blocks", options.ElementCount)

	if val, ok := results["inserts"]; ok {
		if inserts, ok := val.(InsertDeleteMetrics); ok {
			t.Logf("Time taken: %v", inserts.Elapsed)
			t.Logf("Overall avg per block: %v", inserts.Elapsed/time.Duration(options.ElementCount))
		}
	}

	t.Logf("Average time for Proof generation: %v", avgProofTime)
	t.Logf("Average proof size in bytes: %d", avgProofSize)
	t.Logf("Average time for Proof verification: %v", avgVerifyTime)

	if val, ok := results["inserts"]; ok {
		if inserts, ok := val.(InsertDeleteMetrics); ok {
			t.Logf("Memory allocated: %.2f MB", inserts.AllocatedMB)
			t.Logf("Total allocated (including GC'd): %.2f MB", inserts.TotalAllocMB)
			t.Logf("Heap objects: %d", inserts.HeapObjects)
		}
	}

	if val, ok := results["deletes"]; ok {
		if deletes, ok := val.(InsertDeleteMetrics); ok {
			t.Logf("Time taken [Deletes]: %v", deletes.Elapsed)
			t.Logf("Overall avg per block [Deletes]: %v", deletes.Elapsed/time.Duration(options.DeleteCount))
			t.Logf("Memory allocated [Deletes]: %.2f MB", deletes.AllocatedMB)
			t.Logf("Total allocated (including GC'd) [Deletes]: %.2f MB", deletes.TotalAllocMB)
			t.Logf("Heap objects [Deletes]: %d", deletes.HeapObjects)
		}
	}

	// Force GC between tests
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	result := BenchmarkResult{
		ElementCount:    options.ElementCount,
		ProofGenTime:    avgProofTime,
		ProofSize:       avgProofSize,
		ProofVerifyTime: avgVerifyTime,
	}

	if val, ok := results["inserts"]; ok {
		if inserts, ok := val.(InsertDeleteMetrics); ok {
			result.InsertionTime = inserts.Elapsed
			result.AvgPerBlock = inserts.Elapsed / time.Duration(options.ElementCount)
			result.MemoryAllocatedMB = inserts.AllocatedMB
			result.TotalAllocatedMB = inserts.TotalAllocMB
			result.HeapObjects = inserts.HeapObjects
		}
	}

	if val, ok := results["deletes"]; ok {
		if deletes, ok := val.(InsertDeleteMetrics); ok {
			result.DeletionTime = deletes.Elapsed
			result.AvgDeletionPerBlock = deletes.Elapsed / time.Duration(options.DeleteCount)
			result.DeletesMemoryAllocatedMB = deletes.AllocatedMB
			result.DeletesTotalAllocatedMB = deletes.TotalAllocMB
			result.DeletesHeapObjects = deletes.HeapObjects
		}
	}

	return result
}
