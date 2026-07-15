package pikpak

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ── Constants ─────────────────────────────────────────────────────────────

const (
	DefaultAuthBaseURL   = "https://user.mypikpak.com"
	DefaultDriveBaseURL  = "https://api-drive.mypikpak.com"
	DefaultClientID      = "YNxT9w7GMdWvEOKa"
	DefaultClientSecret  = "dbw2OtmVEeuUvIptb1Coyg"
	DefaultUserAgent     = "ANDROID-com.pikcloud.pikpak/1.53.2"
	DefaultConfigDir     = "~/.config/pikpakdriver"
	DefaultConfigFile    = "config.toml"
	DefaultSessionFile   = "session.json"
)

// ── Config ────────────────────────────────────────────────────────────────

type Config struct {
	Email        string `toml:"email,omitempty" json:"email,omitempty"`
	Password     string `toml:"password,omitempty" json:"-"`
	AccessToken  string `toml:"access_token,omitempty" json:"access_token,omitempty"`
	RefreshToken string `toml:"refresh_token,omitempty" json:"refresh_token,omitempty"`
}

func ConfigPath() string  { return filepath.Join(expandPath(DefaultConfigDir), DefaultConfigFile) }
func SessionPath() string { return filepath.Join(expandPath(DefaultConfigDir), DefaultSessionFile) }
func ConfigDir() string   { return expandPath(DefaultConfigDir) }

// ── Session Token ──────────────────────────────────────────────────────────

type SessionToken struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	ExpiresAtUnix int64  `json:"expires_at_unix"`
	DeviceID      string `json:"device_id,omitempty"`
	CaptchaToken  string `json:"captcha_token,omitempty"`
	UserID        string `json:"user_id,omitempty"`
}

func (s *SessionToken) IsExpired() bool { return time.Now().Unix() >= s.ExpiresAtUnix }

// ── Auth Models ────────────────────────────────────────────────────────────

type TokenRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Sub          string `json:"sub,omitempty"`
}

type SigninResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Sub          string `json:"sub"`
}

type CaptchaInitResponse struct {
	CaptchaToken string `json:"captcha_token,omitempty"`
	URL          string `json:"url,omitempty"`
}

// ── Account Models ─────────────────────────────────────────────────────────

type UserInfo struct {
	Sub         string `json:"sub"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type QuotaInfo struct {
	Kind     string      `json:"kind"`
	Quota    QuotaDetail `json:"quota"`
	Messages []string    `json:"messages"`
}

type QuotaDetail struct {
	Limit        string `json:"limit"`
	Usage        string `json:"usage"`
	Remain       string `json:"remain"`
	UsageInTrash string `json:"usage_in_trash,omitempty"`
}

type TransferQuotaResponse struct {
	Base TransferQuotaBase `json:"base"`
}

type TransferQuotaBase struct {
	Offline      *TransferBand `json:"offline,omitempty"`
	Download     *TransferBand `json:"download,omitempty"`
	Upload       *TransferBand `json:"upload,omitempty"`
	DownloadDaily *TransferBand `json:"download_daily,omitempty"`
	ExpireTime   string         `json:"expire_time,omitempty"`
}

type TransferBand struct {
	TotalAssets int64 `json:"total_assets"`
	Assets      int64 `json:"assets"`
}

type VipInfoResponse struct {
	Data *VipData `json:"data,omitempty"`
}

type VipData struct {
	VipType string `json:"type,omitempty"`
	Status  string `json:"status,omitempty"`
	Expire  string `json:"expire,omitempty"`
}

// ── File Models ────────────────────────────────────────────────────────────

type DriveFile struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Kind          string `json:"kind"`
	Size          string `json:"size,omitempty"`
	CreatedTime   string `json:"created_time,omitempty"`
	ModifiedTime  string `json:"modified_time,omitempty"`
	ThumbnailLink string `json:"thumbnail_link,omitempty"`
	ParentID      string `json:"parent_id,omitempty"`
	MimeType      string `json:"mime_type,omitempty"`
	FileExtension string `json:"file_extension,omitempty"`
}

func (f *DriveFile) IsFolder() bool { return strings.Contains(f.Kind, "folder") }

type DriveListResponse struct {
	Files         []DriveFile `json:"files"`
	NextPageToken string      `json:"next_page_token,omitempty"`
}

type FileInfoResponse struct {
	File DriveFile `json:"file"`
}

type BatchActionRequest struct {
	IDs       []string `json:"ids"`
	ToParentID string  `json:"to.parent_id,omitempty"`
}

type BatchActionResponse struct {
	TaskID string `json:"task_id,omitempty"`
}

type CreateFolderResponse struct {
	File DriveFile `json:"file"`
}

// ── Download Models ────────────────────────────────────────────────────────

type DownloadInfo struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type DownloadResponse struct {
	WebContentLink string `json:"web_content_link"`
	Size           string `json:"size"`
}

// ── Offline Download Models ────────────────────────────────────────────────

type OfflineTask struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	UserID      string `json:"user_id"`
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	FileSize    string `json:"file_size"`
	Message     string `json:"message"`
	Phase       string `json:"phase"`
	Progress    int    `json:"progress"`
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	ThirdTaskID string `json:"third_task_id"`
}

type OfflineTaskResponse struct {
	UploadType string      `json:"upload_type"`
	URL        interface{} `json:"url,omitempty"`
	Task       OfflineTask `json:"task"`
}

type OfflineListResponse struct {
	Tasks         []OfflineTask `json:"tasks"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

// ── Share Models ───────────────────────────────────────────────────────────

type CreateShareRequest struct {
	FileIDs     []string `json:"file_ids"`
	ExpireDays  int      `json:"expire_days,omitempty"`
	ShareTo     string   `json:"share_to,omitempty"`
	PassCode    string   `json:"pass_code,omitempty"`
}

type CreateShareResponse struct {
	ShareID   string `json:"share_id"`
	ShareURL  string `json:"share_url"`
	PassCode  string `json:"pass_code"`
	ShareText string `json:"share_text"`
}

type ShareInfoResponse struct {
	ShareStatus   string       `json:"share_status"`
	PassCodeToken string       `json:"pass_code_token"`
	Files         []ShareEntry `json:"files"`
}

type ShareEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	Size string `json:"size,omitempty"`
}

