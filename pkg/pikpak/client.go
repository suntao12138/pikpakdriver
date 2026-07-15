package pikpak

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ── Client ─────────────────────────────────────────────────────────────────

type Client struct {
	httpClient   *http.Client
	authBaseURL  string
	driveBaseURL string
	clientID     string
	clientSecret string
	deviceID     string
	token        *SessionToken
}

// NewLoginClient creates a bare PikPak client suitable for login only.
// proxyURL: optional HTTP proxy URL (e.g. "http://127.0.0.1:7890"). Empty = no proxy.
func NewLoginClient(email string, proxyURL string) *Client {
	deviceID := fmt.Sprintf("%x", md5.Sum([]byte(email)))
	return &Client{
		httpClient:   newHTTPClient(proxyURL),
		authBaseURL:  DefaultAuthBaseURL,
		driveBaseURL: DefaultDriveBaseURL,
		clientID:     DefaultClientID,
		clientSecret: DefaultClientSecret,
		deviceID:     deviceID,
	}
}

// NewClient creates a new PikPak client from its own session file.
// Priority: CLI --proxy flag > config.json proxy > no proxy.
func NewClient(cliProxy string) (*Client, error) {
	c := &Client{
		httpClient:   newHTTPClient(""),
		authBaseURL:  DefaultAuthBaseURL,
		driveBaseURL: DefaultDriveBaseURL,
		clientID:     DefaultClientID,
		clientSecret: DefaultClientSecret,
	}

	// Determine proxy: CLI flag > config file > none
	proxyURL := cliProxy
	if proxyURL == "" {
		if creds, err := LoadCredentials(); err == nil {
			proxyURL = creds.Proxy
		}
	}
	if proxyURL != "" {
		c.httpClient = newHTTPClient(proxyURL)
	}

	token, err := LoadSession()
	if err == nil && token != nil {
		c.token = token
		c.deviceID = token.DeviceID
		if !c.token.IsExpired() {
			return c, nil
		}
		if err := c.refreshToken(); err == nil {
			return c, nil
		}
	}

	creds, credErr := LoadCredentials()
	if credErr != nil {
		return nil, fmt.Errorf("no session at %s and no credentials at %s: %w",
			SessionPath(), CredentialsPath(), credErr)
	}

	c.deviceID = fmt.Sprintf("%x", md5.Sum([]byte(creds.Email)))
	if loginErr := c.autoLogin(creds.Email, creds.Password); loginErr != nil {
		return nil, fmt.Errorf("auto-login failed: %w", loginErr)
	}
	return c, nil
}

// newHTTPClient creates an HTTP client. If proxyURL is non-empty, it uses
// that as the HTTP proxy; otherwise no proxy is configured.
func newHTTPClient(proxyURL string) *http.Client {
	if proxyURL == "" {
		return &http.Client{Timeout: 120 * time.Second}
	}
	return &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				return url.Parse(proxyURL)
			},
		},
	}
}

// ── Login ─────────────────────────────────────────────────────────────────

func (c *Client) autoLogin(email, password string) error {
	if err := c.signIn(email, password, ""); err == nil {
		return nil
	}
	captcha, err := c.initCaptcha(email)
	if err != nil {
		return fmt.Errorf("login failed (captcha unavailable): %w", err)
	}
	if captcha.CaptchaToken != "" && captcha.CaptchaToken != "null" {
		if err := c.signIn(email, password, captcha.CaptchaToken); err == nil {
			return nil
		}
	}
	return fmt.Errorf("CAPTCHA required — open %s in browser", captcha.URL)
}

func (c *Client) Login(email, password string) (*CaptchaInitResponse, error) {
	err := c.signIn(email, password, "")
	if err == nil {
		return nil, nil
	}
	if isRegionRestricted(err) {
		return nil, &RegionError{Err: err}
	}
	captcha, captchaErr := c.initCaptcha(email)
	if captchaErr != nil {
		return nil, fmt.Errorf("signin failed: %v (captcha: %w)", err, captchaErr)
	}
	if captcha.CaptchaToken != "" && captcha.CaptchaToken != "null" {
		if retryErr := c.signIn(email, password, captcha.CaptchaToken); retryErr == nil {
			return nil, nil
		}
	}
	return captcha, fmt.Errorf("CAPTCHA required: open %s", captcha.URL)
}

