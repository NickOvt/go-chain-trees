package test_interfaces

import (
	"fmt"
	"math"
	"math/big"
	randMath "math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/NickOvt/go-chain-trees/avlhashtree"
	"github.com/NickOvt/go-chain-trees/smt"
	"github.com/NickOvt/go-chain-trees/utils"
)

const (
	AVLHASHTREE = "avlhashtree"
	SMT         = "smt"

	defaultSampleSize     float32       = 0.01
	defaultBlockSizeBytes               = 32
	defaultDataSizeBytes                = 8
	runtimeResetDelay     time.Duration = 100 * time.Millisecond
)

type BenchmarkResult struct {
	TreeType             string
	Scenario             string
	PrebuildElementCount int
	FinalElementCount    int

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
	TreeType                 string
	ScenarioName             string
	IncludeInclusionProof    bool
	IncludeExclusionProof    bool
	InclusionProofSequential bool
	SampleSize               float32 // percentage
	BlockSizeBytes           int
	DataSizeBytes            int
	PrebuildElementCount     int
	MeasureInserts           bool
	MeasureExistingInserts   bool
	MeasureDeletes           bool
	ElementCount             int
	DeleteCount              int
	HashAlgo                 utils.HashAlgo
	DeleteSequential         bool
	DisableSMTAppendOnly     bool
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

// var allResults []BenchmarkResult

type InsertDeleteMetrics struct {
	Elapsed      time.Duration
	AllocatedMB  float64
	TotalAllocMB float64
	HeapObjects  uint64
}

type benchmarkProfiler struct {
	t        *testing.T
	dirPath  string
	baseName string
	cpuFile  *os.File
	started  bool
	enabled  bool
}

type keySampleCollector struct {
	inclusionIndices map[int]bool
	inclusionKeys    []utils.Hash
	inclusionIndex   int
	deleteIndices    map[int]bool
	deleteKeys       []utils.Hash
	deleteIndex      int
}

func (collector *keySampleCollector) capture(globalIndex int, key utils.Hash) {
	if collector == nil {
		return
	}

	if collector.inclusionKeys != nil && collector.inclusionIndices[globalIndex] {
		collector.inclusionKeys[collector.inclusionIndex] = key
		collector.inclusionIndex++
	}

	if collector.deleteKeys != nil && collector.deleteIndices[globalIndex] {
		collector.deleteKeys[collector.deleteIndex] = key
		collector.deleteIndex++
	}
}

func NewProofOnlyBenchmarkOptions(treeType string, buildCount int, sampleSize float32) *BenchmarkOptions {
	options := newDefaultBenchmarkOptions(treeType)
	options.ScenarioName = fmt.Sprintf("proof_only_after_%s_build", formatCountLabel(buildCount))
	options.PrebuildElementCount = buildCount
	options.IncludeInclusionProof = true

	if sampleSize > 0 {
		options.SampleSize = sampleSize
	}

	return options
}

func NewPostBuildInsertBenchmarkOptions(treeType string, prebuildCount int, insertCount int) *BenchmarkOptions {
	options := newDefaultBenchmarkOptions(treeType)
	options.ScenarioName = fmt.Sprintf("build_%s_then_add_%s_new", formatCountLabel(prebuildCount), formatCountLabel(insertCount))
	options.PrebuildElementCount = prebuildCount
	options.MeasureInserts = true
	options.ElementCount = insertCount
	return options
}

func NewExistingKeyInsertBenchmarkOptions(treeType string, prebuildCount int, insertCount int) *BenchmarkOptions {
	options := newDefaultBenchmarkOptions(treeType)
	options.ScenarioName = fmt.Sprintf("build_%s_then_reinsert_%s_existing", formatCountLabel(prebuildCount), formatCountLabel(insertCount))
	options.PrebuildElementCount = prebuildCount
	options.MeasureInserts = true
	options.MeasureExistingInserts = true
	options.DisableSMTAppendOnly = true
	options.ElementCount = insertCount
	return options
}

func newDefaultBenchmarkOptions(treeType string) *BenchmarkOptions {
	return &BenchmarkOptions{
		TreeType:       treeType,
		SampleSize:     defaultSampleSize,
		BlockSizeBytes: defaultBlockSizeBytes,
		DataSizeBytes:  defaultDataSizeBytes,
		CPUProfile:     true,
	}
}

func formatCountLabel(count int) string {
	switch {
	case count == 0:
		return "0"
	case count%1_000_000 == 0:
		return fmt.Sprintf("%dm", count/1_000_000)
	case count%1_000 == 0:
		return fmt.Sprintf("%dk", count/1_000)
	default:
		return fmt.Sprintf("%d", count)
	}
}

func normalizeBenchmarkOptions(options *BenchmarkOptions) {
	if options.SampleSize <= 0 {
		options.SampleSize = defaultSampleSize
	}

	if options.BlockSizeBytes == 0 {
		options.BlockSizeBytes = defaultBlockSizeBytes
	}

	if options.DataSizeBytes == 0 {
		options.DataSizeBytes = defaultDataSizeBytes
	}

	if strings.TrimSpace(options.ScenarioName) == "" {
		switch {
		case options.PrebuildElementCount > 0 && options.MeasureExistingInserts:
			options.ScenarioName = fmt.Sprintf("build_%s_then_reinsert_%s_existing", formatCountLabel(options.PrebuildElementCount), formatCountLabel(options.ElementCount))
		case options.PrebuildElementCount > 0 && options.MeasureInserts:
			options.ScenarioName = fmt.Sprintf("build_%s_then_add_%s_new", formatCountLabel(options.PrebuildElementCount), formatCountLabel(options.ElementCount))
		case options.PrebuildElementCount > 0 && (options.IncludeInclusionProof || options.IncludeExclusionProof):
			options.ScenarioName = fmt.Sprintf("proof_only_after_%s_build", formatCountLabel(options.PrebuildElementCount))
		default:
			options.ScenarioName = fmt.Sprintf("build_%s", formatCountLabel(options.ElementCount))
		}
	}
}

func PrintCombinedResults(allResults []BenchmarkResult) {
	fmt.Println("\n\n========================================")
	fmt.Println("BENCHMARK SUITE RESULTS")
	fmt.Println("========================================")

	fmt.Printf("%-14s %-34s %-16s %-16s %-18s %-20s %-20s %-25s %-25s %-22s %-22s %-23s %-25s %-22s %-23s %-25s %-22s %-20s %-20s %-25s %-25s %-22s\n",
		"Tree",
		"Scenario",
		"Prebuild",
		"Final Elems",
		"Timed Inserts",
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

	fmt.Printf("%-14s %-34s %-16s %-16s %-18s %-20s %-20s %-25s %-25s %-22s %-22s %-23s %-25s %-22s %-23s %-25s %-22s %-20s %-20s %-25s %-25s %-22s\n",
		strings.Repeat("-", 14),
		strings.Repeat("-", 34),
		strings.Repeat("-", 16),
		strings.Repeat("-", 16),
		strings.Repeat("-", 18),
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
		fmt.Printf("%-14s %-34s %-16d %-16d %-18d %-20v %-20v %-25.2f %-25.2f %-22d %-22v %-23d %-25v %-22v %-23d %-25v %-22d %-20v %-20v %-25.2f %-25.2f %-22d\n",
			r.TreeType,
			r.Scenario,
			r.PrebuildElementCount,
			r.FinalElementCount,
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

func SaveResultsToCSV(now time.Time, allResults []BenchmarkResult) {
	currentDate := fmt.Sprintf("%02d-%02d-%d-%02d-%02d-%02d", now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second())
	dirPath := currentDate
	err := os.MkdirAll(currentDate, 0755)
	if err != nil {
		panic(err)
	}

	filename := currentDate + ".csv"

	file, err := os.Create(filepath.Join(dirPath, filename))
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer file.Close()

	// Header
	file.WriteString("TreeType,Scenario,PrebuildElements,FinalElements,InsertElements,InsertTime(ns),AvgPerBlock(ns),MemAllocMB,TotalAllocMB,HeapObjects,InclusionProofGen(ns),InclusionProofSize(bytes),InclusionProofVerify(ns),ExclusionProofGen(ns),ExclusionProofSize(bytes),ExclusionProofVerify(ns),DeleteElements,DeleteTime(ns),AvgDeletePerBlock(ns),DeleteMemAllocMB,DeleteTotalAllocMB,DeleteHeapObjects\n")

	// Data rows
	for _, r := range allResults {
		line := fmt.Sprintf("%s,%s,%d,%d,%d,%d,%d,%.2f,%.2f,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.2f,%.2f,%d\n",
			r.TreeType,
			r.Scenario,
			r.PrebuildElementCount,
			r.FinalElementCount,
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

func newBenchmarkProfiler(t *testing.T, options *BenchmarkOptions, now time.Time) *benchmarkProfiler {
	if !options.CPUProfile {
		return nil
	}

	currentDate := fmt.Sprintf("%02d-%02d-%d-%02d-%02d-%02d", now.Day(), now.Month(), now.Year(), now.Hour(), now.Minute(), now.Second())
	if err := os.MkdirAll(currentDate, 0755); err != nil {
		panic(err)
	}

	return &benchmarkProfiler{
		t:        t,
		dirPath:  currentDate,
		baseName: buildProfileBaseName(options),
		enabled:  true,
	}
}

func (profiler *benchmarkProfiler) Start() {
	if profiler == nil || !profiler.enabled || profiler.started {
		return
	}

	cpuFile, err := os.Create(filepath.Join(profiler.dirPath, "cpu_"+profiler.baseName+".prof"))
	if err != nil {
		profiler.t.Fatalf("failed to create CPU profile: %v", err)
	}

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		cpuFile.Close()
		profiler.t.Fatalf("failed to start CPU profile: %v", err)
	}

	profiler.cpuFile = cpuFile
	profiler.started = true
}

func (profiler *benchmarkProfiler) Stop() {
	if profiler == nil || !profiler.started {
		return
	}

	pprof.StopCPUProfile()

	if profiler.cpuFile != nil {
		profiler.cpuFile.Close()
		profiler.cpuFile = nil
	}

	runtime.GC()

	heapFile, err := os.Create(filepath.Join(profiler.dirPath, "heap_"+profiler.baseName+".prof"))
	if err != nil {
		profiler.t.Fatalf("failed to create heap profile: %v", err)
	}
	defer heapFile.Close()

	if err := pprof.WriteHeapProfile(heapFile); err != nil {
		profiler.t.Fatalf("failed to write heap profile: %v", err)
	}

	profiler.started = false
}

func buildProfileBaseName(options *BenchmarkOptions) string {
	parts := []string{sanitizeProfileLabel(strings.ToLower(options.TreeType))}
	if options.ScenarioName != "" {
		parts = append(parts, sanitizeProfileLabel(options.ScenarioName))
	}
	if options.PrebuildElementCount > 0 {
		parts = append(parts, "prebuild_"+formatCountLabel(options.PrebuildElementCount))
	}
	if options.ElementCount > 0 {
		parts = append(parts, "ops_"+formatCountLabel(options.ElementCount))
	}

	baseName := strings.Trim(strings.Join(parts, "_"), "_")
	if baseName == "" {
		return "benchmark"
	}

	return baseName
}

func sanitizeProfileLabel(label string) string {
	label = strings.ToLower(label)
	label = strings.NewReplacer(" ", "_", "-", "_", "/", "_").Replace(label)

	var builder strings.Builder
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '_':
			builder.WriteRune(r)
		}
	}

	if builder.Len() == 0 {
		return "benchmark"
	}

	return builder.String()
}

func hasMeasuredPhase(options *BenchmarkOptions) bool {
	return options.MeasureInserts || options.IncludeInclusionProof || options.IncludeExclusionProof || (options.MeasureDeletes && options.DeleteCount > 0)
}

func totalDistinctElementCount(options *BenchmarkOptions) int {
	total := options.PrebuildElementCount
	if options.MeasureInserts && !options.MeasureExistingInserts {
		total += options.ElementCount
	}
	return total
}

func resetRuntimeState() {
	runtime.GC()
	time.Sleep(runtimeResetDelay)
}

func calculateSelectionCount(totalCount int, ratio float32) int {
	if totalCount <= 0 {
		return 0
	}

	count := int(math.Round(float64(float32(totalCount) * ratio)))
	if count < 1 {
		count = 1
	}
	if count > totalCount {
		count = totalCount
	}

	return count
}

func buildSelectionIndices(totalCount int, selectionCount int, sequential bool) map[int]bool {
	indices := map[int]bool{}
	if totalCount <= 0 || selectionCount <= 0 {
		return indices
	}

	if sequential {
		for idx := 0; idx < selectionCount; idx++ {
			indices[idx] = true
		}
		return indices
	}

	for len(indices) < selectionCount {
		indices[randMath.Intn(totalCount)] = true
	}

	return indices
}

func newDataGenerator(dataSize int, counterStart int64) func() []byte {
	dataHashFn := GetCounterHashFuncFrom(counterStart)

	return func() []byte {
		data := make([]byte, dataSize)
		for i := 0; i < len(data); {
			i += copy(data[i:], dataHashFn())
		}
		return data
	}
}

func runFreshInsertBatch(
	count int,
	globalIndex *int,
	keyFn func() utils.Hash,
	nextData func() []byte,
	insertFn func(key utils.Hash, data []byte) error,
	collector *keySampleCollector,
	measure bool,
) (InsertDeleteMetrics, error) {
	if count <= 0 {
		return InsertDeleteMetrics{}, nil
	}

	var before runtime.MemStats
	if measure {
		runtime.ReadMemStats(&before)
	}

	start := time.Now()
	for idx := 0; idx < count; idx++ {
		key := keyFn()
		data := nextData()

		if collector != nil {
			collector.capture(*globalIndex, key)
		}

		if err := insertFn(key, data); err != nil {
			return InsertDeleteMetrics{}, err
		}

		*globalIndex = *globalIndex + 1
	}
	elapsed := time.Since(start)

	if !measure {
		return InsertDeleteMetrics{}, nil
	}

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	return InsertDeleteMetrics{
		Elapsed:      elapsed,
		AllocatedMB:  float64(after.Alloc-before.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(after.TotalAlloc-before.TotalAlloc) / 1024 / 1024,
		HeapObjects:  after.HeapObjects,
	}, nil
}

func measureExistingInsertBatch(
	count int,
	keyFn func() utils.Hash,
	nextData func() []byte,
	insertFn func(key utils.Hash, data []byte) error,
) (InsertDeleteMetrics, error) {
	if count <= 0 {
		return InsertDeleteMetrics{}, nil
	}

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()
	for idx := 0; idx < count; idx++ {
		if err := insertFn(keyFn(), nextData()); err != nil {
			return InsertDeleteMetrics{}, err
		}
	}
	elapsed := time.Since(start)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	return InsertDeleteMetrics{
		Elapsed:      elapsed,
		AllocatedMB:  float64(after.Alloc-before.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(after.TotalAlloc-before.TotalAlloc) / 1024 / 1024,
		HeapObjects:  after.HeapObjects,
	}, nil
}

func measureProofBatch(
	t *testing.T,
	proofKeys []utils.Hash,
	proveAndVerifyFn func(key utils.Hash) (int, time.Duration, time.Duration, error),
) InclusionExclusionProofResult {
	proofResults := map[int]ProofResult{}

	for idx, proofKey := range proofKeys {
		proofSizeBytes, elapsedProof, elapsedVerif, err := proveAndVerifyFn(proofKey)
		if err != nil {
			t.Fatal(err)
		}

		proofResults[idx] = ProofResult{
			proofSizeBytes:   proofSizeBytes,
			proofTime:        elapsedProof,
			verificationTime: elapsedVerif,
		}
	}

	avgProofTime, avgProofSize, avgVerifyTime := calculateProofBenchmark(proofResults)
	return InclusionExclusionProofResult{
		avgProofTime:  avgProofTime,
		avgProofSize:  avgProofSize,
		avgVerifyTime: avgVerifyTime,
	}
}

func measureDeleteBatch(t *testing.T, deleteKeys []utils.Hash, deleteFn func(key utils.Hash) error) InsertDeleteMetrics {
	if len(deleteKeys) == 0 {
		return InsertDeleteMetrics{}
	}

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()
	for _, deleteKey := range deleteKeys {
		if err := deleteFn(deleteKey); err != nil {
			t.Fatal(err)
		}
	}
	elapsed := time.Since(start)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	return InsertDeleteMetrics{
		Elapsed:      elapsed,
		AllocatedMB:  float64(after.Alloc-before.Alloc) / 1024 / 1024,
		TotalAllocMB: float64(after.TotalAlloc-before.TotalAlloc) / 1024 / 1024,
		HeapObjects:  after.HeapObjects,
	}
}

func runBenchmark(t *testing.T, options *BenchmarkOptions, profiler *benchmarkProfiler) BenchmarkResult {
	normalizeBenchmarkOptions(options)

	t.Logf("\n========================================")
	t.Logf("Scenario: %s", options.ScenarioName)
	t.Logf("Measured inserts: %d | prebuild: %d", options.ElementCount, options.PrebuildElementCount)
	t.Logf("========================================")

	if options.HashAlgo == "" {
		options.HashAlgo = utils.SHA256
	}

	treeType := strings.ToLower(options.TreeType)
	if treeType == "" {
		treeType = AVLHASHTREE
	}

	if options.MeasureExistingInserts {
		if options.PrebuildElementCount == 0 {
			t.Fatalf("existing-key insert measurement requires PrebuildElementCount > 0")
		}
		if options.ElementCount > options.PrebuildElementCount {
			t.Fatalf("cannot reinsert %d existing keys from only %d prebuilt elements", options.ElementCount, options.PrebuildElementCount)
		}
	}

	totalDistinctElements := totalDistinctElementCount(options)
	if totalDistinctElements == 0 && (options.IncludeInclusionProof || options.IncludeExclusionProof || (options.MeasureDeletes && options.DeleteCount > 0)) {
		t.Fatalf("proof/delete measurement requires at least one inserted element")
	}

	var insertFn func(key utils.Hash, data []byte) error
	var deleteFn func(key utils.Hash) error
	var proveAndVerifyFn func(key utils.Hash) (int, time.Duration, time.Duration, error)

	switch treeType {
	case AVLHASHTREE:
		avl := avlhashtree.NewAVLHashTree(options.HashAlgo)

		insertFn = func(key utils.Hash, data []byte) error {
			return avl.Insert(key, data)
		}

		deleteFn = func(key utils.Hash) error {
			return avl.Delete(key)
		}

		proveAndVerifyFn = func(key utils.Hash) (int, time.Duration, time.Duration, error) {
			startProof := time.Now()

			keyCBOR, err := utils.EncodeCBOR(key)
			if err != nil {
				return 0, 0, 0, err
			}

			proof, err := avl.GenerateInclusionExclusionProof(keyCBOR)
			if err != nil {
				return 0, 0, 0, err
			}
			elapsedProof := time.Since(startProof)

			publicProof := proof.ToPublicProof()
			proofCbor, err := utils.EncodeCBOR(publicProof)
			if err != nil {
				return 0, 0, 0, err
			}

			startVerif := time.Now()
			res, err := avl.VerifyPublicProof(publicProof)
			elapsedVerif := time.Since(startVerif)

			if err != nil {
				return 0, 0, 0, err
			}
			if !res {
				return 0, 0, 0, fmt.Errorf("verification failed")
			}

			return len(proofCbor), elapsedProof, elapsedVerif, nil
		}
	case SMT:
		if options.MeasureExistingInserts && !options.DisableSMTAppendOnly {
			t.Fatalf("SMT existing-key insert measurement requires DisableSMTAppendOnly=true")
		}

		appendOnly := !options.DisableSMTAppendOnly
		smtTree := smt.NewSMT(options.HashAlgo, appendOnly)

		insertFn = func(key utils.Hash, data []byte) error {
			_, err := smtTree.Insert(key, data)
			return err
		}

		deleteFn = func(key utils.Hash) error {
			return fmt.Errorf("delete benchmarking is not supported for tree type %q", SMT)
		}

		proveAndVerifyFn = func(key utils.Hash) (int, time.Duration, time.Duration, error) {
			startProof := time.Now()

			proofKey := utils.GenerateHash(options.HashAlgo, key)
			proof, err := smtTree.GenerateInclusionExclusionProof(proofKey)
			if err != nil {
				return 0, 0, 0, err
			}
			elapsedProof := time.Since(startProof)

			publicProof := proof.ToPublicProof()
			proofCbor, err := utils.EncodeCBOR(publicProof)
			if err != nil {
				return 0, 0, 0, err
			}

			startVerif := time.Now()
			res, err := smtTree.VerifyPublicProof(publicProof)
			elapsedVerif := time.Since(startVerif)

			if err != nil {
				return 0, 0, 0, err
			}
			if !res {
				return 0, 0, 0, fmt.Errorf("verification failed")
			}

			return len(proofCbor), elapsedProof, elapsedVerif, nil
		}
	default:
		t.Fatalf("unknown tree type %q, expected %q or %q", options.TreeType, AVLHASHTREE, SMT)
	}

	t.Logf("Tree type: %s", treeType)
	if treeType == SMT {
		t.Logf("SMT append-only mode: %v", !options.DisableSMTAppendOnly)
	}

	sampleHashFn := GetCounterHashFunc()
	nextInsertData := newDataGenerator(options.DataSizeBytes, 0)
	nextExistingInsertData := newDataGenerator(options.DataSizeBytes, int64(totalDistinctElements))

	proofSampleSize := 0
	if options.IncludeInclusionProof || options.IncludeExclusionProof {
		proofSampleSize = calculateSelectionCount(totalDistinctElements, options.SampleSize)
	}

	collector := &keySampleCollector{}
	if options.IncludeInclusionProof {
		collector.inclusionIndices = buildSelectionIndices(totalDistinctElements, proofSampleSize, options.InclusionProofSequential)
		collector.inclusionKeys = make([]utils.Hash, proofSampleSize)
	}

	if options.MeasureDeletes && options.DeleteCount > 0 {
		deleteCount := options.DeleteCount
		if deleteCount > totalDistinctElements {
			deleteCount = totalDistinctElements
		}
		collector.deleteIndices = buildSelectionIndices(totalDistinctElements, deleteCount, options.DeleteSequential)
		collector.deleteKeys = make([]utils.Hash, deleteCount)
	}

	results := map[string]any{}
	globalIndex := 0
	profilerStarted := false
	defer func() {
		if profilerStarted {
			profiler.Stop()
		}
	}()

	if options.PrebuildElementCount > 0 {
		t.Logf("Prebuilding %d blocks before measurement", options.PrebuildElementCount)
		if _, err := runFreshInsertBatch(options.PrebuildElementCount, &globalIndex, sampleHashFn, nextInsertData, insertFn, collector, false); err != nil {
			t.Fatal(err)
		}
	}

	if options.PrebuildElementCount > 0 && hasMeasuredPhase(options) {
		resetRuntimeState()
	}

	if hasMeasuredPhase(options) && profiler != nil {
		profiler.Start()
		profilerStarted = true
	}

	if options.MeasureInserts {
		var (
			insertMetrics InsertDeleteMetrics
			err           error
		)

		if options.MeasureExistingInserts {
			insertMetrics, err = measureExistingInsertBatch(options.ElementCount, GetCounterHashFunc(), nextExistingInsertData, insertFn)
		} else {
			insertMetrics, err = runFreshInsertBatch(options.ElementCount, &globalIndex, sampleHashFn, nextInsertData, insertFn, collector, true)
		}
		if err != nil {
			t.Fatal(err)
		}

		results["inserts"] = insertMetrics
	}

	if len(collector.inclusionKeys) > 0 && collector.inclusionIndex != len(collector.inclusionKeys) {
		t.Fatalf("expected %d proof sample keys, captured %d", len(collector.inclusionKeys), collector.inclusionIndex)
	}

	if len(collector.deleteKeys) > 0 && collector.deleteIndex != len(collector.deleteKeys) {
		t.Fatalf("expected %d delete sample keys, captured %d", len(collector.deleteKeys), collector.deleteIndex)
	}

	if options.IncludeInclusionProof {
		results["inclusionProof"] = measureProofBatch(t, collector.inclusionKeys, proveAndVerifyFn)
	}

	if options.IncludeExclusionProof {
		exclusionKeys := make([]utils.Hash, proofSampleSize)
		for idx := range proofSampleSize {
			exclusionKeys[idx] = sampleHashFn()
		}
		results["exclusionProof"] = measureProofBatch(t, exclusionKeys, proveAndVerifyFn)
	}

	if options.MeasureDeletes && len(collector.deleteKeys) > 0 {
		results["deletes"] = measureDeleteBatch(t, collector.deleteKeys, deleteFn)
	}

	if options.PrebuildElementCount > 0 {
		t.Logf("Prebuild complete: %d blocks", options.PrebuildElementCount)
	}
	t.Logf("Final distinct tree size: %d blocks", totalDistinctElements)

	if inserts, ok := results["inserts"].(InsertDeleteMetrics); ok {
		if options.MeasureExistingInserts {
			t.Logf("Reinserted %d existing blocks", options.ElementCount)
		} else {
			t.Logf("Inserted %d new blocks during measured phase", options.ElementCount)
		}
		t.Logf("Time taken: %v", inserts.Elapsed)
		if options.ElementCount > 0 {
			t.Logf("Overall avg per block: %v", inserts.Elapsed/time.Duration(options.ElementCount))
		}
		t.Logf("Memory allocated: %.2f MB", inserts.AllocatedMB)
		t.Logf("Total allocated (including GC'd): %.2f MB", inserts.TotalAllocMB)
		t.Logf("Heap objects: %d", inserts.HeapObjects)
	}

	if proofResult, ok := results["inclusionProof"].(InclusionExclusionProofResult); ok {
		t.Logf("Average time for Inclusion Proof generation: %v", proofResult.avgProofTime)
		t.Logf("Average Inclusion Proof size in bytes: %d", proofResult.avgProofSize)
		t.Logf("Average time for Inclusion Proof verification: %v", proofResult.avgVerifyTime)
	}

	if deletes, ok := results["deletes"].(InsertDeleteMetrics); ok {
		t.Logf("Time taken [Deletes]: %v", deletes.Elapsed)
		if len(collector.deleteKeys) > 0 {
			t.Logf("Overall avg per block [Deletes]: %v", deletes.Elapsed/time.Duration(len(collector.deleteKeys)))
		}
		t.Logf("Memory allocated [Deletes]: %.2f MB", deletes.AllocatedMB)
		t.Logf("Total allocated (including GC'd) [Deletes]: %.2f MB", deletes.TotalAllocMB)
		t.Logf("Heap objects [Deletes]: %d", deletes.HeapObjects)
	}

	if proofResult, ok := results["exclusionProof"].(InclusionExclusionProofResult); ok {
		t.Logf("Average time for Exclusion Proof generation: %v", proofResult.avgProofTime)
		t.Logf("Average Exclusion Proof size in bytes: %d", proofResult.avgProofSize)
		t.Logf("Average time for Exclusion Proof verification: %v", proofResult.avgVerifyTime)
	}

	result := BenchmarkResult{
		TreeType:             treeType,
		Scenario:             options.ScenarioName,
		PrebuildElementCount: options.PrebuildElementCount,
		FinalElementCount:    totalDistinctElements,
		InsertElementCount:   options.ElementCount,
	}

	if inserts, ok := results["inserts"].(InsertDeleteMetrics); ok {
		result.InsertionTime = inserts.Elapsed
		if options.ElementCount > 0 {
			result.AvgPerBlock = inserts.Elapsed / time.Duration(options.ElementCount)
		}
		result.MemoryAllocatedMB = inserts.AllocatedMB
		result.TotalAllocatedMB = inserts.TotalAllocMB
		result.HeapObjects = inserts.HeapObjects
	}

	if deletes, ok := results["deletes"].(InsertDeleteMetrics); ok {
		result.DeleteElementCount = len(collector.deleteKeys)
		result.DeletionTime = deletes.Elapsed
		if len(collector.deleteKeys) > 0 {
			result.AvgDeletionPerBlock = deletes.Elapsed / time.Duration(len(collector.deleteKeys))
		}
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

func TestWithProfile(t *testing.T, options *BenchmarkOptions, now time.Time) BenchmarkResult {
	normalizeBenchmarkOptions(options)
	resetRuntimeState()
	return runBenchmark(t, options, newBenchmarkProfiler(t, options, now))
}

func GetCounterHashFunc() func() utils.Hash {
	return GetCounterHashFuncFrom(0)
}

func GetCounterHashFuncFrom(start int64) func() utils.Hash {
	counter := big.NewInt(start)
	one := big.NewInt(1)

	return func() utils.Hash {
		counter.Add(counter, one)
		return utils.GenerateHashSha256(counter.Bytes())
	}
}
