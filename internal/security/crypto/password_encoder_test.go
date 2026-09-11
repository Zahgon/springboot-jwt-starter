package crypto_test

import (
	"testing"

	"github.com/bfwg/springboot-jwt-starter/internal/security/crypto"
)

// seededHash is the hash shipped in import.sql for both demo accounts. It was
// produced at cost 4 by BCryptPasswordEncoder and must keep verifying.
const seededHash = "$2a$04$Vbug2lwwJGrvUXTj6z7ff.97IzVBkrJ1XfApfGNl.Z695zqcnPYra"

func TestMatchesTheSeededHash(t *testing.T) {
	encoder := crypto.NewBCryptPasswordEncoder()

	if !encoder.Matches("123", seededHash) {
		t.Error(`the documented password "123" must verify against the seeded hash`)
	}
	if encoder.Matches("wrong", seededHash) {
		t.Error("a wrong password must not verify")
	}
	if encoder.Matches("", seededHash) {
		t.Error("an empty password must not verify")
	}
}

func TestEncodeThenMatches(t *testing.T) {
	encoder := crypto.NewBCryptPasswordEncoder()

	hash, err := encoder.Encode("s3cret")
	if err != nil {
		t.Fatalf("encoding: %v", err)
	}
	if hash == "s3cret" {
		t.Fatal("the password was stored in the clear")
	}
	if !encoder.Matches("s3cret", hash) {
		t.Error("a freshly encoded password must verify")
	}
	if encoder.Matches("s3cre", hash) {
		t.Error("a near-miss password must not verify")
	}
}

func TestEncodeIsSalted(t *testing.T) {
	encoder := crypto.NewBCryptPasswordEncoder()

	first, err := encoder.Encode("same")
	if err != nil {
		t.Fatalf("encoding: %v", err)
	}
	second, err := encoder.Encode("same")
	if err != nil {
		t.Fatalf("encoding: %v", err)
	}
	if first == second {
		t.Error("encoding the same password twice must produce different hashes")
	}
}

func TestDefaultCostMatchesSpringSecurity(t *testing.T) {
	if crypto.BCryptCost != 10 {
		t.Errorf("default BCrypt cost = %d, want 10", crypto.BCryptCost)
	}
}

func TestMatchesRejectsAMalformedHash(t *testing.T) {
	if crypto.NewBCryptPasswordEncoder().Matches("123", "not-a-bcrypt-hash") {
		t.Error("a malformed hash must never verify")
	}
}
