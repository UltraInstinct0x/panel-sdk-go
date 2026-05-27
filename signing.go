package panelsdk

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

func hmacHex(secret string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	_, _ = m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func sha256Hex(body []byte) string {
	s := sha256.Sum256(body)
	return hex.EncodeToString(s[:])
}

func b64u(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func jwtHS256(secret string, payload map[string]interface{}) (string, error) {
	headerB, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}
	payloadB, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal jwt payload: %w", err)
	}
	si := b64u(headerB) + "." + b64u(payloadB)
	m := hmac.New(sha256.New, []byte(secret))
	_, _ = m.Write([]byte(si))
	return si + "." + b64u(m.Sum(nil)), nil
}

func makeSelfSignedAttestation(secret string, body []byte, engineVersion string) (string, error) {
	now := time.Now().Unix()
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", fmt.Errorf("random jti: %w", err)
	}
	payload := map[string]interface{}{
		"jti":            hex.EncodeToString(jti),
		"iat":            now,
		"exp":            now + 300,
		"input_hash":     sha256Hex(body),
		"output_hash":    sha256Hex(body),
		"mode":           "text",
		"engine_version": engineVersion,
	}
	jwt, err := jwtHS256(secret, payload)
	if err != nil {
		return "", fmt.Errorf("jwt sign: %w", err)
	}
	return jwt, nil
}

func scoreCanonicalQuery(ref string, id string) string {
	parts := make([]string, 0, 2)
	if ref != "" {
		parts = append(parts, "ref="+ref)
	}
	if id != "" {
		parts = append(parts, "id="+id)
	}
	sort.Strings(parts)
	return strings.Join(parts, "\n")
}

func scoreCanonicalString(siteKey string, ref string, id string) string {
	return fmt.Sprintf("GET\n/api/units/score\n%s\nsite=%s", scoreCanonicalQuery(ref, id), siteKey)
}
