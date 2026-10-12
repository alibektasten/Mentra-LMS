package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ailms/backend/internal/application/dto"
	"github.com/ailms/backend/internal/application/usecases"
	"github.com/ailms/backend/internal/delivery/http/middleware"
	"github.com/ailms/backend/internal/infrastructure/config"
	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUC      *usecases.AuthUseCase
	googleOAuth config.GoogleOAuthConfig
	stateSecret string
	httpClient  *http.Client
}

func NewAuthHandler(authUC *usecases.AuthUseCase, googleOAuth config.GoogleOAuthConfig, stateSecret string) *AuthHandler {
	return &AuthHandler{authUC: authUC, googleOAuth: googleOAuth, stateSecret: stateSecret, httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.authUC.Register(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	orgSlug := c.GetHeader("X-Org-Slug")
	if orgSlug == "" {
		c.Error(apperrors.ValidationError("X-Org-Slug header is required"))
		return
	}

	resp, err := h.authUC.Login(c.Request.Context(), req, orgSlug)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *AuthHandler) GoogleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"enabled": h.googleEnabled()}})
}

func (h *AuthHandler) GoogleStart(c *gin.Context) {
	if !h.googleEnabled() {
		c.Error(apperrors.ValidationError("Google sign-in is not configured"))
		return
	}
	orgSlug := strings.TrimSpace(c.Query("org"))
	mode := strings.TrimSpace(c.Query("mode"))
	isSuperAdmin := mode == "super_admin"
	if orgSlug == "" && !isSuperAdmin {
		c.Error(apperrors.ValidationError("organization is required"))
		return
	}
	state, err := h.createGoogleState(orgSlug, isSuperAdmin)
	if err != nil {
		c.Error(err)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("mentra_google_oauth_state", state, 600, "/", "", false, true)
	values := url.Values{
		"client_id":     {h.googleOAuth.ClientID},
		"redirect_uri":  {h.googleOAuth.RedirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"prompt":        {"select_account"},
	}
	c.Redirect(http.StatusFound, "https://accounts.google.com/o/oauth2/v2/auth?"+values.Encode())
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if providerError := c.Query("error"); providerError != "" {
		c.Error(apperrors.UnauthorizedError("Google sign-in was cancelled or denied"))
		return
	}
	state := c.Query("state")
	if cookie, err := c.Cookie("mentra_google_oauth_state"); err != nil || subtle.ConstantTimeCompare([]byte(cookie), []byte(state)) != 1 {
		c.Error(apperrors.UnauthorizedError("invalid OAuth state"))
		return
	}
	stateData, err := h.verifyGoogleState(state)
	if err != nil {
		c.Error(err)
		return
	}
	code := c.Query("code")
	if code == "" {
		c.Error(apperrors.ValidationError("Google authorization code is required"))
		return
	}
	email, err := h.googleEmail(c.Request.Context(), code)
	if err != nil {
		c.Error(err)
		return
	}
	var response *dto.TokenResponse
	if stateData.SuperAdmin {
		response, err = h.authUC.SuperAdminLoginWithOAuth(c.Request.Context(), email)
	} else {
		response, err = h.authUC.LoginWithOAuth(c.Request.Context(), email, stateData.Org)
	}
	if err != nil {
		c.Error(err)
		return
	}
	payload, err := json.Marshal(response)
	if err != nil {
		c.Error(apperrors.InternalError("failed to complete Google sign-in"))
		return
	}
	c.SetCookie("mentra_google_oauth_state", "", -1, "/", "", false, true)
	frontendURL := strings.TrimRight(h.googleOAuth.FrontendURL, "/")
	loginPath := "/login"
	if !stateData.SuperAdmin {
		loginPath = fmt.Sprintf("/%s/login", url.PathEscape(stateData.Org))
	}
	redirectURL := fmt.Sprintf("%s%s#oauth=%s", frontendURL, loginPath, base64.RawURLEncoding.EncodeToString(payload))
	c.Redirect(http.StatusFound, redirectURL)
}

func (h *AuthHandler) googleEnabled() bool {
	return h.googleOAuth.ClientID != "" && h.googleOAuth.ClientSecret != ""
}

func (h *AuthHandler) createGoogleState(orgSlug string, superAdmin bool) (string, error) {
	payload, err := json.Marshal(struct {
		Org        string `json:"org"`
		SuperAdmin bool   `json:"super_admin"`
		Exp        int64  `json:"exp"`
	}{Org: orgSlug, SuperAdmin: superAdmin, Exp: time.Now().Add(10 * time.Minute).Unix()})
	if err != nil {
		return "", apperrors.InternalError("failed to prepare OAuth state")
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(h.stateSecret))
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (h *AuthHandler) verifyGoogleState(state string) (struct {
	Org        string `json:"org"`
	SuperAdmin bool   `json:"super_admin"`
	Exp        int64  `json:"exp"`
}, error) {
	var empty struct {
		Org        string `json:"org"`
		SuperAdmin bool   `json:"super_admin"`
		Exp        int64  `json:"exp"`
	}
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return empty, apperrors.UnauthorizedError("invalid OAuth state")
	}
	mac := hmac.New(sha256.New, []byte(h.stateSecret))
	_, _ = mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) != 1 {
		return empty, apperrors.UnauthorizedError("invalid OAuth state")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return empty, apperrors.UnauthorizedError("invalid OAuth state")
	}
	var payload struct {
		Org        string `json:"org"`
		SuperAdmin bool   `json:"super_admin"`
		Exp        int64  `json:"exp"`
	}
	if err := json.Unmarshal(decoded, &payload); err != nil || (!payload.SuperAdmin && payload.Org == "") || time.Now().Unix() > payload.Exp {
		return empty, apperrors.UnauthorizedError("expired OAuth state")
	}
	return payload, nil
}

func (h *AuthHandler) googleEmail(ctx context.Context, code string) (string, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {h.googleOAuth.ClientID},
		"client_secret": {h.googleOAuth.ClientSecret},
		"redirect_uri":  {h.googleOAuth.RedirectURL},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", apperrors.InternalError("failed to prepare Google token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := h.httpClient.Do(req)
	if err != nil {
		return "", apperrors.UnauthorizedError("could not verify Google sign-in")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", apperrors.UnauthorizedError("Google token exchange failed")
	}
	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&token); err != nil || token.AccessToken == "" {
		return "", apperrors.UnauthorizedError("invalid Google token response")
	}
	profileReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return "", apperrors.InternalError("failed to prepare Google profile request")
	}
	profileReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	profileRes, err := h.httpClient.Do(profileReq)
	if err != nil {
		return "", apperrors.UnauthorizedError("could not retrieve Google profile")
	}
	defer profileRes.Body.Close()
	if profileRes.StatusCode != http.StatusOK {
		return "", apperrors.UnauthorizedError("Google profile verification failed")
	}
	var profile struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := json.NewDecoder(io.LimitReader(profileRes.Body, 1<<20)).Decode(&profile); err != nil || profile.Email == "" || !profile.EmailVerified {
		return "", apperrors.UnauthorizedError("Google account email is not verified")
	}
	return strings.ToLower(profile.Email), nil
}

// suppress unused import warning
var _ = middleware.GetUserID
