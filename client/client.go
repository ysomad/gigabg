package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ysomad/gigabg/api"
)

// Client makes HTTP calls to the game server.
type Client struct {
	client *http.Client
	addr   string // host:port
}

// New creates a client for the given server address (host:port).
// If proxyURL is non-empty, all requests are routed through the given HTTP proxy.
func New(addr, proxyURL string) *Client {
	c := &http.Client{Timeout: 10 * time.Second}
	if proxyURL != "" {
		u, _ := url.Parse(proxyURL)
		c.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
	}
	return &Client{
		client: c,
		addr:   addr,
	}
}

// CreateLobby creates a new lobby and returns the lobby ID.
func (c *Client) CreateLobby(ctx context.Context, maxPlayers int) (string, error) {
	var resp api.CreateLobbyResp
	if err := c.sendRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("http://%s/lobbies", c.addr),
		api.CreateLobbyReq{MaxPlayers: maxPlayers},
		&resp,
	); err != nil {
		return "", err
	}
	return resp.LobbyID, nil
}

func (c *Client) sendRequest(ctx context.Context, method, url string, req, resp any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		dump, dumpErr := httputil.DumpRequestOut(httpReq, true)
		if dumpErr != nil {
			return fmt.Errorf("send: %w, dump: %w", err, dumpErr)
		}

		return fmt.Errorf("send: %w\n%s", err, dump)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return fmt.Errorf("%s %s %d: %w", method, url, httpResp.StatusCode, err)
		}

		return fmt.Errorf("%s %s %d: %s", method, url, httpResp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if err := json.NewDecoder(httpResp.Body).Decode(resp); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return nil
}