func (c *Client) initCaptcha(email string) (*CaptchaInitResponse, error) {
	payload := map[string]interface{}{
		"action":    "POST:" + c.authBaseURL + "/v1/auth/signin",
		"client_id": c.clientID,
		"device_id": c.deviceID,
		"meta":      map[string]string{"username": email},
	}
	raw, err := c.rawPOST(c.authBaseURL, "/v1/shield/captcha/init", payload)
	if err != nil {
		return nil, err
	}
	var captcha CaptchaInitResponse
	json.Unmarshal(raw, &captcha)
	return &captcha, nil
}

func (c *Client) signIn(email, password, captchaToken string) error {
	payload := map[string]string{
		"username":      email,
		"password":      password,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"captcha_token": captchaToken,
		"grant_type":    "password",
	}
	raw, err := c.rawPOST(c.authBaseURL, "/v1/auth/signin", payload)
	if err != nil {
		return err
	}
	var signin SigninResponse
	if err := json.Unmarshal(raw, &signin); err != nil {
		return err
	}
	now := time.Now().Unix()
	c.token = &SessionToken{
		AccessToken:   signin.AccessToken,
		RefreshToken:  signin.RefreshToken,
		ExpiresAtUnix: now + int64(signin.ExpiresIn-300),
		DeviceID:      c.deviceID,
		UserID:        signin.Sub,
	}
	return SaveSession(c.token)
}

// ── Token ─────────────────────────────────────────────────────────────────

func (c *Client) refreshToken() error {
	payload := map[string]string{
		"client_id":      c.clientID,
		"client_secret":  c.clientSecret,
		"refresh_token":  c.token.RefreshToken,
		"grant_type":     "refresh_token",
	}
	raw, err := c.rawPOST(c.authBaseURL, "/v1/auth/token", payload)
	if err != nil {
		return err
	}
	var tr TokenRefreshResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return err
	}
	c.token.AccessToken = tr.AccessToken
	if tr.RefreshToken != "" {
		c.token.RefreshToken = tr.RefreshToken
	}
	c.token.ExpiresAtUnix = time.Now().Unix() + int64(tr.ExpiresIn-300)
	if tr.Sub != "" {
		c.token.UserID = tr.Sub
	}
	return SaveSession(c.token)
}

func (c *Client) AccessToken() (string, error) {
	if c.token == nil {
		return "", fmt.Errorf("not logged in")
	}
	if c.token.IsExpired() {
		if err := c.refreshToken(); err != nil {
			return "", err
		}
	}
	return c.token.AccessToken, nil
}

// ── HTTP Infrastructure ───────────────────────────────────────────────────

func (c *Client) rawPOST(baseURL, path string, body interface{}) ([]byte, error) {
	data, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", DefaultUserAgent)
	if c.deviceID != "" {
		req.Header.Set("X-Device-Id", c.deviceID)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}
	return raw, nil
}

func (c *Client) doRequest(method, baseURL, path string, query map[string]string, body []byte) ([]byte, error) {
	token, err := c.AccessToken()
	if err != nil {
		return nil, err
	}
	fullURL := baseURL + path
	if len(query) > 0 {
		vals := url.Values{}
		for k, v := range query {
			vals.Set(k, v)
		}
		fullURL += "?" + vals.Encode()
	}
	req, err := http.NewRequest(method, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Content-Type", "application/json")
	if c.deviceID != "" {
		req.Header.Set("X-Device-Id", c.deviceID)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	// 401 → auto refresh token once
	if resp.StatusCode == 401 {
		if err := c.refreshToken(); err != nil {
			return nil, fmt.Errorf("auth failed: %w", err)
		}
		token, _ = c.AccessToken()
		req.Header.Set("Authorization", "Bearer "+token)
		resp2, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("retry failed: %w", err)
		}
		defer resp2.Body.Close()
		raw, _ = io.ReadAll(resp2.Body)
		if resp2.StatusCode >= 400 {
			return nil, fmt.Errorf("API error (HTTP %d): %s", resp2.StatusCode, string(raw))
		}
		return raw, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(raw))
	}
	return raw, nil
}

