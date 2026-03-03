package services

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"machineauth/internal/models"
)

func TestParseScope(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", nil},
		{"single", "read", []string{"read"}},
		{"multiple space", "read write", []string{"read", "write"}},
		{"multiple comma", "read,write", []string{"read", "write"}},
		{"multiple mixed", "read, write, execute", []string{"read", "write", "execute"}},
		{"with spaces", "  read  ", []string{"read"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseScope(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseScope(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseScope(%q)[%d] = %v, want %v", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestFilterScopes(t *testing.T) {
	allowed := []string{"read", "write", "execute"}

	tests := []struct {
		name      string
		requested []string
		expected  []string
	}{
		{"all allowed", []string{"read", "write"}, []string{"read", "write"}},
		{"some allowed", []string{"read", "admin"}, []string{"read"}},
		{"none allowed", []string{"admin", "superuser"}, []string{}},
		{"empty", []string{}, []string{}},
		{"wildcard not implemented", []string{"*"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterScopes(allowed, tt.requested)
			if len(result) != len(tt.expected) {
				t.Errorf("filterScopes(%v, %v) = %v, want %v", allowed, tt.requested, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("filterScopes[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestGenerateSecureSecret(t *testing.T) {
	secret1 := generateSecureSecret(32)
	secret2 := generateSecureSecret(32)

	if len(secret1) != 32 {
		t.Errorf("generateSecureSecret(32) length = %d, want 32", len(secret1))
	}

	if secret1 == secret2 {
		t.Error("generateSecureSecret should generate unique secrets")
	}
}

func TestAgentModel(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	agent := models.Agent{
		ID:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		Name:      "test-agent",
		ClientID:  "client-123",
		Scopes:    []string{"read", "write"},
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: &expiresAt,
	}

	if agent.Name != "test-agent" {
		t.Errorf("Agent.Name = %v, want test-agent", agent.Name)
	}

	if !agent.IsActive {
		t.Error("Agent.IsActive should be true")
	}

	if len(agent.Scopes) != 2 {
		t.Errorf("Agent.Scopes length = %d, want 2", len(agent.Scopes))
	}
}

func TestTokenResponseModel(t *testing.T) {
	resp := models.TokenResponse{
		AccessToken: "eyJhbGc...",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       "read write",
		IssuedAt:    time.Now().Unix(),
	}

	if resp.TokenType != "Bearer" {
		t.Errorf("TokenResponse.TokenType = %v, want Bearer", resp.TokenType)
	}

	if resp.ExpiresIn != 3600 {
		t.Errorf("TokenResponse.ExpiresIn = %d, want 3600", resp.ExpiresIn)
	}
}

func TestIntrospectResponseModel(t *testing.T) {
	resp := models.IntrospectResponse{
		Active:    true,
		Scope:     "read",
		ClientID:  "client-123",
		Exp:       time.Now().Add(1 * time.Hour).Unix(),
		Iat:       time.Now().Unix(),
		TokenType: "Bearer",
	}

	if !resp.Active {
		t.Error("IntrospectResponse.Active should be true")
	}

	if resp.ClientID != "client-123" {
		t.Errorf("IntrospectResponse.ClientID = %v, want client-123", resp.ClientID)
	}
}

func TestIntrospectResponseInactive(t *testing.T) {
	resp := models.IntrospectResponse{
		Active: false,
		Reason: "expired",
	}

	if resp.Active {
		t.Error("IntrospectResponse.Active should be false")
	}

	if resp.Reason != "expired" {
		t.Errorf("IntrospectResponse.Reason = %v, want expired", resp.Reason)
	}
}
