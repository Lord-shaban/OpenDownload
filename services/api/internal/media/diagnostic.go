package media

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// Optional operator diagnostics contain encrypted stderr only. The private key
// stays off the server. Visitors still receive canonical, credential-free errors.
func diagnosticKey() (*rsa.PublicKey, error) {
	encoded := os.Getenv("OD_DIAGNOSTIC_PUBLIC_KEY")
	if encoded == "" {
		return nil, nil
	}
	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("OD_DIAGNOSTIC_PUBLIC_KEY must be base64 SPKI")
	}
	parsed, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("OD_DIAGNOSTIC_PUBLIC_KEY must be a public RSA key")
	}
	key, ok := parsed.(*rsa.PublicKey)
	if !ok || key.N.BitLen() < 2048 || key.N.BitLen() > 4096 {
		return nil, fmt.Errorf("OD_DIAGNOSTIC_PUBLIC_KEY must be a 2048 to 4096 bit RSA public key")
	}
	return key, nil
}

func encryptedDiagnostic(key *rsa.PublicKey, stderr []byte) (string, error) {
	if len(stderr) > 8192 {
		stderr = stderr[len(stderr)-8192:]
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, key, secret, []byte("OpenDownload diagnostic v1"))
	if err != nil {
		return "", err
	}
	encode := base64.StdEncoding.EncodeToString
	payload, err := json.Marshal(map[string]string{
		"version": "1", "key": encode(wrapped), "nonce": encode(nonce),
		"stderr": encode(gcm.Seal(nil, nonce, stderr, []byte("OpenDownload diagnostic v1"))),
	})
	return string(payload), err
}

func (e *YTDLP) diagnose(phase string, stderr []byte) {
	if e.diagnosticKey == nil {
		return
	}
	payload, err := encryptedDiagnostic(e.diagnosticKey, stderr)
	if err != nil {
		slog.Warn("extractor diagnostic encryption failed", "phase", phase)
		return
	}
	slog.Warn("encrypted extractor diagnostic", "phase", phase, "encrypted", payload)
}