func (c *Client) driveGET(path string, query map[string]string) ([]byte, error) {
	return c.doRequest("GET", c.driveBaseURL, path, query, nil)
}
func (c *Client) drivePOST(path string, body interface{}) ([]byte, error) {
	data, _ := json.Marshal(body)
	return c.doRequest("POST", c.driveBaseURL, path, nil, data)
}
func (c *Client) drivePATCH(path string, body interface{}) ([]byte, error) {
	data, _ := json.Marshal(body)
	return c.doRequest("PATCH", c.driveBaseURL, path, nil, data)
}
func (c *Client) driveDELETE(path string, query ...map[string]string) ([]byte, error) {
	q := map[string]string{}
	if len(query) > 0 {
		q = query[0]
	}
	return c.doRequest("DELETE", c.driveBaseURL, path, q, nil)
}
func (c *Client) authGET(path string) ([]byte, error) {
	return c.doRequest("GET", c.authBaseURL, path, nil, nil)
}

// ── Account API ───────────────────────────────────────────────────────────

func (c *Client) GetUserInfo() (*UserInfo, error) {
	raw, err := c.authGET("/v1/user/me")
	if err != nil {
		return nil, err
	}
	var u UserInfo
	json.Unmarshal(raw, &u)
	return &u, nil
}

func (c *Client) GetQuota() (*QuotaInfo, error) {
	raw, err := c.driveGET("/drive/v1/about", nil)
	if err != nil {
		return nil, err
	}
	var q QuotaInfo
	json.Unmarshal(raw, &q)
	return &q, nil
}

func (c *Client) GetVipInfo() (*VipInfoResponse, error) {
	raw, err := c.driveGET("/drive/v1/privilege/vip", nil)
	if err != nil {
		return nil, err
	}
	var v VipInfoResponse
	json.Unmarshal(raw, &v)
	return &v, nil
}

func (c *Client) GetTransferQuota() (*TransferQuotaResponse, error) {
	raw, err := c.driveGET("/drive/v1/privilege/transfer", nil)
	if err != nil {
		return nil, err
	}
	var t TransferQuotaResponse
	json.Unmarshal(raw, &t)
	return &t, nil
}

// ── File API ──────────────────────────────────────────────────────────────

func (c *Client) ListFiles(parentID string, limit int) (*DriveListResponse, error) {
	q := map[string]string{
		"parent_id":      parentID,
		"limit":          fmt.Sprintf("%d", limit),
		"thumbnail_size": "SIZE_MEDIUM",
	}
	if parentID != "" {
		q["filters"] = `{"trashed":{"eq":false}}`
	}
	raw, err := c.driveGET("/drive/v1/files", q)
	if err != nil {
		return nil, err
	}
	var list DriveListResponse
	json.Unmarshal(raw, &list)
	return &list, nil
}

func (c *Client) GetFileInfo(fileID string) (*FileInfoResponse, error) {
	raw, err := c.driveGET("/drive/v1/files/" + fileID, nil)
	if err != nil {
		return nil, err
	}
	var info FileInfoResponse
	json.Unmarshal(raw, &info)
	return &info, nil
}

func (c *Client) Mkdir(parentID, name string) (*DriveFile, error) {
	body := map[string]string{
		"kind":      "drive#folder",
		"parent_id": parentID,
		"name":      name,
	}
	raw, err := c.drivePOST("/drive/v1/files", body)
	if err != nil {
		return nil, err
	}
	var resp CreateFolderResponse
	json.Unmarshal(raw, &resp)
	return &resp.File, nil
}

func (c *Client) Rename(fileID, newName string) error {
	body := map[string]string{"name": newName}
	_, err := c.drivePATCH("/drive/v1/files/"+fileID, body)
	return err
}

func (c *Client) MoveFiles(ids []string, toParentID string) error {
	body := map[string]interface{}{
		"ids": ids,
		"to":  map[string]string{"parent_id": toParentID},
	}
	_, err := c.drivePOST("/drive/v1/files:batchMove", body)
	return err
}

func (c *Client) CopyFiles(ids []string, toParentID string) error {
	body := map[string]interface{}{
		"ids": ids,
		"to":  map[string]string{"parent_id": toParentID},
	}
	_, err := c.drivePOST("/drive/v1/files:batchCopy", body)
	return err
}

func (c *Client) TrashFiles(ids []string) error {
	body := map[string]interface{}{"ids": ids}
	_, err := c.drivePOST("/drive/v1/files:batchTrash", body)
	return err
}

