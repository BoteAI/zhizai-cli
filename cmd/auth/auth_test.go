package auth

import "testing"

func TestMaskKey(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "****"},
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "*bcde"},
		{"DemoApiKey1234", "**********1234"},
	}
	for _, tt := range tests {
		if got := maskKey(tt.in); got != tt.want {
			t.Fatalf("maskKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidateAPIKey(t *testing.T) {
	if err := validateAPIKey(""); err == nil {
		t.Fatal("empty key should fail")
	}
	if err := validateAPIKey("   "); err == nil {
		t.Fatal("blank key should fail")
	}
	if err := validateAPIKey(" DemoKey "); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
}
