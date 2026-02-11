package qrcode

import (
	"bytes"
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gamesocial/internal/media"
)

type QRCode struct {
	UUID             string          `json:"uuid"`
	Type             string          `json:"type"`
	Scene            string          `json:"scene,omitempty"`
	Token            string          `json:"token"`
	ImageURL         string          `json:"imageUrl"`
	ImageKey         string          `json:"imageKey,omitempty"`
	ImageContentType string          `json:"imageContentType,omitempty"`
	ImageSizeBytes   int64           `json:"imageSizeBytes,omitempty"`
	Payload          json.RawMessage `json:"payload,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	ExpiresAt        time.Time       `json:"expiresAt"`
	UsedAt           *time.Time      `json:"usedAt,omitempty"`
	Status           string          `json:"status"`
}

type CreateRequest struct {
	Type  string `json:"type"`
	Scene string `json:"scene"`

	UserID uint64 `json:"userId"`

	TTLSeconds int64 `json:"ttlSeconds"`

	OneTime bool `json:"oneTime"`

	Data json.RawMessage `json:"data"`

	PNGSize int `json:"pngSize"`
}

type UseRequest struct {
	Token string `json:"token"`
}

type UseResult struct {
	UUID      string          `json:"uuid"`
	Type      string          `json:"type"`
	Scene     string          `json:"scene,omitempty"`
	UserID    uint64          `json:"userId,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	IssuedAt  time.Time       `json:"issuedAt"`
	ExpiresAt time.Time       `json:"expiresAt"`
	UsedAt    *time.Time      `json:"usedAt,omitempty"`
	OneTime   bool            `json:"oneTime"`
}

type Service interface {
	Create(ctx context.Context, req CreateRequest) (QRCode, error)
	Verify(ctx context.Context, token string) (media.QRPayload, error)
	Use(ctx context.Context, uid uint64, req UseRequest) (UseResult, error)
}

type service struct {
	db *sql.DB

	store media.ServerStore

	pubKey  *rsa.PublicKey
	privKey *rsa.PrivateKey

	defaultTTLSeconds int64
	defaultPNGSize    int
}