func (c *Client) DeleteFiles(ids []string) error {
	body := map[string]interface{}{"ids": ids}
	_, err := c.drivePOST("/drive/v1/files:batchDelete", body)
	return err
}

func (c *Client) UntrashFiles(ids []string) error {
	body := map[string]interface{}{"ids": ids}
	_, err := c.drivePOST("/drive/v1/files:batchUntrash", body)
	return err
}

func (c *Client) EmptyTrash() error {
	_, err := c.drivePATCH("/drive/v1/files/trash:empty", map[string]interface{}{})
	return err
}

func (c *Client) ListTrash(limit int) (*TrashListResponse, error) {
	q := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"filters": `{"trashed":{"eq":true}}`,
	}
	raw, err := c.driveGET("/drive/v1/files", q)
	if err != nil {
		return nil, err
	}
	var list TrashListResponse
	json.Unmarshal(raw, &list)
	return &list, nil
}

// ── Star API ──────────────────────────────────────────────────────────────

func (c *Client) StarFiles(ids []string) error {
	body := map[string]interface{}{"ids": ids}
	_, err := c.drivePOST("/drive/v1/files:star", body)
	return err
}

func (c *Client) UnstarFiles(ids []string) error {
	body := map[string]interface{}{"ids": ids}
	_, err := c.drivePOST("/drive/v1/files:unstar", body)
	return err
}

func (c *Client) ListStarred(limit int) (*DriveListResponse, error) {
	q := map[string]string{
		"parent_id":      "*",
		"limit":          fmt.Sprintf("%d", limit),
		"filters":        `{"trashed":{"eq":false},"system_tag":{"in":"STAR"}}`,
		"thumbnail_size": "SIZE_MEDIUM",
	}
	raw, err := c.driveGET("/drive/v1/files", q)
	if err != nil {
		return nil, err
	}
	var list DriveListResponse
	json.Unmarshal(raw, &list)
	return &list, nil
}

// ── Trash API ─────────────────────────────────────────────────────────────

func (c *Client) GetDownloadLink(fileID string) (*DownloadResponse, error) {
	raw, err := c.driveGET("/drive/v1/files/"+fileID, nil)
	if err != nil {
		return nil, err
	}
	// Parse minimally — response has web_content_link
	var resp DownloadResponse
	json.Unmarshal(raw, &resp)
	return &resp, nil
}

// ── Offline Download API ──────────────────────────────────────────────────

func (c *Client) AddOfflineTask(magnetURL, parentID, name string) (*OfflineTaskResponse, error) {
	payload := map[string]interface{}{
		"kind":        "drive#file",
		"upload_type": "UPLOAD_TYPE_URL",
		"url":         map[string]string{"url": magnetURL},
	}
	if parentID != "" {
		payload["parent_id"] = parentID
		payload["folder_type"] = ""
	} else {
		payload["folder_type"] = "DOWNLOAD"
	}
	if name != "" {
		payload["name"] = name
	}
	raw, err := c.drivePOST("/drive/v1/files", payload)
	if err != nil {
		return nil, err
	}
	var resp OfflineTaskResponse
	json.Unmarshal(raw, &resp)
	return &resp, nil
}

func (c *Client) ListOfflineTasks(limit int, phases []string) (*OfflineListResponse, error) {
	q := map[string]string{
		"type": "offline", "thumbnail_size": "SIZE_SMALL",
		"limit": fmt.Sprintf("%d", limit),
	}
	if len(phases) > 0 {
		f, _ := json.Marshal(map[string]interface{}{
			"phase": map[string]string{"in": strings.Join(phases, ",")},
		})
		q["filters"] = string(f)
	}
	raw, err := c.driveGET("/drive/v1/tasks", q)
	if err != nil {
		return nil, err
	}
	var list OfflineListResponse
	json.Unmarshal(raw, &list)
	return &list, nil
}

