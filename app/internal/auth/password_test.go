package auth

import "testing"

func TestPasswordPolicyAndVerification(t *testing.T) {
	password := "Aman-Lokal-123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, password) {
		t.Fatal("kata sandi yang benar tidak terverifikasi")
	}
	if VerifyPassword(hash, "Salah-Lokal-123!") {
		t.Fatal("kata sandi yang salah ikut terverifikasi")
	}
}

func TestPasswordPolicyRejectsWeakPassword(t *testing.T) {
	if err := ValidatePassword("pendek"); err == nil {
		t.Fatal("kata sandi lemah seharusnya ditolak")
	}
}
