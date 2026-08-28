package protocol

import (
	"bytes"
	"encoding/json"
	"memorialcandle/internal/memorial"
	"memorialcandle/internal/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMemorialWorkflow(t *testing.T) {
	db, err := store.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := NewServer(memorial.NewService(db, "http-test"), nil).Routes()
	body, _ := json.Marshal(CreateRequest{Title: "烛光", Dedication: "记得"})
	req := httptest.NewRequest(http.MethodPost, "/memorials", bytes.NewReader(body))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("create status %d", res.Code)
	}
	if res.Header().Get("Content-Type") != "application/json" {
		t.Fatal("missing content type")
	}
}