func (c *Client) GetOfflineTask(taskID string) (*OfflineTask, error) {
	raw, err := c.driveGET("/drive/v1/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}
	// Response wraps task in {"task":{...}} or returns directly
	var resp struct {
		Task OfflineTask `json:"task"`
	}
	json.Unmarshal(raw, &resp)
	if resp.Task.ID == "" {
		var task OfflineTask
		json.Unmarshal(raw, &task)
		resp.Task = task
	}
	return &resp.Task, nil
}

func (c *Client) DeleteOfflineTask(taskID string, deleteFile bool) error {
	q := map[string]string{
		"task_ids":     taskID,
		"delete_files": fmt.Sprintf("%t", deleteFile),
	}
	_, err := c.driveDELETE("/drive/v1/tasks", q)
	return err
}

func (c *Client) RetryOfflineTask(taskID string) error {
	payload := map[string]interface{}{
		"type":        "offline",
		"create_type": "RETRY",
		"id":          taskID,
	}
	_, err := c.drivePOST("/drive/v1/task", payload)
	return err
}

// ── Events API ────────────────────────────────────────────────────────────

func (c *Client) ListEvents(limit int) (*EventsResponse, error) {
	q := map[string]string{
		"limit": fmt.Sprintf("%d", limit),
	}
	raw, err := c.driveGET("/drive/v1/events", q)
	if err != nil {
		return nil, err
	}
	var resp EventsResponse
	json.Unmarshal(raw, &resp)
	return &resp, nil
}

// ── Share API ─────────────────────────────────────────────────────────────

func (c *Client) GetShareInfo(shareID, passCode string) (*ShareInfoResponse, error) {
	q := map[string]string{"share_id": shareID}
	if passCode != "" {
		q["pass_code"] = passCode
	}
	raw, err := c.driveGET("/drive/v1/share", q)
	if err != nil {
		return nil, err
	}
	var resp ShareInfoResponse
	json.Unmarshal(raw, &resp)
	return &resp, nil
}

func (c *Client) SaveShare(shareID, passCodeToken string, fileIDs []string, toParentID string) error {
	body := map[string]interface{}{
		"share_id":        shareID,
		"pass_code_token": passCodeToken,
		"to_parent_id":    toParentID,
	}
	if len(fileIDs) > 0 {
		body["file_ids"] = fileIDs
	}
	_, err := c.drivePOST("/drive/v1/share/restore", body)
	return err
}

func (c *Client) CreateShare(fileIDs []string, expireDays int, passCode string) (*CreateShareResponse, error) {
	shareTo := "publiclink"
	passOption := "NOT_REQUIRED"
	if passCode != "" {
		shareTo = "encryptedlink"
		passOption = "REQUIRED"
	}
	body := map[string]interface{}{
		"file_ids":          fileIDs,
		"share_to":          shareTo,
		"expiration_days":   expireDays,
		"pass_code_option":  passOption,
	}
	if passCode != "" {
		body["pass_code"] = passCode
	}
	raw, err := c.drivePOST("/drive/v1/share", body)
	if err != nil {
		return nil, err
	}
	var resp CreateShareResponse
	json.Unmarshal(raw, &resp)
	return &resp, nil
}

func (c *Client) ShareDetail(shareID, passCodeToken, dirID string) (*ShareDetailResponse, error) {
	q := map[string]string{
		"share_id":  shareID,
		"parent_id": dirID,
		"pass_code": passCodeToken,
		"limit":     "100",
	}
	raw, err := c.driveGET("/drive/v1/share/detail", q)
	if err != nil {
		return nil, err
	}
	var resp ShareDetailResponse
	json.Unmarshal(raw, &resp)
	return &resp, nil
}

func (c *Client) ListShares() (*ShareListResponse, error) {
	raw, err := c.driveGET("/drive/v1/share/list", nil)
	if err != nil {
		return nil, err
	}
	var resp ShareListResponse
	json.Unmarshal(raw, &resp)
	return &resp, nil
}

func (c *Client) DeleteShares(shareIDs []string) error {
	body := map[string]interface{}{"ids": shareIDs}
	_, err := c.drivePOST("/drive/v1/share:batchDelete", body)
	return err
}

// ── Region Error ──────────────────────────────────────────────────────────

type RegionError struct{ Err error }

func (e *RegionError) Error() string { return fmt.Sprintf("region restricted: %v", e.Err) }
func (e *RegionError) Unwrap() error { return e.Err }

func isRegionRestricted(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	kws := []string{
		"region", "territory", "not supported", "location",
		"forbidden", "blocked", "405", "403", "451",
		"connection reset by peer", "connection refused",
		"reset by peer", "i/o timeout", "no route to host",
		"network is unreachable",
	}
	for _, kw := range kws {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}
