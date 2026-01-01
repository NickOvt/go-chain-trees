package test

import (
	"crypto/rand"
	"fmt"
	"math"
	randMath "math/rand"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/utils"
)

type BenchmarkResult struct {
	//// Inclusion Proofs
	InclusionProofGenTime    time.Duration
	InclusionProofSize       int
	InclusionProofVerifyTime time.Duration
	////

	//// Exclusion Proofs
	ExclusionProofGenTime    time.Duration
	ExclusionProofSize       int
	ExclusionProofVerifyTime time.Duration
	////

	//// Inserts
	InsertElementCount int
	InsertionTime      time.Duration
	AvgPerBlock        time.Duration
	MemoryAllocatedMB  float64
	TotalAllocatedMB   float64
	HeapObjects        uint64
	////

	//// Deletes
	DeleteElementCount       int
	DeletionTime             time.Duration
	AvgDeletionPerBlock      time.Duration
	DeletesMemoryAllocatedMB float64
	DeletesTotalAllocatedMB  float64
	DeletesHeapObjects       uint64
	////
}

type BenchmarkOptions struct {
	IncludeInclusionProof    bool
	IncludeExclusionProof    bool
	InclusionProofSequential bool
	SampleSize               float32 // percentage
	BlockSizeBytes           int
	DataSizeBytes            int
	MeasureInserts           bool
	MeasureDeletes           bool
	ElementCount             int
	DeleteCount              int
	HashAlgo                 utils.HashAlgo
	DeleteSequential         bool
	CPUProfile               bool
}

type ProofResult struct {
	proofSizeBytes   int
	proofTime        time.Duration
	verificationTime time.Duration
}

type InclusionExclusionProofResult struct {
	avgProofTime  time.Duration
	avgProofSize  int
	avgVerifyTime time.Duration
}

var allResults []BenchmarkResult

type InsertDeleteMetrics struct {
	Elapsed      time.Duration
	AllocatedMB  float64
	TotalAllocMB float64
	HeapObjects  uint64
}

var now time.Time

func TestMain(m *testing.M) {
	// Run all tests
	now = time.Now()
	exitCode := m.Run()

	// After all tests, print combined results
	if len(allResults) > 0 {
		printCombinedResults()
		saveResultsToCSV(now)
	}

	os.Exit(exitCode)
}

func calculateProofBenchmark(proofResults map[int]ProofResult) (time.Duration, int, time.Duration) {
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

	return avgProofTime, avgProofSize, avgVerifyTime
}

// 1.
// ------------------------------

func TestAVL_10k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 10_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8, MeasureDeletes: true, DeleteCount: 1000})
	allResults = append(allResults, result)
}

func TestAVL_50k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 50_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8, MeasureDeletes: true, DeleteCount: 1000})
	allResults = append(allResults, result)
}

