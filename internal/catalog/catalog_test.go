package catalog

import "testing"

func TestCatalogSupportsQuietLanguage(t *testing.T) {
	catalog := New()
	message, ok := catalog.Lookup("quiet")
	if !ok || message.Text == "" {
		t.Fatal("quiet message missing")
	}
	if err := ValidateMessage(message); err != nil {
		t.Fatal(err)
	}
	if !catalog.Add(Message{Key: "farewell", Language: "zh-CN", Text: "愿你安好"}) {
		t.Fatal("message not added")
	}
	if _, ok := catalog.Lookup("farewell"); !ok {
		t.Fatal("added message missing")
	}
}
