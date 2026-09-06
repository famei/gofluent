package navigation

// Signal helpers.
//
// PyQt5 pyqtSignal instances have no direct miqt equivalent. They are replaced
// by small callback registries with connect/emit semantics (see
// MIGRATION_GUIDE §8). Each registry supports multiple connections, matching
// pyqtSignal.connect, and is only mutated from the GUI thread.

// voidSignal is a signal with no arguments.
type voidSignal struct{ fns []func() }

func (s *voidSignal) connect(f func()) {
	if f != nil {
		s.fns = append(s.fns, f)
	}
}

func (s *voidSignal) emit() {
	for _, f := range s.fns {
		f()
	}
}

// boolSignal carries a single bool argument.
type boolSignal struct{ fns []func(bool) }

func (s *boolSignal) connect(f func(bool)) {
	if f != nil {
		s.fns = append(s.fns, f)
	}
}

func (s *boolSignal) emit(v bool) {
	for _, f := range s.fns {
		f(v)
	}
}

// bool2Signal carries two bool arguments (e.g. itemClicked(triggerByUser,
// clickArrow)).
type bool2Signal struct{ fns []func(bool, bool) }

func (s *bool2Signal) connect(f func(bool, bool)) {
	if f != nil {
		s.fns = append(s.fns, f)
	}
}

func (s *bool2Signal) emit(a, b bool) {
	for _, f := range s.fns {
		f(a, b)
	}
}

// strSignal carries a string argument.
type strSignal struct{ fns []func(string) }

func (s *strSignal) connect(f func(string)) {
	if f != nil {
		s.fns = append(s.fns, f)
	}
}

func (s *strSignal) emit(v string) {
	for _, f := range s.fns {
		f(v)
	}
}

// intSignal carries an int argument.
type intSignal struct{ fns []func(int) }

func (s *intSignal) connect(f func(int)) {
	if f != nil {
		s.fns = append(s.fns, f)
	}
}

func (s *intSignal) emit(v int) {
	for _, f := range s.fns {
		f(v)
	}
}

// modeSignal carries a NavigationDisplayMode argument.
type modeSignal struct{ fns []func(NavigationDisplayMode) }

func (s *modeSignal) connect(f func(NavigationDisplayMode)) {
	if f != nil {
		s.fns = append(s.fns, f)
	}
}

func (s *modeSignal) emit(v NavigationDisplayMode) {
	for _, f := range s.fns {
		f(v)
	}
}
