package model

import "time"

type AdminIdentity struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	SessionID int64  `json:"session_id,omitempty"`
}

type AdminProfile struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
	DeviceID   string `json:"device_id"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	AccessExpiry time.Time `json:"access_expires_at"`
}

type LoginResponse struct {
	Admin       AdminProfile `json:"admin"`
	AccessToken string       `json:"access_token"`
	ExpiresIn   int          `json:"expires_in"` // seconds
}
