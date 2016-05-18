package longtest

import (
	"testing"
	"time"
	"os"
	"path/filepath"
	"math"
	filelogger ".."
	"github.com/7phones/tools"
	"io/ioutil"
	"strings"
	"strconv"
	"sort"
)

func TestLogRotate(t *testing.T) {
	testPath := "data/logs"

	defer tools.FilesRemoveAllUpTo(testPath)
	tools.FilesRemoveAllUpTo(testPath)

	waitSeconds := 4

	log, err := filelogger.NewFileLogger(testPath, 0, filelogger.RotateSecond)
	if log == nil {
		t.Error("the result should be a value, but got nil", tools.ErrorsDump(err))
		return
	}

	<-time.After(time.Duration(waitSeconds) * time.Second)

	log.Shutdown()

	// 1. Check
	infoDir, errDir := os.Stat(testPath)
	if errDir != nil {
		t.Error("the base filelogger directory", testPath, "is should exist, but it isn't", tools.ErrorsDump(errDir))
	}
	if !infoDir.IsDir() {
		t.Error("the base path", testPath, "is should be directory, but it isn't", infoDir)
	}

	testFilePattern := filepath.Join(testPath, time.Now().Format("20060102.*.log"))
	catchFiles, _ := filepath.Glob(testFilePattern)
	if math.Abs(float64(len(catchFiles) - waitSeconds)) > 2.0 {
		t.Error("the base path", testPath, "has", len(catchFiles), "but it should have ~", waitSeconds)
	}

	// 2. Check file names
	if files, err := ioutil.ReadDir(testPath); err != nil || len(files) == 0 {
		t.Error("failed to get files list for the test path", testPath)
	} else {
		prevName := ""
		prevNameParts := []string{}

		filesName := []string{}
		for _, file := range files {
			filesName = append(filesName, file.Name())
		}

		sort.Strings(filesName)

		for i, name := range filesName {
			nameParts := strings.SplitN(name, ".", 3)

			if i > 0 {
				prevDate, _ := strconv.ParseInt(prevNameParts[0], 10, 32)
				prevTime, _ := strconv.ParseInt(prevNameParts[1], 10, 32)

				date, _ := strconv.ParseInt(nameParts[0], 10, 32)
				time, _ := strconv.ParseInt(nameParts[1], 10, 32)

				if math.Abs(float64(prevDate - date)) > 1.0 || time > 9 && math.Abs(float64(prevTime - time)) > 1.0 {
					t.Error("the result log file", name, "has a wrong name, the previous file is", prevName, ". The difference between the file expect a one second")
				}
			}

			prevName = name
			prevNameParts = nameParts
		}
	}
}