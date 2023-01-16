package chlog

import (
	"go.uber.org/zap"
	"sync"
	"testing"
	"time"
)

func TestProLog(t *testing.T) {
	c := Config{
		Name:              "pro",
		Roll:              false,
		RollConfig:        RollConfig{},
		LogFiles:          []string{"d:/log.txt"},
		InnerFiles:        []string{"d:/loglog.txt"},
		Sampler:           false,
		SamplingConfig:    SamplingConfig{},
		InitialFields:     map[string]interface{}{"common": true},
		DisableCaller:     false,
		DisableStacktrace: false,
	}
	p := Product(c)
	defer p.Close()
	defer p.Sync()

	p.Debug("debug", zap.String("debug", "it is debug"))
	p.Info("info")
	p.Warn("warn", zap.String("warn", "it is warn"))
	p.Error("error", zap.String("error", "it is error"))
	p.DPanic("DPanic", zap.String("DPanic", "it is DPanic"))
	go func() {
		defer func() {
			_ = recover()
		}()
		p.Panic("Panic", zap.String("Panic", "it is Panic"))
	}()
	time.Sleep(time.Second)
	p.Fatal("Fatal", zap.String("Fatal", "it is  fatal"))
}
func TestProLogRoll(t *testing.T) {
	c := Config{
		Name: "pro",
		Roll: true,
		RollConfig: RollConfig{
			File:       "d:/logroll.txt",
			MaxSize:    200,
			MaxBackups: 100,
			MaxAge:     7,
			Compress:   false,
		},
		LogFiles:          []string{"d:/log.txt"},
		InnerFiles:        []string{"d:/loglog.txt"},
		Sampler:           false,
		SamplingConfig:    SamplingConfig{},
		InitialFields:     map[string]interface{}{"common": true},
		DisableCaller:     false,
		DisableStacktrace: false,
	}
	p := Product(c)
	defer p.Close()
	defer p.Sync()

	wg := &sync.WaitGroup{}
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			p.Debug("debug", zap.String("debug", "it is debug"))
			p.Info("info")
			p.Warn("warn", zap.String("warn", "it is warn"))
			p.Error("error", zap.String("error", "it is error"))
			p.DPanic("DPanic", zap.String("DPanic", "it is DPanic"))
			wg.Done()
		}()
	}
	wg.Wait()
}
func TestProLogRollSampler(t *testing.T) {
	c := Config{
		Name: "pro",
		Roll: true,
		RollConfig: RollConfig{
			File:       "d:/logroll.txt",
			MaxSize:    200,
			MaxBackups: 100,
			MaxAge:     7,
			Compress:   false,
		},
		LogFiles:   []string{"d:/log.txt"},
		InnerFiles: []string{"d:/loglog.txt"},
		Sampler:    true,
		SamplingConfig: SamplingConfig{
			Tick:       100 * time.Millisecond,
			Initial:    100,
			Thereafter: 100,
		},
		InitialFields:     map[string]interface{}{"common": true},
		DisableCaller:     false,
		DisableStacktrace: false,
	}
	p := Product(c)
	defer p.Close()
	defer p.Sync()

	wg := &sync.WaitGroup{}
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			p.Debug("debug", zap.String("debug", "it is debug"))
			p.Info("info")
			p.Warn("warn", zap.String("warn", "it is warn"))
			p.Error("error", zap.String("error", "it is error"))
			p.DPanic("DPanic", zap.String("DPanic", "it is DPanic"))
			wg.Done()
		}()
	}
	wg.Wait()
}
