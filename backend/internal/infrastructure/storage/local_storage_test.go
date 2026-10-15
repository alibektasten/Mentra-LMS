package storage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorageResolvePrivatePathRejectsTraversalAndOtherTenant(t *testing.T) {
	store := NewLocalStorage(t.TempDir())
	orgID, anotherOrgID := uuid.New(), uuid.New()

	resolved, err := store.ResolvePrivatePath(orgID.String(), "uploads/"+orgID.String()+"/report.pdf")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(store.BaseDir, orgID.String(), "report.pdf"), resolved)

	for _, storedPath := range []string{
		"uploads/" + orgID.String() + "/../secret.pdf",
		"uploads/" + anotherOrgID.String() + "/report.pdf",
		"/etc/passwd",
		"uploads/" + orgID.String() + "/nested/report.pdf",
	} {
		_, err := store.ResolvePrivatePath(orgID.String(), storedPath)
		assert.True(t, errors.Is(err, ErrInvalidStoredPath), storedPath)
	}
}

func TestLocalStorageSaveUsesTenantDirectoryAndPrivatePermissions(t *testing.T) {
	store := NewLocalStorage(t.TempDir())
	orgID := uuid.New()
	sourcePath := filepath.Join(t.TempDir(), "report.txt")
	require.NoError(t, os.WriteFile(sourcePath, []byte("private report"), 0600))
	source, err := os.Open(sourcePath)
	require.NoError(t, err)
	defer source.Close()

	storedPath, err := store.Save(source, &multipart.FileHeader{Filename: "report.txt", Size: 14}, orgID.String())
	require.NoError(t, err)
	resolved, err := store.ResolvePrivatePath(orgID.String(), storedPath)
	require.NoError(t, err)
	info, err := os.Stat(resolved)
	require.NoError(t, err)
	assert.Zero(t, info.Mode().Perm()&0077)

	content, err := io.ReadAll(mustOpen(t, resolved))
	require.NoError(t, err)
	assert.Equal(t, "private report", string(content))
}

func TestLocalStorageSaveRejectsInvalidOrganization(t *testing.T) {
	store := NewLocalStorage(t.TempDir())
	sourcePath := filepath.Join(t.TempDir(), "report.txt")
	require.NoError(t, os.WriteFile(sourcePath, []byte("report"), 0600))
	source, err := os.Open(sourcePath)
	require.NoError(t, err)
	defer source.Close()

	_, err = store.Save(source, &multipart.FileHeader{Filename: "report.txt", Size: 6}, "../../other-org")
	assert.True(t, errors.Is(err, ErrInvalidUpload))
}

func TestAvatarNamespaceDoesNotResolvePrivateUpload(t *testing.T) {
	store := NewLocalStorage(t.TempDir())
	orgID, userID := uuid.New(), uuid.New()
	orgDir := filepath.Join(store.BaseDir, orgID.String())
	require.NoError(t, os.MkdirAll(orgDir, 0700))
	privatePath := "uploads/" + orgID.String() + "/photo.png"
	require.NoError(t, os.WriteFile(filepath.Join(orgDir, "photo.png"), []byte("image data"), 0600))
	_, err := store.ResolvePublicAvatarPath(userID.String(), privatePath)
	assert.ErrorIs(t, err, ErrInvalidStoredPath)

	avatarPath, err := store.PromoteAvatar(orgID.String(), userID.String(), privatePath)
	require.NoError(t, err)
	resolved, err := store.ResolvePublicAvatarPath(userID.String(), avatarPath)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(store.BaseDir, "avatars", userID.String(), "photo.png"), resolved)
	_, err = store.ResolvePublicAvatarPath(uuid.NewString(), avatarPath)
	assert.ErrorIs(t, err, ErrInvalidStoredPath)
	_, err = os.Stat(filepath.Join(orgDir, "photo.png"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestUploadClaimBindsFileToUploaderAndTenant(t *testing.T) {
	store := NewLocalStorage(t.TempDir(), "test-upload-secret")
	orgID, otherOrgID, uploaderID, otherUserID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	orgDir := filepath.Join(store.BaseDir, orgID.String())
	require.NoError(t, os.MkdirAll(orgDir, 0700))
	storedPath := "uploads/" + orgID.String() + "/lesson.pdf"
	require.NoError(t, os.WriteFile(filepath.Join(orgDir, "lesson.pdf"), []byte("%PDF-1.4"), 0600))
	token, err := store.IssueUploadClaim(orgID.String(), uploaderID.String(), storedPath)
	require.NoError(t, err)
	require.NoError(t, store.ValidateUploadClaim(orgID.String(), uploaderID.String(), storedPath, token))
	for _, mismatch := range []struct{ orgID, userID, path, token string }{
		{orgID.String(), otherUserID.String(), storedPath, token},
		{otherOrgID.String(), uploaderID.String(), storedPath, token},
		{orgID.String(), uploaderID.String(), "uploads/" + orgID.String() + "/other.pdf", token},
		{orgID.String(), uploaderID.String(), storedPath, ""},
		{orgID.String(), uploaderID.String(), storedPath, token + "tamper"},
	} {
		assert.ErrorIs(t, store.ValidateUploadClaim(mismatch.orgID, mismatch.userID, mismatch.path, mismatch.token), ErrInvalidUploadClaim)
	}
	expiredBody, err := json.Marshal(uploadClaim{OrgID: orgID.String(), UserID: uploaderID.String(), Path: storedPath, Expires: time.Now().Add(-time.Minute).Unix()})
	require.NoError(t, err)
	expiredPayload := base64.RawURLEncoding.EncodeToString(expiredBody)
	expiredToken := expiredPayload + "." + base64.RawURLEncoding.EncodeToString(store.signClaim(expiredPayload))
	assert.ErrorIs(t, store.ValidateUploadClaim(orgID.String(), uploaderID.String(), storedPath, expiredToken), ErrInvalidUploadClaim)
	// A copied token cannot be used once its file is gone.
	require.NoError(t, os.Remove(filepath.Join(orgDir, "lesson.pdf")))
	assert.ErrorIs(t, store.ValidateUploadClaim(orgID.String(), uploaderID.String(), storedPath, token), ErrInvalidUploadClaim)
}

func mustOpen(t *testing.T, filePath string) *os.File {
	t.Helper()
	file, err := os.Open(filePath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })
	return file
}
