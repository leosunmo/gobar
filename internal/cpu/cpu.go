package cpu

import (
	"fmt"
	"time"

	"barista.run/bar"
	"barista.run/base/notifier"
	"barista.run/base/value"
	"barista.run/outputs"
	"barista.run/timing"
	"github.com/mackerelio/go-osstat/cpu"
)

type CPUStat struct {
	User   float64
	System float64
	Idle   float64
}

// Module represents a CPU module that updates on a timer or on demand.
type Module struct {
	interval   time.Duration
	outputFunc value.Value // of func(string) bar.Output
	notifyCh   <-chan struct{}
	notifyFn   func()
	scheduler  *timing.Scheduler
}

// New constructs a new cpu module.
func New(interval time.Duration) *Module {
	m := &Module{interval: interval}
	m.notifyFn, m.notifyCh = notifier.New()
	m.scheduler = timing.NewScheduler()
	m.outputFunc.Set(func(text string) bar.Output {
		return outputs.Text(text)
	})
	return m
}

// Stream starts the module.
func (m *Module) Stream(s bar.Sink) {
	out, err := getCPUStats(m.interval)
	outputFunc := m.outputFunc.Get().(func(CPUStat) bar.Output)
	for {
		if s.Error(err) {
			fmt.Printf("error: %v\n", err)
			return
		}
		s.Output(outputFunc(out))
		select {
		case <-m.outputFunc.Next():
			outputFunc = m.outputFunc.Get().(func(CPUStat) bar.Output)
		case <-m.notifyCh:
			out, err = getCPUStats(m.interval)
		case <-m.scheduler.C:
			out, err = getCPUStats(m.interval)
		}
	}
}

// Output sets the output format. The format func will be passed the entire
// trimmed output from the command once it's done executing. To process output
// by lines, see Tail().
func (m *Module) Output(outputFunc func(CPUStat) bar.Output) *Module {
	m.outputFunc.Set(outputFunc)
	return m
}

// Every sets the refresh interval for the module. The command will be executed
// repeatedly at the given interval, and the output updated. A zero interval
// stops automatic repeats (but Refresh will still work).
func (m *Module) Every(interval time.Duration) *Module {
	if interval == 0 {
		m.scheduler.Stop()
	} else {
		m.scheduler.Every(interval)
	}
	return m
}

// Refresh executes the command and updates the output.
func (m *Module) Refresh() {
	m.notifyFn()
}

func getCPUStats(interval time.Duration) (CPUStat, error) {
	before, err := cpu.Get()
	if err != nil {
		return CPUStat{}, err
	}
	time.Sleep(interval)
	after, err := cpu.Get()
	if err != nil {
		return CPUStat{}, err
	}
	total := float64(after.Total - before.Total)
	return CPUStat{
		User:   float64(after.User-before.User) / total * 100,
		System: float64(after.System-before.System) / total * 100,
		Idle:   float64(after.Idle-before.Idle) / total * 100,
	}, nil
}
