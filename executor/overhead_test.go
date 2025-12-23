package executor

import (
	"os/exec"
	"testing"
)

// BenchmarkShellOverhead measures the time to run 'curl --version'
func BenchmarkShellOverhead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("curl", "--version")
		err := cmd.Run()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFunctionCallOverhead measures a simple function call (simulating CGO)
func BenchmarkFunctionCallOverhead(b *testing.B) {
	mockFunc := func() {}
	for i := 0; i < b.N; i++ {
		mockFunc()
	}
}
