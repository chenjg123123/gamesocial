package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

func ParseTokenV1(token string, secret []byte, now time.Time) (uint64, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, fmt.Errorf("empty token")
	}
	if len(secret) == 0 {
		return 0, fmt.Errorf("empty secret")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid token format")
	}

	enc := base64.RawURLEncoding
	headerBytes, err := enc.DecodeString(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid header")
	}
	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return 0, fmt.Errorf("invalid header")
	}
	if header.Alg != "HS256" {
		return 0, fmt.Errorf("unsupported alg")
	}

	payloadBytes, err := enc.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid payload")
	}
	var payload struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return 0, fmt.Errorf("invalid payload")
	}
	if payload.Exp <= 0 {
		return 0, fmt.Errorf("invalid exp")
	}
	if now.Unix() > payload.Exp {
		return 0, fmt.Errorf("token expired")
	}

	userID, err := strconv.ParseUint(payload.Sub, 10, 64)
	if err != nil || userID == 0 {
		return 0, fmt.Errorf("invalid sub")
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)
	gotSig, err := enc.DecodeString(parts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid signature")
	}
	if !hmac.Equal(gotSig, expectedSig) {
		return 0, fmt.Errorf("invalid signature")
	}
	return userID, nil
}
