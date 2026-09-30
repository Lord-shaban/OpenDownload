package media

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestOperatorDiagnosticEncryptsAndBoundsStderr(t *testing.T) {
	private, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKIXPublicKey(&private.PublicKey)
	t.Setenv("OD_DIAGNOSTIC_PUBLIC_KEY", base64.StdEncoding.EncodeToString(der))
	key, err := diagnosticKey()
	if err != nil {
		t.Fatal(err)
	}
	stderr := []byte(strings.Repeat("private-source-data", 700))
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	defer slog.SetDefault(previous)
	e := &YTDLP{diagnosticKey: key}
	e.diagnose("analysis", stderr)
	if strings.Contains(output.String(), "private-source-data") {
		t.Fatal("diagnostic logged plaintext stderr")
	}
	var logged struct{ Encrypted string }
	if err := json.Unmarshal(output.Bytes(), &logged); err != nil {
		t.Fatal(err)
	}
	var envelope map[string]string
	if err := json.Unmarshal([]byte(logged.Encrypted), &envelope); err != nil {
		t.Fatal(err)
	}
	decode := func(name string) []byte {
		value, err := base64.StdEncoding.DecodeString(envelope[name])
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	secret, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, private, decode("key"), []byte("OpenDownload diagnostic v1"))
	if err != nil {
		t.Fatal(err)
	}
	block, _ := aes.NewCipher(secret)
	gcm, _ := cipher.NewGCM(block)
	plain, err := gcm.Open(nil, decode("nonce"), decode("stderr"), []byte("OpenDownload diagnostic v1"))
	if err != nil || !bytes.Equal(plain, stderr[len(stderr)-8192:]) {
		t.Fatal("diagnostic payload did not round-trip or was unbounded")
	}
}

func TestOperatorDiagnosticDefaultsOffAndRejectsInvalidKeys(t *testing.T) {
	t.Setenv("OD_DIAGNOSTIC_PUBLIC_KEY", "")
	if key, err := diagnosticKey(); err != nil || key != nil {
		t.Fatal("diagnostics should default off")
	}
	t.Setenv("OD_DIAGNOSTIC_PUBLIC_KEY", "not-a-public-key")
	if _, err := diagnosticKey(); err == nil {
		t.Fatal("malformed diagnostic key was accepted")
	}
}
