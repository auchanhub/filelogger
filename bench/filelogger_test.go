package bench

import (
	"testing"
	"github.com/7phones/tools"
	filelogger "../"
)

func BenchmarkLogBuffer(b *testing.B) {
	testPath := "data/logs"

	defer tools.FilesRemoveAllUpTo(testPath)
	tools.FilesRemoveAllUpTo(testPath)

	logger, err := filelogger.NewFileLogger(testPath, 0, filelogger.RotateDaily)
	if logger == nil {
		b.Error("the result should be a value, but got nil", tools.ErrorsDump(err))
		return
	}

	for n := 0; n < b.N; n++ {
		logger.Info("Help me", 12, "to use the", 156, "apples", "Help me", 12, "to use the", 156, "apples")
	}

	logger.Shutdown()

	b.ReportAllocs()
}
