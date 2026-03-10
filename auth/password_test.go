package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
		{
			name:     "Long password",
			password: "this-is-a-very-long-password-with-many-characters-1234567890",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			// 驗證 hash 不等於原始密碼
			if !tt.wantErr && hash == tt.password {
				t.Error("HashPassword() hash should not equal original password")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "test-password-123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "Incorrect password",
			password: "wrong-password",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalid-hash",
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

func TestHashPasswordConsistency(t *testing.T) {
	password := "test-password"

	// 同一個密碼，每次 hash 應該都不同（因為使用隨機 salt）
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() should generate different hashes for same password (due to salt)")
	}

	// 但兩個 hash 都應該能驗證原始密碼
	if !CheckPassword(password, hash1) {
		t.Error("CheckPassword() failed to verify first hash")
	}
	if !CheckPassword(password, hash2) {
		t.Error("CheckPassword() failed to verify second hash")
	}
}
