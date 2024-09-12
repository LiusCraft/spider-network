package xlog

import (
	"sync"
	"testing"
)

func TestPrintLog(t *testing.T) {
	xl := NewLogger()
	xl.Debugf("test %s", "debug")
	xl.Infof("test %s", "info")
	xl.Warnf("test %s", "warn")
	xl.Errorf("test %s", "error")
}

func TestConcurrency(t *testing.T) {
	xl := NewLogger()
	g := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		g.Add(1)
		go func() {
			defer func() { g.Done() }()
			xl.Debugf("test %s, %d", "debug", i)
			xl.Infof("test %s, %d", "info", i)
			xl.Warnf("test %s, %d", "warn", i)
			xl.Errorf("test %s, %d", "error", i)
		}()
	}
	g.Wait()
}
