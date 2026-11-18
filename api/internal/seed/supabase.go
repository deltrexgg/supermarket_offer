package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// adminClient talks to Supabase Auth's admin API with the secret/service-role key.
type adminClient struct {
	base, key string
	http      *http.Client
}

func newAdminClient(url, key string) *adminClient {
	return &adminClient{base: strings.TrimRight(url, "/"), key: key, http: &http.Client{Timeout: 20 * time.Second}}
}

func (a *adminClient) do(ctx context.Context, method, path string, body any) (int, []byte, error) {
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

// ensureUser returns the Supabase user id for email, creating a confirmed user
// when none exists yet.
func (a *adminClient) ensureUser(ctx context.Context, email, password string) (string, error) {
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
			return "", fmt.Errorf("create %s: unexpected response", email)
		}
		return u.ID, nil
	}
	// Already registered: find it in the user list.
	for page := 1; page <= 20; page++ {
		status, out, err = a.do(ctx, http.MethodGet, fmt.Sprintf("/auth/v1/admin/users?page=%d&per_page=100", page), nil)
		if err != nil {
			return "", err
		}
		if status != http.StatusOK {
			return "", fmt.Errorf("list users: status %d: %s", status, out)
		}
		var list struct {
			Users []struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"users"`
		}
		if err := json.Unmarshal(out, &list); err != nil {
			return "", err
		}
		for _, u := range list.Users {
			if strings.EqualFold(u.Email, email) {
				return u.ID, nil
			}
		}
		if len(list.Users) < 100 {
			break
		}
	}
	return "", fmt.Errorf("create %s failed (status %d: %s) and user not found", email, status, out)
}
