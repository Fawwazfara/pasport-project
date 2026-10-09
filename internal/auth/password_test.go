package auth

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("rahasia123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}

	if !VerifyPassword("rahasia123", hash) {
		t.Error("kata sandi benar seharusnya cocok")
	}
	if VerifyPassword("salah", hash) {
		t.Error("kata sandi salah tidak boleh cocok")
	}
	if VerifyPassword("rahasia123", "bukan-hash") {
		t.Error("hash tidak valid tidak boleh cocok")
	}
}
