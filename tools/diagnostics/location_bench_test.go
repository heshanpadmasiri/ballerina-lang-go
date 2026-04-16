package diagnostics_test

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"testing"

	"ballerina-lang-go/context"
	"ballerina-lang-go/semtypes"
	"ballerina-lang-go/test_util/testphases"
)

func BenchmarkPipelineBIR(b *testing.B) {
	inputPath := filepath.Join("..", "..", "corpus", "bal", "subset6", "06-bench", "3-v.bal")
	if _, err := os.Stat(inputPath); err != nil {
		b.Skipf("benchmark input not found: %s", inputPath)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		env := context.NewCompilerEnvironment(semtypes.CreateTypeEnv(), false)
		cx := context.NewCompilerContext(env)
		_, err := testphases.RunPipeline(cx, testphases.PhaseBIR, inputPath)
		if err != nil {
			b.Fatalf("pipeline failed: %v", err)
		}
	}
}

// TestLocationMemProfile runs the pipeline once and writes a heap profile so that
// Location-specific allocations can be inspected with:
//
//	go tool pprof -alloc_space -focus='NewBLangDiagnosticLocation' -text mem.prof
//	go tool pprof -alloc_space -focus='LineRange|LinePositionFromLineAndOffset' -text mem.prof
func TestLocationMemProfile(t *testing.T) {
	inputPath := filepath.Join("..", "..", "corpus", "bal", "subset6", "06-bench", "3-v.bal")
	if _, err := os.Stat(inputPath); err != nil {
		t.Skipf("benchmark input not found: %s", inputPath)
	}

	profPath := filepath.Join(t.TempDir(), "mem.prof")

	// Force GC to get a clean baseline
	runtime.GC()

	env := context.NewCompilerEnvironment(semtypes.CreateTypeEnv(), false)
	cx := context.NewCompilerContext(env)
	_, err := testphases.RunPipeline(cx, testphases.PhaseBIR, inputPath)
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}

	// Write alloc profile (captures all allocations since program start)
	runtime.GC()
	f, err := os.Create(profPath)
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
		f.Close()
		t.Fatalf("write profile: %v", err)
	}
	f.Close()

	t.Logf("Memory profile written to: %s", profPath)
	t.Logf("Inspect with:")
	t.Logf("  go tool pprof -alloc_space -focus='NewBLangDiagnosticLocation' -text %s", profPath)
	t.Logf("  go tool pprof -alloc_space -focus='LineRange|LinePositionFromLineAndOffset' -text %s", profPath)
}
