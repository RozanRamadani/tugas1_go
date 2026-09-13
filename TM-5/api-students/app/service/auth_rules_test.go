package service_test

import (
	"testing"

	"api-students/app/service"
)

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected string
	}{
		{
			name:     "Password terlalu pendek (< 8 karakter)",
			password: "pass1",
			expected: "minimal 8 karakter",
		},
		{
			name:     "Password hanya huruf tanpa angka",
			password: "rahasiaa",
			expected: "harus memuat huruf dan angka",
		},
		{
			name:     "Password hanya angka tanpa huruf",
			password: "123456789",
			expected: "harus memuat huruf dan angka",
		},
		{
			name:     "Password terlalu umum dalam daftar pasaran",
			password: "password123",
			expected: "password terlalu umum",
		},
		{
			name:     "Password kuat yang memenuhi syarat",
			password: "rahasia123",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CheckPasswordStrength(tt.password)
			if got != tt.expected {
				t.Errorf("CheckPasswordStrength(%q) = %q; expected %q", tt.password, got, tt.expected)
			}
		})
	}
}
