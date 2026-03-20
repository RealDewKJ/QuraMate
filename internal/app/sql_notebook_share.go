package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type NotebookShareRelayResult struct {
	Success     bool   `json:"success"`
	Error       string `json:"error"`
	Code        string `json:"code"`
	ExpiresAt   string `json:"expiresAt"`
	PayloadJSON string `json:"payloadJson"`
}

type notebookShareRelayCreateRequest struct {
	Scope       string          `json:"scope"`
	SenderLabel string          `json:"senderLabel,omitempty"`
	Payload     json.RawMessage `json:"payload"`
}

type notebookShareRelayCreateResponse struct {
	Code      string `json:"code"`
	ExpiresAt string `json:"expiresAt"`
	Error     string `json:"error"`
}

type notebookShareRelayResolveResponse struct {
	Code      string          `json:"code"`
	ExpiresAt string          `json:"expiresAt"`
	Payload   json.RawMessage `json:"payload"`
	Error     string          `json:"error"`
}

func (a *App) CreateNotebookShareCode(relayURL string, senderLabel string, payloadJSON string, scope string) NotebookShareRelayResult {
	normalizedURL, err := normalizeNotebookShareRelayURL(relayURL)
	if err != nil {
		return NotebookShareRelayResult{Error: err.Error()}
	}

	trimmedPayload := strings.TrimSpace(payloadJSON)
	if trimmedPayload == "" {
		return NotebookShareRelayResult{Error: "Share payload is required."}
	}

	requestPayload, err := json.Marshal(notebookShareRelayCreateRequest{
		Scope:       strings.ToLower(strings.TrimSpace(scope)),
		SenderLabel: strings.TrimSpace(senderLabel),
		Payload:     json.RawMessage(trimmedPayload),
	})
	if err != nil {
		return NotebookShareRelayResult{Error: fmt.Sprintf("Encode share payload: %v", err)}
	}

	request, err := http.NewRequest(http.MethodPost, normalizedURL+"/api/notebook-shares", bytes.NewReader(requestPayload))
	if err != nil {
		return NotebookShareRelayResult{Error: fmt.Sprintf("Create share request: %v", err)}
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := notebookShareRelayHTTPClient().Do(request)
	if err != nil {
		return NotebookShareRelayResult{Error: fmt.Sprintf("Reach share service: %v", err)}
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return NotebookShareRelayResult{Error: fmt.Sprintf("Read share response: %v", err)}
	}

	var parsed notebookShareRelayCreateResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return NotebookShareRelayResult{Error: "Share service returned an invalid response."}
	}

	if response.StatusCode >= http.StatusBadRequest {
		return NotebookShareRelayResult{Error: firstNonEmpty(parsed.Error, http.StatusText(response.StatusCode))}
	}

	return NotebookShareRelayResult{
		Success:   true,
		Code:      parsed.Code,
		ExpiresAt: parsed.ExpiresAt,
	}
}

func (a *App) ResolveNotebookShareCode(relayURL string, code string) NotebookShareRelayResult {
	normalizedURL, err := normalizeNotebookShareRelayURL(relayURL)
	if err != nil {
		return NotebookShareRelayResult{Error: err.Error()}
	}

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	if normalizedCode == "" {
		return NotebookShareRelayResult{Error: "Share code is required."}
	}

	request, err := http.NewRequest(http.MethodGet, normalizedURL+"/api/notebook-shares/"+normalizedCode, nil)
	if err != nil {
		return NotebookShareRelayResult{Error: fmt.Sprintf("Create resolve request: %v", err)}
	}

	response, err := notebookShareRelayHTTPClient().Do(request)
	if err != nil {
		return NotebookShareRelayResult{Error: fmt.Sprintf("Reach share service: %v", err)}
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return NotebookShareRelayResult{Error: fmt.Sprintf("Read share response: %v", err)}
	}

	var parsed notebookShareRelayResolveResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return NotebookShareRelayResult{Error: "Share service returned an invalid response."}
	}

	if response.StatusCode >= http.StatusBadRequest {
		return NotebookShareRelayResult{Error: firstNonEmpty(parsed.Error, http.StatusText(response.StatusCode))}
	}

	return NotebookShareRelayResult{
		Success:     true,
		Code:        parsed.Code,
		ExpiresAt:   parsed.ExpiresAt,
		PayloadJSON: string(parsed.Payload),
	}
}

func notebookShareRelayHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
	}
}

func normalizeNotebookShareRelayURL(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("Share service URL is required.")
	}

	trimmed = strings.TrimRight(trimmed, "/")
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return "", fmt.Errorf("Share service URL must start with http:// or https://")
	}

	return trimmed, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
