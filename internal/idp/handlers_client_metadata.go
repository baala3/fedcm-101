package idp

import (
	"encoding/json"
	"net/http"
)

// handleClientMetadata serves per-RP metadata (privacy policy / terms of
// service links) shown in the browser's FedCM consent UI. Fetched
// uncredentialed.
func (s *Server) handleClientMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", RPOrigin)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"privacy_policy_url":   RPOrigin + "/privacy",
		"terms_of_service_url": RPOrigin + "/tos",
	})
}
