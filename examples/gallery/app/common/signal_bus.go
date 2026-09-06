package common

import "sync"

// SignalBus is the application-wide signal bus (port of app/common/signal_bus.py).
// pyqtSignal emissions are represented by callback registries.
type SignalBus struct {
	mu sync.Mutex

	switchToSampleListeners []func(routeKey string, index int)
	micaEnableListeners     []func(bool)
	supportListeners        []func()
}

// SignalBusInstance is the process-wide signal bus (the Go equivalent of the
// Python module-level `signalBus`).
var SignalBusInstance = &SignalBus{}

// OnSwitchToSample registers a listener for switchToSampleCard(str, int).
func (b *SignalBus) OnSwitchToSample(l func(routeKey string, index int)) {
	b.mu.Lock()
	b.switchToSampleListeners = append(b.switchToSampleListeners, l)
	b.mu.Unlock()
}

// EmitSwitchToSample emits switchToSampleCard.
func (b *SignalBus) EmitSwitchToSample(routeKey string, index int) {
	ls := b.snapshotSwitchToSample()
	for _, l := range ls {
		l(routeKey, index)
	}
}

// OnMicaEnableChanged registers a listener for micaEnableChanged(bool).
func (b *SignalBus) OnMicaEnableChanged(l func(bool)) {
	b.mu.Lock()
	b.micaEnableListeners = append(b.micaEnableListeners, l)
	b.mu.Unlock()
}

// EmitMicaEnableChanged emits micaEnableChanged.
func (b *SignalBus) EmitMicaEnableChanged(enabled bool) {
	ls := b.snapshotMicaEnable()
	for _, l := range ls {
		l(enabled)
	}
}

// OnSupport registers a listener for supportSignal().
func (b *SignalBus) OnSupport(l func()) {
	b.mu.Lock()
	b.supportListeners = append(b.supportListeners, l)
	b.mu.Unlock()
}

// EmitSupport emits supportSignal.
func (b *SignalBus) EmitSupport() {
	ls := b.snapshotSupport()
	for _, l := range ls {
		l()
	}
}

func (b *SignalBus) snapshotSwitchToSample() []func(string, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]func(string, int){}, b.switchToSampleListeners...)
}

func (b *SignalBus) snapshotMicaEnable() []func(bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]func(bool){}, b.micaEnableListeners...)
}

func (b *SignalBus) snapshotSupport() []func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]func(){}, b.supportListeners...)
}
