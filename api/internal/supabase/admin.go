// Package supabase calls Supabase Auth's admin API (creating and deleting users)
// with the secret / service-role key. It is used only by the API, never the browser.
package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrEmailExists = errors.New("email already registered")

type Admin struct {
	base, key string
	http      *http.Client
}

func NewAdmin(url, key string) *Admin {
	return &Admin{base: strings.TrimRight(url, "/"), key: key, http: &http.Client{Timeout: 20 * time.Second}}
}

func (a *Admin) do(ctx context.Context, method, path string, body any) (int, []byte, error) {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, a.base+path, rd)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("apikey", a.key)
	req.Header.Set("Authorization", "Bearer "+a.key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, out, nil
}

// CreateUser makes a confirmed user and returns its id.
func (a *Admin) CreateUser(ctx context.Context, email, password string) (string, error) {
	status, out, err := a.do(ctx, http.MethodPost, "/auth/v1/admin/users", map[string]any{
		"email": email, "password": password, "email_confirm": true,
	})
	if err != nil {
		return "", err
	}
	if status == http.StatusOK || status == http.StatusCreated {
		var u struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(out, &u); err != nil || u.ID == "" {
			return "", errors.New("supabase: unexpected create-user response")
		}
		return u.ID, nil
	}
	var e struct {
		ErrorCode string `json:"error_code"`
	}
	_ = json.Unmarshal(out, &e)
	if e.ErrorCode == "email_exists" || e.ErrorCode == "user_already_exists" {
		return "", ErrEmailExists
	}
	return "", fmt.Errorf("supabase create user: status %d", status)
}

func (a *Admin) DeleteUser(ctx context.Context, id string) error {
	status, _, err := a.do(ctx, http.MethodDelete, "/auth/v1/admin/users/"+id, nil)
	if err != nil {
		return err
	}
	if status >= 300 && status != http.StatusNotFound {
		return fmt.Errorf("supabase delete user: status %d", status)
	}
	return nil
}
