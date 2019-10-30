package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mullvad/message-queue/api/handler"
)

func TestRecovery(t *testing.T) {
	ts := httptest.NewServer(handler.Recovery(http.HandlerFunc(panicHandler)))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("wrong status code %#v", res.StatusCode)
	}
}

func panicHandler(rw http.ResponseWriter, req *http.Request) {
	panic("panicking handler")
}
