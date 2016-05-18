package fasttest

import (
	"testing"
	filelogger ".."
	"os"
	"path/filepath"
	"time"
	"github.com/7phones/tools"
	"io/ioutil"
	"bytes"
	"fmt"
	"reflect"
	"io"
)

func TestLogErrorEmpty(t *testing.T) {
	log, err := filelogger.NewFileLogger("", 0, filelogger.RotateDaily)

	if log != nil {
		t.Error("the result should be nil, but got a value", tools.ErrorsDump(err))

		log.Shutdown()
	}
}

func TestLogCreate(t *testing.T) {
	testPath := "data/logs"

	defer tools.FilesRemoveAllUpTo(testPath)
	tools.FilesRemoveAllUpTo(testPath)

	log, err := filelogger.NewFileLogger(testPath, 0, filelogger.RotateDaily)
	if log == nil {
		t.Error("the result should be a value, but got nil", tools.ErrorsDump(err))
		return
	}
	log.Shutdown()

	infoDir, errDir := os.Stat(testPath)
	if errDir != nil {
		t.Error("the base filelogger directory", testPath, "is should exist, but it isn't", tools.ErrorsDump(errDir))
	}
	if !infoDir.IsDir() {
		t.Error("the base path", testPath, "is should be directory, but it isn't", infoDir)
	}

	testFileName := filepath.Join(testPath, time.Now().Format("20060102.log"))
	infoFile, errFile := os.Stat(testFileName)
	if errFile != nil {
		t.Error("the file", testFileName, "is should exist, but it isn't", tools.ErrorsDump(errDir))
	}
	if infoFile.IsDir() || !infoFile.Mode().IsRegular() {
		t.Error("the base path", testPath, "is should be a regular file, but it isn't", infoFile)
	}
}

func TestLogWrite(t *testing.T) {
	testPath := "data/logs"

	defer tools.FilesRemoveAllUpTo(testPath)
	tools.FilesRemoveAllUpTo(testPath)

	// generate a test data
	logger, err := filelogger.NewFileLogger(testPath, 0, filelogger.RotateDaily)
	if logger == nil {
		t.Error("the result should be a value, but got nil", tools.ErrorsDump(err))
		return
	}
	testFileName := filepath.Join(testPath, time.Now().Format("20060102.log"))
	expect := bytes.NewBuffer(nil)

	for i := 0; i < 10; i++ {
		logger.Info("hello", i, "friends")
		expect.WriteString(fmt.Sprintf("%v hello\t%v\tfriends\n", time.Now().Format("2006/01/02 15:04:05"), i))
	}

	logger.Shutdown()

	exist, err := ioutil.ReadFile(testFileName)
	if err != nil {
		t.Error("failed to read the file", testFileName, ":", err)
		return
	}

	// check the result
	if expect.Len() == 0 || !reflect.DeepEqual(exist, expect.Bytes()) {
		t.Error("failed to write lines to the log file", testFileName, ". The result file contains\r\n",
			string(exist),
			"\r\n, but the expect result should contains\r\n",
			expect)
	}
}

func TestLogMultiWriter(t *testing.T) {
	testPath := "data/logs"

	defer tools.FilesRemoveAllUpTo(testPath)
	tools.FilesRemoveAllUpTo(testPath)

	testBuffers := []*bytes.Buffer{
		bytes.NewBufferString(""),
		bytes.NewBufferString(""),
	}
	testCopyWriters := []io.Writer{}
	for _, writer := range testBuffers {
		testCopyWriters = append(testCopyWriters, writer)
	}

	// generate a test data
	logger, err := filelogger.NewFileLogger(testPath, 0, filelogger.RotateDaily, testCopyWriters...)
	if logger == nil {
		t.Error("the result should be a value, but got nil", tools.ErrorsDump(err))
		return
	}
	testFileName := filepath.Join(testPath, time.Now().Format("20060102.log"))
	expect := bytes.NewBuffer(nil)

	for i := 0; i < 10; i++ {
		logger.Info("hello", i, "friends")
		expect.WriteString(fmt.Sprintf("%v hello\t%v\tfriends\n", time.Now().Format("2006/01/02 15:04:05"), i))
	}

	logger.Shutdown()

	exist, err := ioutil.ReadFile(testFileName)
	if err != nil {
		t.Error("failed to read the file", testFileName, ":", err)
		return
	}

	// check the result
	if expect.Len() == 0 || !reflect.DeepEqual(exist, expect.Bytes()) {
		t.Error("failed to write lines to the log file", testFileName, ". The result file contains\r\n",
			string(exist),
			"\r\n, but the expect result should contains\r\n",
			expect)
	}

	for i, buffer := range testBuffers {
		if expect.Len() == 0 || !reflect.DeepEqual(buffer.Bytes(), expect.Bytes()) {
			t.Error("failed to write lines to one of the copy writers #", i, ". The result buffer contains\r\n",
				buffer,
				"\r\n, but the expect result should contains\r\n",
				expect)
		}
	}
}
