package media

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

type QRPayload struct {
	// V 是 payload 版本号，用于后续结构演进与兼容判断。
	V int `json:"v"`

	// UUID 是二维码的唯一标识，通常对应一条后端落库记录的主键/唯一键。
	UUID string `json:"uuid"`

	// Type 表示二维码用途/业务类型（例如：CHECKIN、REDEEM 等）。
	Type string `json:"type"`

	// Scene 用于区分同一 Type 下的不同业务场景（可选）。
	Scene string `json:"scene,omitempty"`

	// UserID 为可选的签发用户标识；若二维码需要与用户绑定，可写入并在核销时校验。
	UserID uint64 `json:"userId,omitempty"`

	// IssuedAt/ExpiresAt 使用 Unix 秒时间戳表示签发时间与过期时间。
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`

	// Data 存放业务自定义字段（JSON 对象/数组）。建议仅放核销所需的关键 ID，
	// 业务详情由服务端通过 UUID/ID 再查库获取，避免二维码内容膨胀或暴露敏感信息。
	Data json.RawMessage `json:"data,omitempty"`
}

// Validate 用于校验 payload 的基本合法性与是否过期。
func (p QRPayload) Validate(now time.Time) error {
	if p.V <= 0 {
		return errors.New("invalid payload version")
	}
	if strings.TrimSpace(p.UUID) == "" {
		return errors.New("uuid is empty")
	}
	if strings.TrimSpace(p.Type) == "" {
		return errors.New("type is empty")
	}
	if p.IssuedAt <= 0 || p.ExpiresAt <= 0 {
		return errors.New("invalid iat/exp")
	}
	if p.ExpiresAt < p.IssuedAt {
		return errors.New("exp earlier than iat")
	}
	if now.Unix() > p.ExpiresAt {
		return errors.New("qrcode expired")
	}
	return nil
}

// NewUUIDv4 生成 RFC 4122 的 UUIDv4（随机）。
func NewUUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	buf := make([]byte, 32)
	hex.Encode(buf[0:32], b)
	return string(buf), nil
}

// EncryptPayloadToToken 将 payload 加密成不透明 token，适合放入二维码内容。
//
// 加密采用混合方案：
// - 随机生成 AES-256 key + 12 字节 nonce，使用 AES-GCM 加密 payload JSON
// - 使用 RSA-OAEP(SHA-256) 加密 AES key
//
// token 格式：
//
//	v1.<base64url(encKey)>.<base64url(nonce)>.<base64url(ciphertext)>
//
// 其中 base64url 使用 RawURLEncoding（无 padding），更适合在 URL/二维码中携带。
func EncryptPayloadToToken(pub *rsa.PublicKey, payload QRPayload) (string, error) {
	if pub == nil {
		return "", errors.New("rsa public key is nil")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// 生成一次性对称密钥（32 bytes = AES-256）。
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return "", err
	}

	// 生成 GCM nonce（推荐 12 bytes）。
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// AES-GCM 密文内包含认证标签（tag），用于完整性校验。
	ciphertext := gcm.Seal(nil, nonce, raw, nil)

	// 使用 RSA-OAEP 加密对称密钥，避免直接用 RSA 加密大 payload。
	encKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, aesKey, nil)
	if err != nil {
		return "", err
	}

	b64 := base64.RawURLEncoding
	token := strings.Join([]string{
		"v1",
		b64.EncodeToString(encKey),
		b64.EncodeToString(nonce),
		b64.EncodeToString(ciphertext),
	}, ".")
	return token, nil
}

// DecryptTokenToPayload 将 token 解密回 payload。
// 要求 token 格式符合 EncryptPayloadToToken 的输出，并使用匹配的 RSA 私钥。
func DecryptTokenToPayload(priv *rsa.PrivateKey, token string) (QRPayload, error) {
	if priv == nil {
		return QRPayload{}, errors.New("rsa private key is nil")
	}
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) != 4 {
		return QRPayload{}, errors.New("invalid token format")
	}
	if parts[0] != "v1" {
		return QRPayload{}, errors.New("unsupported token version")
	}
	b64 := base64.RawURLEncoding
	encKey, err := b64.DecodeString(parts[1])
	if err != nil {
		return QRPayload{}, errors.New("invalid token key part")
	}
	nonce, err := b64.DecodeString(parts[2])
	if err != nil {
		return QRPayload{}, errors.New("invalid token nonce part")
	}
	ciphertext, err := b64.DecodeString(parts[3])
	if err != nil {
		return QRPayload{}, errors.New("invalid token ciphertext part")
	}

	// 先解密得到 AES-256 key。
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, encKey, nil)
	if err != nil {
		return QRPayload{}, errors.New("decrypt token key failed")
	}
	if len(aesKey) != 32 {
		return QRPayload{}, errors.New("invalid aes key length")
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return QRPayload{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return QRPayload{}, err
	}

	// GCM Open 会同时验证密文完整性；校验失败会返回错误。
	raw, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return QRPayload{}, errors.New("decrypt token payload failed")
	}

	var payload QRPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return QRPayload{}, errors.New("invalid payload json")
	}
	return payload, nil
}

// GenerateQRPNG 将内容编码为二维码 PNG。
// content 一般为 EncryptPayloadToToken 生成的 token；size 为像素边长，<=0 默认 256。
func GenerateQRPNG(content string, size int) ([]byte, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("content is empty")
	}
	if size <= 0 {
		size = 256
	}
	return qrcode.Encode(content, qrcode.Medium, size)
}

// ParseRSAPrivateKeyFromPEM 从 PEM 文本解析 RSA 私钥，支持 PKCS#1 与 PKCS#8。
func ParseRSAPrivateKeyFromPEM(pemText string) (*rsa.PrivateKey, error) {
	pemText = strings.TrimSpace(pemText)
	if pemText == "" {
		return nil, errors.New("empty private key pem")
	}
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("invalid private key pem")
	}

	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	k2, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rk, ok := k2.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not rsa")
	}
	return rk, nil
}

// ParseRSAPublicKeyFromPEM 从 PEM 文本解析 RSA 公钥，支持：
// - PKIX Public Key
// - X.509 Certificate
// - PKCS#1 Public Key
func ParseRSAPublicKeyFromPEM(pemText string) (*rsa.PublicKey, error) {
	pemText = strings.TrimSpace(pemText)
	if pemText == "" {
		return nil, errors.New("empty public key pem")
	}
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, errors.New("invalid public key pem")
	}

	if pk, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rk, ok := pk.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("public key is not rsa")
		}
		return rk, nil
	}

	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		rk, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("certificate public key is not rsa")
		}
		return rk, nil
	}

	if rk, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return rk, nil
	}

	return nil, errors.New("unsupported public key pem")
}

// DecodePEMFromEnv 支持两种环境变量形式：
// - 直接传入 PEM 文本（包含 "BEGIN "）
// - 传入 PEM 文本的 base64 编码（便于写入 .env / K8s secret）
func DecodePEMFromEnv(v string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", errors.New("empty pem")
	}
	if strings.Contains(v, "BEGIN ") {
		return v, nil
	}
	raw, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return "", fmt.Errorf("invalid base64 pem: %w", err)
	}
	return strings.TrimSpace(string(raw)), nil
}

func GenerateEncryptedQR(ctx context.Context, pub *rsa.PublicKey, payload QRPayload, size int) (string, []byte, error) {
	_ = ctx
	token, err := EncryptPayloadToToken(pub, payload)
	if err != nil {
		return "", nil, err
	}
	png, err := GenerateQRPNG(token, size)
	if err != nil {
		return "", nil, err
	}
	return token, png, nil
}
