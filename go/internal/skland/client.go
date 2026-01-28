package skland

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// This file contains a minimal HTTP client for the Skland APIs used by the
// original TypeScript project. It does not attempt to be a full reimplementation
// of skland-kit, only the parts needed for daily attendance.

const (
	baseURL = "https://zonai.skland.com"
)

type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Skland client with a default HTTP client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// AppBindingPlayer corresponds to the subset of fields used by the attendance logic.
type AppBindingPlayer struct {
	AppCode     string        `json:"appCode"`
	GameID      int           `json:"gameId"`
	GameName    string        `json:"gameName"`
	UID         string        `json:"uid"`
	DefaultRole *DefaultRole  `json:"defaultRole"`
}

type DefaultRole struct {
	ServerID string `json:"serverId"`
	RoleID   string `json:"roleId"`
	NickName string `json:"nickName"`
}

// BindingItem corresponds to a per-app binding entry.
type BindingItem struct {
	AppCode     string             `json:"appCode"`
	BindingList []AppBindingPlayer `json:"bindingList"`
}

// Attendance status structures for different games.

// ArknightsAttendanceStatus: we only care about records.ts.
type ArknightsAttendanceStatus struct {
	Records []struct {
		TS int64 `json:"ts"`
	} `json:"records"`
}

// EndfieldAttendanceStatus: we only care about hasToday flag.
type EndfieldAttendanceStatus struct {
	HasToday bool `json:"hasToday"`
}

// GameAttendanceStatus is a generic form, we will unmarshal into one of
// the above based on the game.
type GameAttendanceStatus json.RawMessage

// GameAttendanceResult for non-gameId=3.
type GameAttendanceResult struct {
	Awards []struct {
		Resource struct {
			Name string `json:"name"`
		} `json:"resource"`
		Count int `json:"count"`
	} `json:"awards"`
}

// EndfieldAttendanceResult for gameId=3.
type EndfieldAttendanceResult struct {
	AwardIDs []struct {
		ID string `json:"id"`
	} `json:"awardIds"`
	ResourceInfoMap map[string]struct {
		Name string `json:"name"`
	} `json:"resourceInfoMap"`
}

// GrantAuthorizeCode exchanges the token for an authorize code.
func (c *Client) GrantAuthorizeCode(ctx context.Context, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/grant", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("grant authorize code failed: %s", resp.Status)
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Code == "" {
		return "", fmt.Errorf("empty authorize code")
	}
	return body.Code, nil
}

// SignIn exchanges authorize code for a session token.
func (c *Client) SignIn(ctx context.Context, code string) (string, error) {
	// Note: the exact sign-in endpoint and payload depend on Skland API.
	// This is a placeholder structure that should be adjusted to match the
	// real API if needed.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/login", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Set("code", code)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sign in failed: %s", resp.Status)
	}

	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Token == "" {
		return "", fmt.Errorf("empty session token")
	}
	return body.Token, nil
}

// GetBinding returns binding list for the current account.
func (c *Client) GetBinding(ctx context.Context, sessionToken string) ([]BindingItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/game/player/binding", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+sessionToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get binding failed: %s", resp.Status)
	}

	var body struct {
		List []BindingItem `json:"list"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body.List, nil
}

