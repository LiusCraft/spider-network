package xlog

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockW struct {
	t *testing.T
}

func (w *mockW) Write(p []byte) (n int, err error) {
	w.t.Log("test-write", string(p))
	return len(p), nil
}

func TestPrintLog(t *testing.T) {
	xl := New()
	xl.SetOutput(&mockW{t})
	xl.Debugf("test %s", "debug")
	xl.Infof("test %s", "info")
	xl.Warnf("test %s", "warn")
	xl.Errorf("test %s", "error")
	xl.Warn("warn")
	xl.Error("error")
	xl.Info("info")
	xl.Debug("debug")
}

func TestConcurrency(t *testing.T) {
	xl := New()
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

func TestLogger(t *testing.T) {
	t.Run("WithLogger", func(t *testing.T) {
		xl := New()
		xl.Debug("test")
		xl = WithCtx(nil, nil)
		xl.Debug("test")
		WithLogId(xl, "test-id")
		xl.Debug("test")
	})

	t.Run("LoggerContext", func(t *testing.T) {
		xl := WithCtx(New(), context.WithValue(context.Background(), "test", "context"))
		xlCtx := xl.(context.Context)
		if xlCtx.Value("test") != "context" {
			t.Error("context value not match")
		}
		deadline, cancelFunc := context.WithDeadline(xlCtx, time.Now().Add(2*time.Second))
		xlCtx = WithCtx(xl, deadline).(context.Context)
		go func() {
			time.Sleep(time.Second)
			cancelFunc()
		}()
		i := 1
		for i != 0 {
			select {
			case <-xlCtx.Done():
				t.Log("context deadline succeeded")
				i = 0
			default:
				time.Sleep(time.Second)
				i++
				if i > 2 {
					t.Error("context deadline failed")
				}
			}
		}
		if xlCtx.Err() == nil {
			t.Error("context deadline failed")
		}
	})
}
