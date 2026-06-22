package main

import (
	"context"
	"errors"
	"net/http"
	"time"
)

func (c *cli) login(force bool) error {
	global, _ := loadGlobalConfig()
	if !force && global.AccessToken != "" && time.Until(global.ExpiresAt) > time.Minute {
		return nil
	}

	repoPath, repoRemote := c.repoMetadata()
	startReq := map[string]string{
		"repo_path":   repoPath,
		"repo_remote": repoRemote,
	}
	var start authStartResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/cli/auth/start", startReq, false, &start); err != nil {
		return err
	}

	c.println("Open this URL to authenticate Prilog:")
	c.println(start.AuthURL)
	_ = c.openURL(start.AuthURL)

	interval := time.Duration(start.Interval) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	deadline := time.Now().Add(time.Duration(start.ExpiresIn) * time.Second)
	if start.ExpiresIn <= 0 {
		deadline = time.Now().Add(10 * time.Minute)
	}

	for time.Now().Before(deadline) {
		time.Sleep(interval)

		var poll authPollResponse
		err := c.doJSON(context.Background(), http.MethodPost, "/cli/auth/poll", map[string]string{"device_code": start.DeviceCode}, false, &poll)
		if err != nil {
			var apiErr apiError
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusAccepted {
				continue
			}
			return err
		}
		if poll.AccessToken == "" || poll.RefreshToken == "" {
			continue
		}

		global.APIURL = c.apiURL
		global.AccessToken = poll.AccessToken
		global.RefreshToken = poll.RefreshToken
		global.TokenType = firstNonEmpty(poll.TokenType, "Bearer")
		global.ExpiresAt = time.Now().Add(time.Duration(poll.ExpiresIn) * time.Second)
		global.UserID = poll.UserID
		global.CLISessionID = poll.CLISessionID
		global.DashboardURL = poll.DashboardURL
		if err := saveGlobalConfig(global); err != nil {
			return err
		}

		c.println("Authenticated with Prilog.")
		return nil
	}
	return errors.New("authentication timed out")
}

func (c *cli) ensureAuth() error {
	global, _ := loadGlobalConfig()
	if global.APIURL != "" {
		c.apiURL = trimConfiguredURL(global.APIURL)
	}
	if global.AccessToken == "" {
		return c.login(false)
	}
	if time.Until(global.ExpiresAt) > time.Minute {
		return nil
	}
	if global.RefreshToken == "" {
		return c.login(true)
	}

	var response authRefreshResponse
	if err := c.doJSON(context.Background(), http.MethodPost, "/cli/auth/refresh", map[string]string{"refresh_token": global.RefreshToken}, false, &response); err != nil {
		return c.login(true)
	}

	global.AccessToken = response.AccessToken
	global.TokenType = firstNonEmpty(response.TokenType, "Bearer")
	global.ExpiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
	if response.UserID != "" {
		global.UserID = response.UserID
	}
	return saveGlobalConfig(global)
}
