package evolution

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(bever *httptest.Server) *Client {
	return &Client{
		baseURL:    bever.URL,
		instance:   "test-instance",
		apiKey:     "test-key",
		httpClient: bever.Client(),
	}
}

func TestSendText_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"key": map[string]string{"id": "MSG-123"},
		})
	}))
	defer srv.Close()

	id, err := newTestClient(srv).SendText(context.Background(), "5511999@s.whatsapp.net", "oi")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if id != "MSG-123" {
		t.Errorf("id expected 'MSG-123', got '%s'", id)
	}
}

func TestSendText_StripAtSign(t *testing.T) {
	var capturedBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedBody)
		json.NewEncoder(w).Encode(map[string]any{"key": map[string]string{"id": "x"}})
	}))
	defer srv.Close()

	newTestClient(srv).SendText(context.Background(), "5511999@s.whatsapp.net", "texto")

	if capturedBody["number"] != "5511999" {
		t.Errorf("expected number sem @, got '%s'", capturedBody["number"])
	}
}

func TestSendText_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	_, err := newTestClient(srv).SendText(context.Background(), "5511999", "oi")
	if err == nil {
		t.Fatal("expected error for status 4xx")
	}
}

func TestSendText_InvalidJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not is json"))
	}))
	defer srv.Close()

	_, err := newTestClient(srv).SendText(context.Background(), "5511999", "oi")
	if err == nil {
		t.Fatal("expected error for JSON invalid na resposta")
	}
}

func TestSendText_HeadersAndEndpoint(t *testing.T) {
	var capturedAPIKey, capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAPIKey = r.Header.Get("apikey")
		capturedPath = r.URL.Path
		json.NewEncoder(w).Encode(map[string]any{"key": map[string]string{"id": ""}})
	}))
	defer srv.Close()

	newTestClient(srv).SendText(context.Background(), "5511999", "oi")

	if capturedAPIKey != "test-key" {
		t.Errorf("apikey expected 'test-key', got '%s'", capturedAPIKey)
	}
	if capturedPath != "/monthsage/sendText/test-instance" {
		t.Errorf("path incorrect: %s", capturedPath)
	}
}

func TestFetchImageBase64_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"base64": "abc123=="})
	}))
	defer srv.Close()

	result, err := newTestClient(srv).FetchImageBase64(context.Background(), "jid", true, "MSG-1")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if result != "abc123==" {
		t.Errorf("base64 expected 'abc123==', got '%s'", result)
	}
}

func TestFetchImageBase64_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := newTestClient(srv).FetchImageBase64(context.Background(), "jid", false, "MSG-2")
	if err == nil {
		t.Fatal("expected error for status 4xx")
	}
}

func TestFetchImageBase64_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not is json"))
	}))
	defer srv.Close()

	_, err := newTestClient(srv).FetchImageBase64(context.Background(), "jid", false, "MSG-3")
	if err == nil {
		t.Fatal("expected error for JSON invalid")
	}
}

func TestFetchConnectionState_Open(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"instance": map[string]string{"state": "open"},
		})
	}))
	defer srv.Close()

	state, err := newTestClient(srv).FetchConnectionState(context.Background())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if state != "open" {
		t.Errorf("expected state 'open', got '%s'", state)
	}
}

func TestFetchConnectionState_Close(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"instance": map[string]string{"state": "close"},
		})
	}))
	defer srv.Close()

	state, err := newTestClient(srv).FetchConnectionState(context.Background())
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if state != "close" {
		t.Errorf("expected state 'close', got '%s'", state)
	}
}

func TestFetchConnectionState_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := newTestClient(srv).FetchConnectionState(context.Background())
	if err == nil {
		t.Fatal("expected error for status 4xx")
	}
}

func TestEnsureInstance_Created(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	created, err := newTestClient(srv).EnsureInstance(context.Background(), "5511999999999")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !created {
		t.Error("expected created=true for status 201")
	}
}

func TestEnsureInstance_AlreadyExists(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	created, err := newTestClient(srv).EnsureInstance(context.Background(), "5511999999999")
	if err != nil {
		t.Fatalf("expected success for existing instance, got: %v", err)
	}
	if created {
		t.Error("expected created=false for status 403")
	}
}

func TestEnsureInstance_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := newTestClient(srv).EnsureInstance(context.Background(), "5511999999999")
	if err == nil {
		t.Fatal("expected error for status 500")
	}
}

func TestFetchImageBase64_HeadersAndEndpoint(t *testing.T) {
	var capturedAPIKey, capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAPIKey = r.Header.Get("apikey")
		capturedPath = r.URL.Path
		json.NewEncoder(w).Encode(map[string]string{"base64": ""})
	}))
	defer srv.Close()

	newTestClient(srv).FetchImageBase64(context.Background(), "jid", false, "MSG-4")

	if capturedAPIKey != "test-key" {
		t.Errorf("apikey expected 'test-key', got '%s'", capturedAPIKey)
	}
	if capturedPath != "/chat/getBase64FromMediaMessage/test-instance" {
		t.Errorf("path incorrect: %s", capturedPath)
	}
}


