package app

import (
	"encoding/json"
	"testing"
)

func TestNormalizeAIProvider(t *testing.T) {
	t.Parallel()

	provider, err := normalizeAIProvider("  OpenAI-Compatible_1  ")
	if err != nil {
		t.Fatalf("expected provider to normalize, got error: %v", err)
	}

	if provider != "openai-compatible_1" {
		t.Fatalf("expected normalized provider, got %q", provider)
	}
}

func TestNormalizeAIProviderRejectsInvalidCharacters(t *testing.T) {
	t.Parallel()

	if _, err := normalizeAIProvider("openai!"); err == nil {
		t.Fatal("expected invalid provider to fail")
	}
}

func TestSanitizeUserSettingsForStorage(t *testing.T) {
	t.Parallel()

	input := `{"ai":{"provider":"openai","apiKey":"secret","apiKeys":{"openai":"secret"},"providerState":{"openai":{"status":"configured"}}}}`
	sanitized := sanitizeUserSettingsForStorage(input)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(sanitized), &parsed); err != nil {
		t.Fatalf("expected sanitized json to remain valid: %v", err)
	}

	aiMap, ok := parsed["ai"].(map[string]any)
	if !ok {
		t.Fatal("expected ai settings to exist")
	}

	if _, exists := aiMap["apiKey"]; exists {
		t.Fatal("expected apiKey to be removed")
	}
	if _, exists := aiMap["apiKeys"]; exists {
		t.Fatal("expected apiKeys to be removed")
	}
	if _, exists := aiMap["providerState"]; !exists {
		t.Fatal("expected providerState to be preserved")
	}
}
