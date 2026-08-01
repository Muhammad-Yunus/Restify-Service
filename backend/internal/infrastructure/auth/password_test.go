package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordServiceHashProducesNonEmptyHash(t *testing.T) {
	p := NewPasswordService(12)

	hash, err := p.Hash("S3curePass!")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash == "" {
		t.Fatal("Hash() returned empty hash")
	}

	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
		t.Errorf("Hash() = %q, want bcrypt prefix", hash)
	}
}

func TestPasswordServiceVerifyMatchingPassword(t *testing.T) {
	p := NewPasswordService(12)

	hash, err := p.Hash("S3curePass!")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !p.Verify("S3curePass!", hash) {
		t.Fatal("Verify() = false, want true for matching password")
	}
}

func TestPasswordServiceVerifyWrongPassword(t *testing.T) {
	p := NewPasswordService(12)

	hash, err := p.Hash("S3curePass!")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if p.Verify("WrongPass1", hash) {
		t.Fatal("Verify() = true, want false for wrong password")
	}
}

func TestPasswordServiceNeedsRehashWhenCostTooLow(t *testing.T) {
	p := NewPasswordService(12)

	lowHash, err := NewPasswordService(10).Hash("S3curePass!")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if !p.NeedsRehash(lowHash) {
		t.Fatal("NeedsRehash() = false, want true for low-cost hash")
	}
}

func TestPasswordServiceNeedsRehashFalseForSameCost(t *testing.T) {
	p := NewPasswordService(12)

	hash, err := p.Hash("S3curePass!")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if p.NeedsRehash(hash) {
		t.Fatal("NeedsRehash() = true, want false for same-cost hash")
	}
}

func TestPasswordServiceDefaultCostIsTwelve(t *testing.T) {
	p := NewPasswordService(0)

	hash, err := p.Hash("S3curePass!")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt cost: %v", err)
	}

	if cost != 12 {
		t.Errorf("default cost = %d, want 12", cost)
	}
}
