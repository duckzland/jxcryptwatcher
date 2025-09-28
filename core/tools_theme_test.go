package core

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestAppThemeInit(t *testing.T) {
	th := &appTheme{}
	th.Init()

	th.mu.Lock()
	defer th.mu.Unlock()
	if th.variant != theme.VariantDark {
		t.Errorf("Expected default variant to be VariantDark, got %v", th.variant)
	}
}

func TestAppThemeSetVariant(t *testing.T) {
	th := &appTheme{}
	th.Init()
	th.SetVariant(theme.VariantLight)

	th.mu.Lock()
	defer th.mu.Unlock()
	if th.variant != theme.VariantLight {
		t.Errorf("Expected variant to be VariantLight, got %v", th.variant)
	}
}

func TestAppThemeGetColor(t *testing.T) {
	th := &appTheme{}
	th.Init()

	c := th.GetColor(theme.ColorNameBackground)
	expected := color.RGBA{R: 13, G: 20, B: 33, A: 255}
	if c != expected {
		t.Errorf("Expected background color %v, got %v", expected, c)
	}
}

func TestAppThemeSize(t *testing.T) {
	th := &appTheme{}
	size := th.Size(theme.SizeNameText)
	if size != 14 {
		t.Errorf("Expected text size 14, got %f", size)
	}
}

func TestAppThemeFont(t *testing.T) {
	th := &appTheme{}
	font := th.Font(fyne.TextStyle{Bold: true})
	if font == nil {
		t.Error("Expected non-nil font resource for bold style")
	}
}

func TestAppThemeIcon(t *testing.T) {
	th := &appTheme{}
	icon := th.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Expected non-nil icon resource for IconNameHome")
	}
}

func TestRegisterAndUseThemeManager(t *testing.T) {
	th1 := RegisterThemeManager()
	th2 := UseTheme()

	if th1 != th2 {
		t.Error("Expected RegisterThemeManager and UseTheme to return the same instance")
	}
}
