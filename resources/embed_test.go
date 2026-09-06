package resources

import "testing"

func TestEmbeddedQSS(t *testing.T) {
	for _, theme := range []string{"light", "dark"} {
		qss := QSS(theme, "button")
		if qss == "" {
			t.Fatalf("expected non-empty QSS for %s/button.qss", theme)
		}
	}
}

func TestEmbeddedIcons(t *testing.T) {
	if b := IconSVG("Up", "black"); len(b) == 0 {
		t.Fatal("expected Up_black.svg to be embedded")
	}
	if b := IconSVG("Up", "white"); len(b) == 0 {
		t.Fatal("expected Up_white.svg to be embedded")
	}
}

func TestEmbeddedI18n(t *testing.T) {
	if b := Translation("zh_CN"); len(b) == 0 {
		t.Fatal("expected qfluentwidgets.zh_CN.qm to be embedded")
	}
}
