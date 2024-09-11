package xlog

import "testing"

func TestPrintLog(t *testing.T) {
	xl := NewLogger()
	xl.Debugf("test %s", "debug")
	xl.Infof("test %s", "info")
	xl.Warnf("test %s", "warn")
	xl.Errorf("test %s", "error")
}
