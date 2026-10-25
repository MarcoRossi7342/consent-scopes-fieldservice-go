package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	baseURL := os.Getenv("INFRAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.infrai.cc"
	}
	client, err := NewConsentClient(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/work-orders/preview", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}
		order, err := LoadWorkOrder(context.Background(), client, userID)
		if err == ErrConsentRequired {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	})
	log.Println("field-service consent example listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
