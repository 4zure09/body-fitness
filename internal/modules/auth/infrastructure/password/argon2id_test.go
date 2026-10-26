package password

import "testing"

func TestHasherHashAndVerify(t *testing.T) {
	hasher := Hasher{
		memory:      8 * 1024,
		iterations:  1,
		parallelism: 1,
		saltLength:  16,
		keyLength:   32,
	}
	encoded, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	valid, err := hasher.Verify(encoded, "correct horse battery staple")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !valid {
		t.Fatal("expected password to verify")
	}

	valid, err = hasher.Verify(encoded, "wrong password")
	if err != nil {
		t.Fatalf("Verify() wrong password error = %v", err)
	}
	if valid {
		t.Fatal("expected wrong password not to verify")
	}
}

func TestHasherRejectsMalformedHash(t *testing.T) {
	if _, err := New().Verify("not-an-argon-hash", "password"); err == nil {
		t.Fatal("expected malformed hash to return an error")
	}
}