type ShareDetailResponse struct {
	Files         []ShareEntry `json:"files"`
	NextPageToken string       `json:"next_page_token,omitempty"`
}

type MyShare struct {
	ShareID       string `json:"share_id"`
	ShareURL      string `json:"share_url"`
	Title         string `json:"title"`
	PassCode      string `json:"pass_code"`
	CreateTime    string `json:"create_time"`
	ExpirationDays string `json:"expiration_days"`
	ViewCount     string `json:"view_count"`
	FileNum       string `json:"file_num"`
	ShareStatus   string `json:"share_status"`
}

type ShareListResponse struct {
	Data          []MyShare `json:"data"`
	NextPageToken string    `json:"next_page_token,omitempty"`
}

// ── Events Models ──────────────────────────────────────────────────────────

type EventsResponse struct {
	Events []EventEntry `json:"events"`
}

type EventEntry struct {
	Type             string           `json:"type"`
	TypeName         string           `json:"type_name"`
	FileName         string           `json:"file_name"`
	CreatedTime      string           `json:"created_time"`
	ReferenceResource *EventResource  `json:"reference_resource,omitempty"`
}

type EventResource struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
}

// ── Trash Models ───────────────────────────────────────────────────────────

type TrashListResponse struct {
	Files         []DriveFile `json:"files"`
	NextPageToken string      `json:"next_page_token,omitempty"`
}

// ── Credentials ────────────────────────────────────────────────────────────

const ConfigFile = "config.json"

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Proxy    string `json:"proxy,omitempty"`
}

func CredentialsPath() string { return filepath.Join(expandPath(DefaultConfigDir), ConfigFile) }

func LoadCredentials() (*Credentials, error) {
	path := CredentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read credentials %s: %w", path, err)
	}
	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	if creds.Email == "" || creds.Password == "" {
		return nil, fmt.Errorf("credentials file %s is incomplete", path)
	}
	return &creds, nil
}

func SaveCredentials(email, password, proxy string) error {
	if err := EnsureDir(); err != nil {
		return err
	}
	path := CredentialsPath()
	creds := Credentials{Email: email, Password: password, Proxy: proxy}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// ── Utility ────────────────────────────────────────────────────────────────

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}

func EnsureDir() error {
	return os.MkdirAll(ConfigDir(), 0700)
}

func LoadSession() (*SessionToken, error) {
	path := SessionPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session %s: %w", path, err)
	}
	var token SessionToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	return &token, nil
}

func SaveSession(token *SessionToken) error {
	if err := EnsureDir(); err != nil {
		return err
	}
	path := SessionPath()
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
