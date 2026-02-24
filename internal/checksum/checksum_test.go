package checksum

import (
	"testing"
)

const testMerchantKey = "kbzk1DSbJiV_O3p5" // 16-byte test key

func TestGenerateAndVerify(t *testing.T) {
	body := `{"mid":"TEST_MID","orderId":"ORDER_001"}`

	sig, err := Generate(body, testMerchantKey)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if sig == "" {
		t.Fatal("Generate returned empty string")
	}

	valid, err := Verify(body, sig, testMerchantKey)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid {
		t.Error("Verify returned false for valid signature")
	}
}

func TestVerify_TamperedBody(t *testing.T) {
	body := `{"mid":"TEST_MID","orderId":"ORDER_001"}`

	sig, err := Generate(body, testMerchantKey)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	valid, err := Verify(`{"mid":"WRONG"}`, sig, testMerchantKey)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if valid {
		t.Error("Verify returned true for tampered body")
	}
}

func TestVerify_InvalidBase64(t *testing.T) {
	body := `{"mid":"TEST_MID","orderId":"ORDER_001"}`
	_, err := Verify(body, "not-valid-base64!!!", testMerchantKey)
	if err == nil {
		t.Error("expected error for invalid base64 signature")
	}
}

func TestGenerateWithSalt_Deterministic(t *testing.T) {
	body := `{"mid":"TEST_MID","orderId":"ORDER_001"}`
	salt := "abcd1234"

	sig1, err := generateWithSalt(body, testMerchantKey, salt)
	if err != nil {
		t.Fatalf("generateWithSalt failed: %v", err)
	}
	sig2, err := generateWithSalt(body, testMerchantKey, salt)
	if err != nil {
		t.Fatalf("generateWithSalt failed: %v", err)
	}
	if sig1 != sig2 {
		t.Errorf("same inputs produced different signatures: %q vs %q", sig1, sig2)
	}

	valid, err := Verify(body, sig1, testMerchantKey)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid {
		t.Error("deterministic signature failed verification")
	}
}

func TestVerify_WrongKey(t *testing.T) {
	body := `{"mid":"TEST_MID","orderId":"ORDER_001"}`

	sig, err := Generate(body, testMerchantKey)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	valid, _ := Verify(body, sig, "differentKey1234")
	if valid {
		t.Error("Verify returned true with wrong key")
	}
}

func TestEmptyBody(t *testing.T) {
	sig, err := Generate("", testMerchantKey)
	if err != nil {
		t.Fatalf("Generate failed for empty body: %v", err)
	}

	valid, err := Verify("", sig, testMerchantKey)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid {
		t.Error("empty body signature failed verification")
	}
}

func TestPKCS7PadUnpad(t *testing.T) {
	tests := []struct {
		input []byte
		bs    int
	}{
		{[]byte("hello"), 16},
		{[]byte("exactly16chars!!"), 16},
		{[]byte("a"), 16},
		{[]byte(""), 16},
	}

	for _, tt := range tests {
		padded := pkcs7Pad(tt.input, tt.bs)
		if len(padded)%tt.bs != 0 {
			t.Errorf("padded length %d not a multiple of %d", len(padded), tt.bs)
		}
		unpadded, err := pkcs7Unpad(padded)
		if err != nil {
			t.Errorf("pkcs7Unpad failed: %v", err)
			continue
		}
		if string(unpadded) != string(tt.input) {
			t.Errorf("expected %q, got %q", tt.input, unpadded)
		}
	}
}

func TestAESEncryptDecrypt(t *testing.T) {
	key := []byte(testMerchantKey)
	plaintext := []byte("test plaintext data for encryption")

	encrypted, err := aesEncrypt(plaintext, key)
	if err != nil {
		t.Fatalf("aesEncrypt failed: %v", err)
	}

	decrypted, err := aesDecrypt(encrypted, key)
	if err != nil {
		t.Fatalf("aesDecrypt failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted %q != original %q", decrypted, plaintext)
	}
}
