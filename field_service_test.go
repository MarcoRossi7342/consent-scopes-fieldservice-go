package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoadWorkOrderRequiresPhotoConsent(t *testing.T) {
	tests := []struct {
		name    string
		granted bool
		wantErr error
	}{
		{name: "granted user receives field data", granted: true},
		{name: "revoked user is denied", granted: false, wantErr: ErrConsentRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v1/auth/consent/check/tech-7/work_order_photos" {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				if tt.granted {
					w.Write([]byte(`{"ok":true,"data":{"result":true},"error":null,"metadata":{}}`))
				} else {
					w.Write([]byte(`{"ok":true,"data":{"result":false},"error":null,"metadata":{}}`))
				}
			}))
			defer server.Close()
			client := &ConsentClient{baseURL: server.URL, key: "test-key", http: server.Client(), sleep: func(_ time.Duration) {}}
			client.sleep = func(_ time.Duration) {}
			got, err := LoadWorkOrder(context.Background(), client, "tech-7")
			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.DispatchStatus != "en_route" {
				t.Fatalf("unexpected order: %+v", got)
			}
		})
	}
}