func NewService(db *sql.DB, store media.ServerStore, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey, defaultTTLSeconds int64, defaultPNGSize int) Service {
	if defaultTTLSeconds <= 0 {
		defaultTTLSeconds = 300
	}
	if defaultPNGSize <= 0 {
		defaultPNGSize = 320
	}
	return &service{
		db:                db,
		store:             store,
		pubKey:            pubKey,
		privKey:           privKey,
		defaultTTLSeconds: defaultTTLSeconds,
		defaultPNGSize:    defaultPNGSize,
	}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (QRCode, error) {
	if s.db == nil {
		return QRCode{}, errors.New("database disabled")
	}
	if s.store == nil {
		return QRCode{}, errors.New("media store not configured")
	}
	if s.pubKey == nil {
		return QRCode{}, errors.New("qrcode public key not configured")
	}
	if s.privKey == nil {
		return QRCode{}, errors.New("qrcode private key not configured")
	}

	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		return QRCode{}, errors.New("type is empty")
	}
	req.Scene = strings.Trim(strings.TrimSpace(req.Scene), "/")
	if req.TTLSeconds <= 0 {
		req.TTLSeconds = s.defaultTTLSeconds
	}
	if req.TTLSeconds > 24*3600 {
		return QRCode{}, errors.New("ttlSeconds too large")
	}
	if req.PNGSize <= 0 {
		req.PNGSize = s.defaultPNGSize
	}

	uuid, err := media.NewUUIDv4()
	if err != nil {
		return QRCode{}, err
	}
	now := time.Now()
	exp := now.Add(time.Duration(req.TTLSeconds) * time.Second)
	payload := media.QRPayload{
		V:         1,
		UUID:      uuid,
		Type:      req.Type,
		Scene:     req.Scene,
		UserID:    req.UserID,
		IssuedAt:  now.Unix(),
		ExpiresAt: exp.Unix(),
		Data:      nil,
	}
	if len(req.Data) != 0 {
		payload.Data = req.Data
	}
	token, png, err := media.GenerateEncryptedQR(ctx, s.pubKey, payload, req.PNGSize)
	if err != nil {
		return QRCode{}, err
	}

	contentType := "image/png"
	uploadRes, err := s.store.Upload(ctx, bytes.NewReader(png), contentType, "qrcode-"+uuid+".png")
	if err != nil {
		return QRCode{}, err
	}

	payloadJSON, _ := json.Marshal(payload)
	status := "ACTIVE"
	if !req.OneTime {
		status = "ACTIVE_MULTI"
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO qr_code (uuid, purpose, scene, user_id, token, image_url, image_key, image_content_type, image_size_bytes, payload_json, status, created_at, expires_at, used_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?, NULL, NOW())
	`, uuid, req.Type, req.Scene, nullableUint64(req.UserID), token, uploadRes.URL, uploadRes.Key, contentType, int64(len(png)), payloadJSON, status, exp)
	if err != nil {
		return QRCode{}, err
	}

	out := QRCode{
		UUID:             uuid,
		Type:             req.Type,
		Scene:            req.Scene,
		Token:            token,
		ImageURL:         uploadRes.URL,
		ImageKey:         uploadRes.Key,
		ImageContentType: contentType,
		ImageSizeBytes:   int64(len(png)),
		Payload:          payloadJSON,
		CreatedAt:        now,
		ExpiresAt:        exp,
		Status:           status,
	}
	return out, nil
}

func (s *service) Verify(ctx context.Context, token string) (media.QRPayload, error) {
	_ = ctx
	if s.privKey == nil {
		return media.QRPayload{}, errors.New("qrcode private key not configured")
	}
	p, err := media.DecryptTokenToPayload(s.privKey, token)
	if err != nil {
		return media.QRPayload{}, err
	}
	if err := p.Validate(time.Now()); err != nil {
		return media.QRPayload{}, err
	}
	return p, nil
}

func (s *service) Use(ctx context.Context, uid uint64, req UseRequest) (UseResult, error) {
	if s.db == nil {
		return UseResult{}, errors.New("database disabled")
	}
	if uid == 0 {
		return UseResult{}, errors.New("invalid uid")
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return UseResult{}, errors.New("token is empty")
	}

	payload, err := s.Verify(ctx, token)
	if err != nil {
		return UseResult{}, err
	}
	if payload.UserID != 0 && payload.UserID != uid {
		return UseResult{}, errors.New("qrcode not belongs to current user")
	}

	issuedAt := time.Unix(payload.IssuedAt, 0)
	expiresAt := time.Unix(payload.ExpiresAt, 0)

	var status string
	var usedAt sql.NullTime
	row := s.db.QueryRowContext(ctx, `
		SELECT status, used_at
		FROM qr_code
		WHERE uuid = ?
		LIMIT 1
	`, payload.UUID)
	if err := row.Scan(&status, &usedAt); err != nil {
		if err == sql.ErrNoRows {
			return UseResult{}, errors.New("qrcode not found")
		}
		return UseResult{}, err
	}

	isOneTime := status == "ACTIVE"
	if isOneTime {
		now := time.Now()
		res, err := s.db.ExecContext(ctx, `
			UPDATE qr_code
			SET status = 'USED', used_at = ?, updated_at = NOW()
			WHERE uuid = ? AND status = 'ACTIVE'
		`, now, payload.UUID)
		if err != nil {
			return UseResult{}, err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return UseResult{}, errors.New("qrcode already used")
		}
		usedAt = sql.NullTime{Time: now, Valid: true}
		status = "USED"
	}

	if strings.EqualFold(payload.Type, "CHECKIN") {
		_, _ = s.db.ExecContext(ctx, `
			INSERT INTO checkin_log (user_id, checkin_at, source)
			VALUES (?, NOW(), ?)
		`, uid, "QR")
	}

	var usedAtPtr *time.Time
	if usedAt.Valid {
		t := usedAt.Time
		usedAtPtr = &t
	}

	out := UseResult{
		UUID:      payload.UUID,
		Type:      payload.Type,
		Scene:     payload.Scene,
		UserID:    payload.UserID,
		Data:      payload.Data,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		UsedAt:    usedAtPtr,
		OneTime:   isOneTime,
	}
	return out, nil
}

func nullableUint64(v uint64) any {
	if v == 0 {
		return nil
	}
	return v
}
