package filelogger

import (
	"os"
	"strings"
	"github.com/pkg/errors"
	"sync"
	"time"
	"bufio"
	"io"
	"log"
	"path/filepath"
	"fmt"
	"bytes"
)

const (
	_ = iota
	RotateSecond
	RotateMinute
	RotateHourly
	RotateDaily
)

type Logger struct {
	lock           sync.Mutex

	log            *log.Logger
	logFlag        int

	baseDir        string
	nameTemplate   string

	logFile        io.WriteCloser
	logBuffer      *bufio.Writer

	rotateTimer    *time.Timer
	rotateDuration time.Duration

	outC           chan string
	shutdownC      chan bool

	copyWriters    []io.Writer
}

func NewFileLogger(baseDir string, flag int, rotateType int, copyWriters ... io.Writer) (logger *Logger, err error) {
	if baseDir = strings.TrimSpace(baseDir); baseDir == "" {
		return nil, errors.Wrap(os.ErrInvalid, "failed to create the logger with an empty path of the base directory")
	}

	logger, err = (&Logger{
		baseDir: baseDir,
		copyWriters: copyWriters,
	}).init(flag, rotateType)

	if logger != nil && err != nil {
		logger.Shutdown()
		logger = nil
	}

	return
}

func (o *Logger) Shutdown() {
	if o.shutdownC != nil {
		o.shutdownC <- true
	}

	if o.rotateTimer != nil {
		o.rotateTimer.Stop()
	}

	o.closeFile()
}

func (o *Logger) init(flag int, rotateType int) (logger *Logger, err error) {
	logger = o

	o.logFlag = log.Ldate | log.Ltime | flag;
	o.setRotateDuration(rotateType)
	o.setFilenameTemplate(rotateType)

	if err = o.createDir(); err != nil {
		err = errors.Wrapf(err, "failed to create the base directory %v in the logger initialization", o.baseDir)

		return
	}

	if err = o.rotate(time.Now()); err != nil {
		err = errors.Wrap(err, "failed to rotate a log file in the logger initialization")
		return
	}

	// create the manage channel after file initializaton to prevent a deadlock
	o.outC = make(chan string)//, 100)
	o.shutdownC = make(chan bool)

	go o.outLine()

	o.resetRotation()

	return
}

func (o *Logger) createDir() error {
	fileInfo, err := os.Stat(o.baseDir)

	if err == nil && !fileInfo.IsDir() {
		return err
	}

	if err = os.MkdirAll(o.baseDir, os.ModePerm); err != nil {
		err = errors.Wrapf(err, "failed to create the base directory %v", o.baseDir)
	}

	return err
}

func (o *Logger) rotate(now time.Time) (err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.closeFile()

	var (
		name string
	)

	name, err = o.openFile(now)
	if err != nil {
		err = errors.Wrapf(err, "failed to open or create the log file %v in a rotate process", name)
	}

	var (
		writer io.Writer = o.logBuffer
	)

	if len(o.copyWriters) > 0 {
		writer = io.MultiWriter(append([]io.Writer{writer}, o.copyWriters...)...)
	}

	o.log = log.New(writer, "", o.logFlag)

	return
}

func (o *Logger) openFile(now time.Time) (name string, err error) {
	name = filepath.Join(o.baseDir, now.Format(o.nameTemplate) + ".log")

	o.logFile, err = os.OpenFile(name, os.O_WRONLY | os.O_APPEND | os.O_CREATE, os.ModePerm)
	if err != nil {
		err = errors.Wrapf(err, "failed to open or create the log file %v", name)
		return
	}

	o.logBuffer = bufio.NewWriter(o.logFile)

	return
}

func (o *Logger) closeFile() {
	if o.logBuffer != nil {
		o.logBuffer.Flush()

		o.logBuffer = nil
	}

	if o.logFile != nil {
		o.logFile.Close()

		o.logFile = nil
	}
}

func (o *Logger) setFilenameTemplate(rotateType int) {
	switch rotateType {
	case RotateSecond:
		o.nameTemplate = "20060102.150405"

	case RotateMinute:
		o.nameTemplate = "20060102.1504"

	case RotateHourly:
		o.nameTemplate = "20060102.15"

	case RotateDaily:
		o.nameTemplate = "20060102"

	default:
		o.nameTemplate = "20060102"
	}
}

func (o *Logger) setRotateDuration(rotateType int) {
	switch rotateType {
	case RotateSecond:
		o.rotateDuration = time.Second

	case RotateMinute:
		o.rotateDuration = time.Minute

	case RotateHourly:
		o.rotateDuration = time.Hour

	case RotateDaily:
		o.rotateDuration = 24 * time.Hour

	default:
		o.rotateDuration = 24 * time.Hour
	}
}

// The timer will raise the channel in a tick if duration is <= 0ns
func (o *Logger) resetRotation() {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.rotateTimer == nil {
		o.rotateTimer = time.NewTimer(o.nextDuration())
	} else {
		o.rotateTimer.Reset(o.nextDuration())
	}
}

// Generate the duration is for next period for rotate a log file
func (o *Logger) nextDuration() time.Duration {
	// The result is <= 0ns on the edge of the day
	return time.Now().Round(o.rotateDuration).Sub(time.Now())
}

func (o *Logger) Info(args ...interface{}) {
	buffer := bytes.NewBufferString("")

	for i, arg := range args {
		if i > 0 {
			buffer.Write([]byte("\t"))
		}

		switch v := arg.(type) {
		case error:
			buffer.WriteString(v.Error())
			return

		case fmt.Stringer:
			buffer.WriteString(v.String())

		default:
			fmt.Fprintf(buffer, "%v", arg)
		}
	}

	o.outC <- buffer.String()
}

// TODO: The method is for write a line to the log with different reasons (e.g. DEBUG, TEST, WARNING)

func (o *Logger) bannerOut() {
	// TODO: The method is for write header of a log session (from the start to the end of the application)
}

func (o *Logger) outLine() {
	for {
		select {
		case line := <-o.outC:
			o.log.Println(line)

		// the edge of the day
		case now := <-o.rotateTimer.C:
			o.rotate(now)
			o.resetRotation()

		case <-o.shutdownC:
			return
		}
	}
}