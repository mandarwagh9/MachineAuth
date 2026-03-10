package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestWriteAuthorizeError_WithRedirectURI(t *testing.T) {
	handler := &OAuthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?redirect_uri=https://example.com&state=xyz", nil)
	w := httptest.NewRecorder()

	handler.writeAuthorizeError(w, req, "invalid_request", "test error", "xyz")

	if w.Code != http.StatusFound {
		t.Errorf("expected redirect status 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header")
	}

	parsed, _ := url.Parse(location)
	if parsed.Query().Get("error") != "invalid_request" {
		t.Errorf("expected error in query, got %s", location)
	}
}

func TestWriteAuthorizeError_WithoutRedirectURI(t *testing.T) {
	handler := &OAuthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	w := httptest.NewRecorder()

	handler.writeAuthorizeError(w, req, "invalid_request", "test error", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestWriteAuthorizeError_WithState(t *testing.T) {
	handler := &OAuthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?redirect_uri=https://example.com&state=abc123", nil)
	w := httptest.NewRecorder()

	handler.writeAuthorizeError(w, req, "access_denied", "user denied", "abc123")

	if w.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	parsed, _ := url.Parse(location)

	if parsed.Query().Get("state") != "abc123" {
		t.Errorf("expected state to be preserved, got %s", location)
	}

	if parsed.Query().Get("error") != "access_denied" {
		t.Errorf("expected error in query, got %s", location)
	}
}

func TestWriteAuthorizeError_InvalidRedirectURI(t *testing.T) {
	handler := &OAuthHandler{}

	// No redirect_uri in query - should return 400
	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize", nil)
	w := httptest.NewRecorder()

	handler.writeAuthorizeError(w, req, "invalid_request", "test error", "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHtmlEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>", "&lt;script&gt;"},
		{"a & b", "a &amp; b"},
		{"\"quotes\"", "&quot;quotes&quot;"},
		{"'single'", "&#39;single&#39;"},
		{"normal", "normal"},
		{"", ""},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"test & <test>", "test &amp; &lt;test&gt;"},
	}

	for _, tt := range tests {
		result := htmlEscape(tt.input)
		if result != tt.expected {
			t.Errorf("htmlEscape(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestWriteAuthorizeError_ErrorDescriptionInResponse(t *testing.T) {
	handler := &OAuthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?redirect_uri=https://example.com", nil)
	w := httptest.NewRecorder()

	handler.writeAuthorizeError(w, req, "invalid_request", "missing required parameter", "")

	location := w.Header().Get("Location")
	parsed, _ := url.Parse(location)

	if parsed.Query().Get("error_description") != "missing required parameter" {
		t.Errorf("expected error_description in query, got %s", location)
	}
}
