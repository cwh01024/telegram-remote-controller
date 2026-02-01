# Ticket 003: Auth

## Context
安全驗證層，只允許特定 Telegram 用戶使用。這個程式有電腦最高權限，必須限制存取。

## Goal
- 用戶 ID 白名單驗證
- 拒絕未授權請求
- 記錄所有嘗試

## Files
- `internal/auth/whitelist.go`
- `internal/auth/whitelist_test.go`

## Key Interface
```go
type Authenticator interface {
    IsAuthorized(userID int64) bool
    GetAllowedUsers() []int64
}
```

## Config
```yaml
allowed_users:
  - 123456789  # Your Telegram ID
```

## Acceptance Criteria
- [x] 白名單內用戶通過驗證
- [x] 非白名單用戶被拒絕
- [x] 所有請求被記錄

## TDD
```go
func TestWhitelistAuth(t *testing.T) {
    auth := NewWhitelist([]int64{123})
    assert.True(t, auth.IsAuthorized(123))
    assert.False(t, auth.IsAuthorized(999))
}
```
