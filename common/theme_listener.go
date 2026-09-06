package common

import (
	"sync"
	"time"
)

// SystemThemeListener watches the operating system theme and emits a change
// notification when it flips (used with ThemeAuto).
//
// The Python implementation runs a QThread and polls `darkdetect`. The Go port
// uses a plain goroutine; detectSystemTheme currently degrades to ThemeLight
// (see config.go) until a native detector is wired in, so the listener never
// fires spuriously.
type SystemThemeListener struct {
	mu        sync.Mutex
	listeners []func()
	stop      chan struct{}
	once      sync.Once
}

// NewSystemThemeListener builds a listener.
func NewSystemThemeListener() *SystemThemeListener {
	return &SystemThemeListener{stop: make(chan struct{})}
}

// OnSystemThemeChanged registers a listener for system theme changes.
func (l *SystemThemeListener) OnSystemThemeChanged(f func()) {
	l.mu.Lock()
	l.listeners = append(l.listeners, f)
	l.mu.Unlock()
}

func (l *SystemThemeListener) emit() {
	l.mu.Lock()
	ls := append([]func(){}, l.listeners...)
	l.mu.Unlock()
	for _, f := range ls {
		f()
	}
}

// Start begins polling the system theme in the background.
func (l *SystemThemeListener) Start() {
	go func() {
		last := detectSystemTheme()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-l.stop:
				return
			case <-ticker.C:
				t := detectSystemTheme()
				if t != last {
					last = t
					QConfigInstance.SetTheme(ThemeAuto)
					QConfigInstance.Set(QConfigInstance.ThemeMode(), ThemeAuto, false)
					l.emit()
				}
			}
		}
	}()
}

// Stop stops the polling loop.
func (l *SystemThemeListener) Stop() {
	l.once.Do(func() { close(l.stop) })
}