func TestAVL_100k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 100_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_200k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 200_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_300k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 300_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_500k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 500_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_700k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 700_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_900k(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 900_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_1M(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 1_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_2M(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 2_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_5M(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 5_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

func TestAVL_10M(t *testing.T) {
	result := testWithProfile(t, &BenchmarkOptions{CPUProfile: true, ElementCount: 10_000_000, SampleSize: 0.01, MeasureInserts: true, IncludeInclusionProof: true, BlockSizeBytes: 1024, DataSizeBytes: 8})
	allResults = append(allResults, result)
}

// ------------------------------

func printCombinedResults() {
	fmt.Println("\n\n========================================")
	fmt.Println("BENCHMARK SUITE RESULTS")
	fmt.Println("========================================")

	fmt.Printf("%-25s %-20s %-20s %-25s %-25s %-22s %-22s %-23s %-25s %-22s %-23s %-25s %-22s %-20s %-20s %-25s %-25s %-22s\n",
		"Elements [Inserts]",
		"Insert Time [Inserts]",
		"Avg/Block [Inserts]",
		"Mem Alloc MB [Inserts]",
		"Total Alloc MB [Inserts]",
		"Heap Objs [Inserts]",
		"[Inclusion] Proof Gen",
		"[Inclusion] Proof Size",
		"[Inclusion] Proof Verify",
		"[Exclusion] Proof Gen",
		"[Exclusion] Proof Size",
		"[Exclusion] Proof Verify",
		"Elements [Deletes]",
		"Delete Time [Deletes]",
		"Avg/Block [Deletes]",
		"Mem Alloc MB [Deletes]",
		"Total Alloc MB [Deletes]",
		"Heap Objs [Deletes]")

	fmt.Printf("%-25s %-20s %-20s %-25s %-25s %-22s %-22s %-23s %-25s %-22s %-23s %-25s %-22s %-20s %-20s %-25s %-25s %-22s\n",
		strings.Repeat("-", 25),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20),
		strings.Repeat("-", 25),
		strings.Repeat("-", 25),
		strings.Repeat("-", 22),
		strings.Repeat("-", 22),
		strings.Repeat("-", 23),
		strings.Repeat("-", 25),
		strings.Repeat("-", 22),
		strings.Repeat("-", 23),
		strings.Repeat("-", 25),
		strings.Repeat("-", 22),
		strings.Repeat("-", 20),
		strings.Repeat("-", 20),
		strings.Repeat("-", 25),
		strings.Repeat("-", 25),
		strings.Repeat("-", 22))
	for _, r := range allResults {
		fmt.Printf("%-25d %-20v %-20v %-25.2f %-25.2f %-22d %-22v %-23d %-25v %-22v %-23d %-25v %-22d %-20v %-20v %-25.2f %-25.2f %-22d\n",
			r.InsertElementCount,
			r.InsertionTime,
			r.AvgPerBlock,
			r.MemoryAllocatedMB,
			r.TotalAllocatedMB,
			r.HeapObjects,
			r.InclusionProofGenTime,
			r.InclusionProofSize,
			r.InclusionProofVerifyTime,
			r.ExclusionProofGenTime,
			r.ExclusionProofSize,
			r.ExclusionProofVerifyTime,
			r.DeleteElementCount,
			r.DeletionTime,
			r.AvgDeletionPerBlock,
			r.DeletesMemoryAllocatedMB,
			r.DeletesTotalAllocatedMB,
			r.DeletesHeapObjects,
		)
	}
}

func testWithProfile(t *testing.T, options *BenchmarkOptions) BenchmarkResult {
	if !options.CPUProfile {
		return runBenchmark(t, options)
	}

	now := time.Now()
	filename := fmt.Sprintf("cpu_%dk_%02d-%02d-%d-%02d-%02d-%02d.prof", options.ElementCount/1000, now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second())
	f, _ := os.Create(filename)
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	return runBenchmark(t, options)
}

func saveResultsToCSV(now time.Time) {
	filename := fmt.Sprintf("%02d-%02d-%d-%02d-%02d-%02d.csv",
		now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second())

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer file.Close()

	// Header
	file.WriteString("InsertElements,InsertTime(ns),AvgPerBlock(ns),MemAllocMB,TotalAllocMB,HeapObjects,InclusionProofGen(ns),InclusionProofSize(bytes),InclusionProofVerify(ns),ExclusionProofGen(ns),ExclusionProofSize(bytes),ExclusionProofVerify(ns),DeleteElements,DeleteTime(ns),AvgDeletePerBlock(ns),DeleteMemAllocMB,DeleteTotalAllocMB,DeleteHeapObjects\n")

	// Data rows
	for _, r := range allResults {
		line := fmt.Sprintf("%d,%d,%d,%.2f,%.2f,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f,%d\n",
			r.InsertElementCount,
			r.InsertionTime.Nanoseconds(),
			r.AvgPerBlock.Nanoseconds(),
			r.MemoryAllocatedMB,
			r.TotalAllocatedMB,
			r.HeapObjects,
			r.InclusionProofGenTime.Nanoseconds(),
			r.InclusionProofSize,
			r.InclusionProofVerifyTime.Nanoseconds(),
			r.ExclusionProofGenTime.Nanoseconds(),
			r.ExclusionProofSize,
			r.ExclusionProofVerifyTime.Nanoseconds(),
			r.DeleteElementCount,
			r.DeletionTime.Nanoseconds(),
			r.AvgDeletionPerBlock.Nanoseconds(),
			r.DeletesMemoryAllocatedMB,
			r.DeletesTotalAllocatedMB,
			r.DeletesHeapObjects)
		file.WriteString(line)
	}

	fmt.Printf("\nResults saved to: %s\n", filename)
}

func runBenchmark(t *testing.T, options *BenchmarkOptions) BenchmarkResult {
	t.Logf("\n========================================")
	t.Logf("Starting test with %d elements", options.ElementCount)
	t.Logf("========================================")

	//// Tree
	if options.HashAlgo == "" {
		options.HashAlgo = utils.SHA256
	}

	avl := avlhashtree.NewAVLHashTree(options.HashAlgo)
	////

	//// Block size and Data
	blockSize := options.BlockSizeBytes
	if blockSize == 0 {
		blockSize = 1024
	}
	dataSize := options.DataSizeBytes
	if dataSize == 0 {
		dataSize = 8
	}
	////

	//// Inclusion proof
	sampleSize := int(math.Round(float64(float32(options.ElementCount) * options.SampleSize)))
	if sampleSize < 1 {
		sampleSize = 1
	}

	sampleIndices := map[int]bool{}
	var proofGenerationKeySample []utils.Hash
	if options.IncludeInclusionProof {
		proofGenerationKeySample = make([]utils.Hash, sampleSize)
		if options.InclusionProofSequential {
			// nodes taken for proof sequentially
			for j := 0; j < sampleSize; j++ {
				sampleIndices[j] = true
			}
		} else {
			for len(sampleIndices) < sampleSize {
				// nodes taken for proof randomly
				sampleIndices[randMath.Intn(options.ElementCount)] = true
			}
		}
	}
	////

	//// General benchmark results
	results := map[string]any{}
	////

	//// Deletes
	deleteIndices := map[int]bool{}
	var deletesKeySample []utils.Hash

	if options.MeasureDeletes && options.DeleteCount > 0 {
		deletesKeySample = make([]utils.Hash, options.DeleteCount)
		if options.DeleteSequential {
			// sequential deletes
			for j := 0; j < options.DeleteCount; j++ {
				deleteIndices[j] = true
			}
		} else {
			// random deletes
			for len(deleteIndices) < options.DeleteCount {
				deleteIndices[randMath.Intn(options.ElementCount)] = true
			}
		}
	}
	////

	//// Inserts
	sampleIndicesI := 0
	deleteSampleIndicesI := 0
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
				proofGenerationKeySample[sampleIndicesI] = hashOfBlock
				sampleIndicesI++
			}

			if _, ok := deleteIndices[i]; ok {
				deletesKeySample[deleteSampleIndicesI] = hashOfBlock
				deleteSampleIndicesI++
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
	////

	//// Inclusion Proofs
	if options.IncludeInclusionProof {
		inclusionProofResults := map[int]ProofResult{}
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

			inclusionProofResults[idx] = ProofResult{
				proofSizeBytes:   len(proofCbor),
				proofTime:        elapsedProof,
				verificationTime: elapsedVerif,
			}

			avgProofTime, avgProofSize, avgVerifyTime := calculateProofBenchmark(inclusionProofResults)

			results["inclusionProof"] = InclusionExclusionProofResult{
				avgProofTime,
				avgProofSize,
				avgVerifyTime,
			}
		}
	}
	////

	//// Exclusion Proofs
	if options.IncludeExclusionProof {
		exclusionProofResults := map[int]ProofResult{}
		for idx := range sampleSize {
			startProof := time.Now()
			block := make([]byte, blockSize)
			rand.Read(block)

			hashOfBlock := utils.GenerateHashSha256(block)
			randKeyCBOR, _ := utils.EncodeCBOR(hashOfBlock)

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

			exclusionProofResults[idx] = ProofResult{
				proofSizeBytes:   len(proofCbor),
				proofTime:        elapsedProof,
				verificationTime: elapsedVerif,
			}

			avgProofTime, avgProofSize, avgVerifyTime := calculateProofBenchmark(exclusionProofResults)

			results["exclusionProof"] = InclusionExclusionProofResult{
				avgProofTime,
				avgProofSize,
				avgVerifyTime,
			}
		}
	}
	////

	//// Deletes
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
	////

	// Log individual test results
	t.Logf("Inserted %d blocks", options.ElementCount)

	if val, ok := results["inserts"]; ok {
		if inserts, ok := val.(InsertDeleteMetrics); ok {
			t.Logf("Time taken: %v", inserts.Elapsed)
			t.Logf("Overall avg per block: %v", inserts.Elapsed/time.Duration(options.ElementCount))
		}
	}

	if val, ok := results["inclusionProof"]; ok {
		if proofResult, ok := val.(InclusionExclusionProofResult); ok {
			t.Logf("Average time for Inclusion Proof generation: %v", proofResult.avgProofTime)
			t.Logf("Average Inclusion Proof size in bytes: %d", proofResult.avgProofSize)
			t.Logf("Average time for Inclusion Proof verification: %v", proofResult.avgVerifyTime)
		}
	}

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

	if val, ok := results["exclusionProof"]; ok {
		if proofResult, ok := val.(InclusionExclusionProofResult); ok {
			t.Logf("Average time for Exclusion Proof generation: %v", proofResult.avgProofTime)
			t.Logf("Average Exclusion Proof size in bytes: %d", proofResult.avgProofSize)
			t.Logf("Average time for Exclusion Proof verification: %v", proofResult.avgVerifyTime)
		}
	}

	// Force GC between tests
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	result := BenchmarkResult{
		InsertElementCount: options.ElementCount,
	}

	if inserts, ok := results["inserts"].(InsertDeleteMetrics); ok {
		result.InsertionTime = inserts.Elapsed
		result.AvgPerBlock = inserts.Elapsed / time.Duration(options.ElementCount)
		result.MemoryAllocatedMB = inserts.AllocatedMB
		result.TotalAllocatedMB = inserts.TotalAllocMB
		result.HeapObjects = inserts.HeapObjects
	}

	if deletes, ok := results["deletes"].(InsertDeleteMetrics); ok {
		result.DeleteElementCount = options.DeleteCount
		result.DeletionTime = deletes.Elapsed
		result.AvgDeletionPerBlock = deletes.Elapsed / time.Duration(options.DeleteCount)
		result.DeletesMemoryAllocatedMB = deletes.AllocatedMB
		result.DeletesTotalAllocatedMB = deletes.TotalAllocMB
		result.DeletesHeapObjects = deletes.HeapObjects
	}

	if proofResult, ok := results["inclusionProof"].(InclusionExclusionProofResult); ok {
		result.InclusionProofGenTime = proofResult.avgProofTime
		result.InclusionProofSize = proofResult.avgProofSize
		result.InclusionProofVerifyTime = proofResult.avgVerifyTime
	}

	if proofResult, ok := results["exclusionProof"].(InclusionExclusionProofResult); ok {
		result.ExclusionProofGenTime = proofResult.avgProofTime
		result.ExclusionProofSize = proofResult.avgProofSize
		result.ExclusionProofVerifyTime = proofResult.avgVerifyTime
	}

	return result
}
