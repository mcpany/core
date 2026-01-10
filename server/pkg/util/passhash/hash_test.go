// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package passhash

import (
	"testing"
)

func TestPassword(t *testing.T) {
	password := "secret123"
	hash, err := Password(password)
	if err != nil {
		t.Fatalf("Password() error = %v", err)
	}
	if hash == "" {
		t.Error("Password() returned empty string")
	}
	if hash == password {
		t.Error("Password() returned unhashed password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "secret123"
	hash, err := Password(password)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "RightPassword",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "WrongPassword",
			password: "wrong",
			hash:     hash,
			want:     false,
		},
		{
			name:     "EmptyPassword",
			password: "",
			hash:     hash,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPassword(tt.password, tt.hash); got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
