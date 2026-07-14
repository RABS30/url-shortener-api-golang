package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"shorter-url/internal/domain"
	"shorter-url/internal/middleware"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

type oauthGoogleHandler struct {
	Service      domain.UserService
	Config       *oauth2.Config
	CookieConfig CookieConfig
}

func NewOauthGoogleHandler(service domain.UserService, config *oauth2.Config, cookieConfig CookieConfig) *oauthGoogleHandler {
	return &oauthGoogleHandler{
		Service:      service,
		Config:       config,
		CookieConfig: cookieConfig,
	}
}

func (h *oauthGoogleHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	oauthState := generateStateOauthCookie(w)
	u := h.Config.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (h *oauthGoogleHandler) Callback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()
	loginPageURL := "/login"

	oauthState, err := r.Cookie("oauthState")
	if err != nil {
		http.Redirect(w, r, loginPageURL+"?error=invalid_oauth_request", http.StatusTemporaryRedirect)
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("google oauth callback: failed to get oauth state cookie: %w", err))
		}
		return
	}

	state := r.URL.Query().Get("state")
	if state != oauthState.Value {
		http.Redirect(w, r, loginPageURL+"?error=state_mismatch", http.StatusTemporaryRedirect)
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("google oauth callback: state mismatch"))
		}
		return
	}

	if errParam := r.URL.Query().Get("error"); errParam != "" {
		http.Redirect(w, r, loginPageURL+"?error=access_denied", http.StatusTemporaryRedirect)
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("google oauth callback: user denied access: %s", errParam))
		}
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, loginPageURL+"?error=missing_code", http.StatusTemporaryRedirect)
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("google oauth callback: missing code parameter"))
		}
		return
	}

	tokenResponse, err := getTokenfromGoogle(ctx, code, h.Config)
	if err != nil {
		http.Redirect(w, r, loginPageURL+"?error=token_exchange_failed", http.StatusTemporaryRedirect)
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("google oauth callback: %w", err))
		}
		return
	}

	userInfo, err := ValidateGoogleIdToken(ctx, tokenResponse, h.Config)
	if err != nil {
		http.Redirect(w, r, loginPageURL+"?error=invalid_id_token", http.StatusTemporaryRedirect)
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("google oauth callback: %w", err))
		}
		return
	}

	jwtToken, err := h.Service.LoginWithGoogle(ctx, userInfo)
	if err != nil {
		http.Redirect(w, r, loginPageURL+"?error=login_failed", http.StatusTemporaryRedirect)
		if wrapper, ok := w.(*middleware.LogResponseWriter); ok {
			wrapper.WriteError(fmt.Errorf("google oauth callback: %w", err))
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthState",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.CookieConfig.Secure,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Domain:   h.CookieConfig.Domain,
		Path:     h.CookieConfig.Path,
		MaxAge:   h.CookieConfig.MaxAge,
		HttpOnly: true,
		Secure:   h.CookieConfig.Secure,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "http://localhost:5173/auth/callback", http.StatusTemporaryRedirect)

}

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(5 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := &http.Cookie{
		Name:     "oauthState",
		Value:    state,
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	return state
}

func getTokenfromGoogle(ctx context.Context, code string, googleOauthConfig *oauth2.Config) (*oauth2.Token, error) {
	response, err := googleOauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get user data from google: %w", err)
	}

	return response, nil
}

func ValidateGoogleIdToken(ctx context.Context, token *oauth2.Token, googleOauthConfig *oauth2.Config) (*domain.GoogleUserInfo, error) {
	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("invalid id_token")
	}

	payload, err := idtoken.Validate(ctx, rawIdToken, googleOauthConfig.ClientID)
	if err != nil {
		return nil, err
	}

	claimsJSON, err := json.Marshal(payload.Claims)
	if err != nil {
		return nil, err
	}

	var userInfo domain.GoogleUserInfo

	if err := json.Unmarshal(claimsJSON, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
