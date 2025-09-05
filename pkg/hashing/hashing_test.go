package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func expectedHMACHex(data, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func TestHashByKey_MatchesStdLibHMAC_SHA256(t *testing.T) {
	data := "hello world"
	key := "secret-key-123"

	want := expectedHMACHex(data, key)
	got := HashByKey(data, key)

	if got != want {
		t.Fatalf("unexpected hmac: got %q, want %q", got, want)
	}
}

func TestHashByKey_DifferentKeysProduceDifferentHashes(t *testing.T) {
	data := "payload"
	key1 := "key-a"
	key2 := "key-b"

	h1 := HashByKey(data, key1)
	h2 := HashByKey(data, key2)

	if h1 == h2 {
		t.Fatalf("hashes should differ for different keys: %q == %q", h1, h2)
	}
}

func TestHashByKey_EmptyData(t *testing.T) {
	data := ""
	key := "k"

	want := expectedHMACHex(data, key)
	got := HashByKey(data, key)
	if got != want {
		t.Fatalf("unexpected hmac for empty data: got %q, want %q", got, want)
	}
}

func TestHashByKey_EmptyKey(t *testing.T) {
	data := "abc"
	key := ""

	want := expectedHMACHex(data, key)
	got := HashByKey(data, key)
	if got != want {
		t.Fatalf("unexpected hmac for empty key: got %q, want %q", got, want)
	}
}

func TestHashByKey_EmptyDataAndKey(t *testing.T) {
	data := ""
	key := ""

	want := expectedHMACHex(data, key)
	got := HashByKey(data, key)
	if got != want {
		t.Fatalf("unexpected hmac for empty data/key: got %q, want %q", got, want)
	}
}
