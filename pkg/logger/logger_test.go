package logger

import (
	"os"
	"strings"
	"testing"

	"github.com/number571/go-peer/pkg/filesystem"
)

const (
	tcPathInfo    = "logger_info_test.txt"
	tcPathWarning = "logger_warning_test.txt"
	tcPathError   = "logger_error_test.txt"
)

const (
	tcTestInfo    = "test_info_text"
	tcTestWarning = "test_warning_text"
	tcTestError   = "test_error_text"
)

func TestLogger(t *testing.T) {
	defer func() {
		os.Remove(tcPathInfo)
		os.Remove(tcPathWarning)
		os.Remove(tcPathError)
	}()

	fileInfo, err := os.OpenFile(tcPathInfo, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Error(err.Error())
		return
	}

	fileWarn, err := os.OpenFile(tcPathWarning, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Error(err.Error())
		return
	}

	fileErro, err := os.OpenFile(tcPathError, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Error(err.Error())
		return
	}

	logger := NewLogger(NewSettings(&SSettings{
		FInfo: fileInfo,
		FWarn: fileWarn,
		FErro: fileErro,
	}))

	logger.PushInfo(tcTestInfo)
	logger.PushWarn(tcTestWarning)
	logger.PushErro(tcTestError)

	res, err := filesystem.OpenFile(tcPathInfo).Read()
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(string(res), tcTestInfo) {
		t.Error("info does not contains tcTestInfo")
	}

	res, err = filesystem.OpenFile(tcPathWarning).Read()
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(string(res), tcTestWarning) {
		t.Error("warning does not contains tcTestWarning")
	}

	res, err = filesystem.OpenFile(tcPathError).Read()
	if err != nil {
		t.Error(err.Error())
	}
	if !strings.Contains(string(res), tcTestError) {
		t.Error("error does not contains tcTestError")
	}
}
