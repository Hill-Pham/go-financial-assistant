package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type QRProvider interface {
	FetchConnectionState(ctx context.Context) (string, error)
	FetchConnectCode(ctx context.Context) (code, base64 string, err error)
}

type qrcodeHandler struct {
	secret     string
	qrProvider QRProvider
}

func (h *qrcodeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if h.secret == "" {
		http.Error(w, "endpoint disabled: configure ADMIN_SECRET", http.StatusServiceUnavailable)
		return
	}

	if extractBearerSecret(r) != h.secret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	state, err := h.qrProvider.FetchConnectionState(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error checking connection: %v", err), http.StatusInternalServerError)
		return
	}

	if state == "open" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"connected": true,
			"message":   "WhatsApp is already connected",
		})
		return
	}

	_, base64QR, err := h.qrProvider.FetchConnectCode(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error searching for QR code.: %v", err), http.StatusInternalServerError)
		return
	}

	if base64QR == "" {
		http.Error(w, "QR code unavailable, please try again later", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, qrcodeHTML, base64QR)
}

const qrcodeHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta http-equiv="refresh" content="30">
  <title>WhatsApp QR Code</title>
  <style>
    body { font-family: sans-serif; background: #f0f2f5; display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 100vh; margin: 0; }
    .card { background: white; border-radius: 16px; padding: 40px; box-shadow: 0 4px 24px rgba(0,0,0,.1); text-align: center; }
    h2 { color: #111b21; margin: 0 0 8px; }
    p { color: #667781; margin: 8px 0; font-size: 14px; }
    img { display: block; margin: 24px auto; border-radius: 8px; }
    .hint { background: #f0f2f5; border-radius: 8px; padding: 12px 16px; font-size: 13px; color: #54656f; margin-top: 16px; }
  </style>
</head>
<body>
  <div class="card">
	<h2>Connect WhatsApp</h2>
	<p>Scan the QR code below with your WhatsApp</p>
    <img src="%s" width="280" height="280" alt="QR Code WhatsApp">
    <div class="hint">
	WhatsApp &rarr; <strong>Linked devices</strong> &rarr; <strong>Link a device</strong>
    </div>
	<p style="margin-top:16px">This page refreshes automatically every 30 seconds.</p>
  </div>
</body>
</html>`
