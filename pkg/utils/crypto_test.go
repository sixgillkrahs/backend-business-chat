package utils

import (
	"testing"
)

func TestAESGCM_EncryptAndDecrypt(t *testing.T) {
	secretKey := "my-super-secret-jwt-key-32byteslong!!"
	plaintext := "gho_16C7e42F292c6912E7710c838347Ae178B4a"

	encrypted, err := EncryptAESGCM(plaintext, secretKey)
	if err != nil {
		t.Fatalf("EncryptAESGCM failed: %v", err)
	}

	if encrypted == "" || encrypted == plaintext {
		t.Fatalf("Encrypted ciphertext must be non-empty and different from plaintext")
	}

	decrypted, err := DecryptAESGCM(encrypted, secretKey)
	if err != nil {
		t.Fatalf("DecryptAESGCM failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("expected decrypted '%s', got '%s'", plaintext, decrypted)
	}
}

func TestAESGCM_WrongKeyFails(t *testing.T) {
	secretKey1 := "key-one-12345678901234567890"
	secretKey2 := "key-two-12345678901234567890"
	plaintext := "gho_secret_token_value"

	encrypted, err := EncryptAESGCM(plaintext, secretKey1)
	if err != nil {
		t.Fatalf("EncryptAESGCM failed: %v", err)
	}

	_, err = DecryptAESGCM(encrypted, secretKey2)
	if err == nil {
		t.Fatalf("expected decryption failure with wrong key, but got success")
	}
}

func TestAESGCM_TamperedCiphertextFails(t *testing.T) {
	secretKey := "secure-key"
	plaintext := "confidential_token"

	encrypted, err := EncryptAESGCM(plaintext, secretKey)
	if err != nil {
		t.Fatalf("EncryptAESGCM failed: %v", err)
	}

	// Tamper with ciphertext by modifying a character
	tampered := []byte(encrypted)
	if tampered[5] == 'A' {
		tampered[5] = 'B'
	} else {
		tampered[5] = 'A'
	}

	_, err = DecryptAESGCM(string(tampered), secretKey)
	if err == nil {
		t.Fatalf("expected authentication failure with tampered ciphertext")
	}
}

func TestAESGCM_EmptyInputs(t *testing.T) {
	secretKey := "some-key"

	_, err := EncryptAESGCM("", secretKey)
	if err != ErrEmptyPlaintext {
		t.Fatalf("expected ErrEmptyPlaintext, got %v", err)
	}

	_, err = DecryptAESGCM("", secretKey)
	if err != ErrInvalidCiphertext {
		t.Fatalf("expected ErrInvalidCiphertext, got %v", err)
	}
}
