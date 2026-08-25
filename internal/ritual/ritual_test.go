package ritual

import "testing"

func TestQuietCeremonyAndViewport(t *testing.T) {
	ceremony := DefaultCeremony()
	if !ceremony.Valid() {
		t.Fatal("default ceremony invalid")
	}
	if ceremony.Step(5) != "light" {
		t.Fatalf("unexpected step %q", ceremony.Step(5))
	}
	viewport := NewViewport(120, 100).Drag(20, 3).Scale(5)
	if viewport.Width != 320 || viewport.Height != 240 || viewport.Zoom != 3 {
		t.Fatalf("unexpected viewport %#v", viewport)
	}
	if ceremony.Event("visitor", "candle", "return").Quiet != true {
		t.Fatal("return should be quiet")
	}
}
