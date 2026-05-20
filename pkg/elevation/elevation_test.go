package elevation

import "testing"

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		subs     []string
		expected bool
	}{
		{"Match first", "access is denied", []string{"access is denied", "admin"}, true},
		{"Match second", "needs admin rights", []string{"access is denied", "admin"}, true},
		{"No match", "everything is fine", []string{"access is denied", "admin"}, false},
		{"Empty string", "", []string{"access is denied"}, false},
		{"Empty subs", "some text", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.s, tt.subs)
			if got != tt.expected {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tt.s, tt.subs, got, tt.expected)
			}
		})
	}
}

func TestCheckAdminRequirement(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{"Error code 0x80070005", "Error: 0x80070005 Access is denied", true},
		{"Access is denied lowercase", "error: access is denied for this operation", true},
		{"Access is denied uppercase", "ERROR: ACCESS IS DENIED", true},
		{"Admin privileges text", "Administrator privileges required to run this", true},
		{"Admin privileges mixed case", "administrator privileges required", true},
		{"No admin issue", "Everything completed successfully", false},
		{"Empty output", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckAdminRequirement(tt.output)
			if got != tt.expected {
				t.Errorf("CheckAdminRequirement(%q) = %v, want %v", tt.output, got, tt.expected)
			}
		})
	}
}
