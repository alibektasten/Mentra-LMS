package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

var ErrInvalidUploadClaim = errors.New("invalid upload claim")

const uploadClaimLifetime = time.Hour

type uploadClaim struct {
	OrgID   string `json:"org_id"`
	UserID  string `json:"user_id"`
	Path    string `json:"path"`
	Expires int64  `json:"expires"`
}

// IssueUploadClaim grants the uploading user one hour to associate this exact
// new file with an authorized course resource. The token is never stored in
// public DTOs or persisted file paths.
func (s *LocalStorage) IssueUploadClaim(orgID, userID, storedPath string) (string, error) {
	if len(s.claimSecret) == 0 || userID == "" {
		return "", ErrInvalidUploadClaim
	}
	if !s.uploadExists(orgID, storedPath) {
		return "", ErrInvalidUploadClaim
	}
	payload, err := json.Marshal(uploadClaim{OrgID: orgID, UserID: userID, Path: storedPath, Expires: time.Now().Add(uploadClaimLifetime).Unix()})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + base64.RawURLEncoding.EncodeToString(s.signClaim(encoded)), nil
}

func (s *LocalStorage) ValidateUploadClaim(orgID, userID, storedPath, token string) error {
	if len(s.claimSecret) == 0 || token == "" || userID == "" {
		return ErrInvalidUploadClaim
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return ErrInvalidUploadClaim
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, s.signClaim(parts[0])) {
		return ErrInvalidUploadClaim
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrInvalidUploadClaim
	}
	var claim uploadClaim
	if err := json.Unmarshal(payload, &claim); err != nil {
		return ErrInvalidUploadClaim
	}
	if claim.OrgID != orgID || claim.UserID != userID || claim.Path != storedPath || time.Now().Unix() > claim.Expires {
		return ErrInvalidUploadClaim
	}
	if !s.uploadExists(orgID, storedPath) {
		return ErrInvalidUploadClaim
	}
	return nil
}

func (s *LocalStorage) signClaim(encodedPayload string) []byte {
	mac := hmac.New(sha256.New, s.claimSecret)
	_, _ = mac.Write([]byte("mentra-upload-claim-v1:" + encodedPayload))
	return mac.Sum(nil)
}

func (s *LocalStorage) uploadExists(orgID, storedPath string) bool {
	resolved, err := s.ResolvePrivatePath(orgID, storedPath)
	if err != nil {
		return false
	}
	info, err := os.Stat(resolved)
	return err == nil && info.Mode().IsRegular()
}
