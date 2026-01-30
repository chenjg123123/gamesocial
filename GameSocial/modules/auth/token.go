package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// MakeTokenV1 生成一个 JWT token（HS256）。
func MakeTokenV1(userID uint64, expiresAt time.Time, secret []byte) (string, error) {
	if userID == 0 {
		return "", fmt.Errorf("invalid userID")
	}
	if len(secret) == 0 {
		return "", fmt.Errorf("empty secret")
	}

	exp := expiresAt.Unix()
	if exp <= 0 {
		return "", fmt.Errorf("invalid expiresAt")
	}

	headerJSON, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(map[string]any{
		"sub": strconv.FormatUint(userID, 10),
		"exp": exp,
	})
	if err != nil {
		return "", err
	}

	enc := base64.RawURLEncoding
	header := enc.EncodeToString(headerJSON)
	payload := enc.EncodeToString(payloadJSON)
	signingInput := header + "." + payload

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signingInput))
	sig := enc.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sig, nil
}
