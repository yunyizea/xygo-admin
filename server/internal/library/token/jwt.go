package token

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/golang-jwt/jwt/v5"

	"xygo/internal/consts"
	"xygo/internal/library/cache"
	"xygo/internal/model"
)

const (
	AppAdmin  = "admin"
	AppMember = "member"
)

// Config JWT 配置
type Config struct {
	Secret         string
	Expires        int64 // accessToken 有效期（秒）
	RefreshExpires int64 // refreshToken 最长有效期（秒）
	MultiLogin     bool  // 是否允许多端登录（false=单点登录，新登录踢掉旧会话）
}

// Claims JWT 载荷（包含完整用户信息）
type Claims struct {
	*model.AuthUser
	jwt.RegisteredClaims
}

// MemberClaims 会员 JWT 载荷
type MemberClaims struct {
	*model.MemberUser
	jwt.RegisteredClaims
}

func getConfig(ctx context.Context) Config {
	secret := g.Cfg().MustGet(ctx, "auth.jwt.secret").String()
	if secret == "" {
		secret = "xygo-secret-key"
	}
	expires := g.Cfg().MustGet(ctx, "auth.jwt.expires").Int64()
	if expires <= 0 {
		expires = 7200
	}
	refreshExpires := g.Cfg().MustGet(ctx, "auth.jwt.refreshExpires").Int64()
	if refreshExpires <= 0 {
		refreshExpires = 172800
	}
	multiLogin := g.Cfg().MustGet(ctx, "auth.jwt.multiLogin").Bool()
	return Config{
		Secret:         secret,
		Expires:        expires,
		RefreshExpires: refreshExpires,
		MultiLogin:     multiLogin,
	}
}

// TokenMeta accessToken 元数据（存储在缓存中）
type TokenMeta struct {
	ExpireAt int64 `json:"exp"`
	Kicked   bool  `json:"kicked"`
}

// RefreshMeta refreshToken 元数据（存储在缓存中）
type RefreshMeta struct {
	UserId   uint64 `json:"uid"`
	TokenKey string `json:"tk"`
	ExpireAt int64  `json:"exp"`
}

// ErrTokenKicked 自定义错误：Token 被踢下线
var ErrTokenKicked = fmt.Errorf("token kicked")

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ==================== Cache Key ====================

func GetAuthKey(token string) string {
	return gmd5.MustEncryptString("xygo" + token)
}

func GetTokenKey(appName, authKey string) string {
	return fmt.Sprintf("%s:token:%s:%s", consts.CachePrefix, appName, authKey)
}

func GetBindKey(appName string, userId uint64) string {
	return fmt.Sprintf("%s:token:bind:%s:%d", consts.CachePrefix, appName, userId)
}

func GetRefreshKey(appName, refreshToken string) string {
	return fmt.Sprintf("%s:refresh:%s:%s", consts.CachePrefix, appName, refreshToken)
}

func GetRefreshBindKey(appName string, userId uint64) string {
	return fmt.Sprintf("%s:refresh:bind:%s:%d", consts.CachePrefix, appName, userId)
}

// ==================== Generate ====================

// Generate 登录成功后生成双 Token（后台管理员专用，保持兼容）
func Generate(ctx context.Context, user model.AuthUser) (accessToken, refreshToken string, expiresIn, refreshExpiresIn int64, err error) {
	return GenerateAdmin(ctx, user)
}

// GenerateAdmin 后台管理员登录生成 accessToken + refreshToken
func GenerateAdmin(ctx context.Context, user model.AuthUser) (accessToken, refreshToken string, expiresIn, refreshExpiresIn int64, err error) {
	cfg := getConfig(ctx)
	now := time.Now()
	expireAt := now.Add(time.Duration(cfg.Expires) * time.Second)

	claims := &Claims{
		AuthUser: &user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = tok.SignedString([]byte(cfg.Secret))
	if err != nil {
		return
	}

	refreshToken, err = generateRefreshToken()
	if err != nil {
		return
	}

	var (
		authKey         = GetAuthKey(accessToken)
		tokenKey        = GetTokenKey(AppAdmin, authKey)
		bindKey         = GetBindKey(AppAdmin, user.Id)
		refreshKey      = GetRefreshKey(AppAdmin, refreshToken)
		refreshBindKey  = GetRefreshBindKey(AppAdmin, user.Id)
		duration        = time.Second * gconv.Duration(cfg.Expires)
		refreshDuration = time.Second * gconv.Duration(cfg.RefreshExpires)
	)

	if !cfg.MultiLogin {
		kickOldSession(ctx, AppAdmin, bindKey, refreshBindKey)
	}

	tokenMeta := &TokenMeta{ExpireAt: expireAt.Unix()}
	if err = cache.Instance().Set(ctx, tokenKey, tokenMeta, duration); err != nil {
		return
	}
	if err = cache.Instance().Set(ctx, bindKey, tokenKey, duration); err != nil {
		return
	}

	refreshMeta := &RefreshMeta{
		UserId:   user.Id,
		TokenKey: tokenKey,
		ExpireAt: now.Add(time.Duration(cfg.RefreshExpires) * time.Second).Unix(),
	}
	if err = cache.Instance().Set(ctx, refreshKey, refreshMeta, refreshDuration); err != nil {
		return
	}
	if err = cache.Instance().Set(ctx, refreshBindKey, refreshKey, refreshDuration); err != nil {
		return
	}

	expiresIn = cfg.Expires
	refreshExpiresIn = cfg.RefreshExpires
	return
}

// GenerateMember 前台会员登录生成 accessToken + refreshToken
func GenerateMember(ctx context.Context, user model.MemberUser) (accessToken, refreshToken string, expiresIn, refreshExpiresIn int64, err error) {
	cfg := getConfig(ctx)
	now := time.Now()
	expireAt := now.Add(time.Duration(cfg.Expires) * time.Second)

	claims := &MemberClaims{
		MemberUser: &user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = tok.SignedString([]byte(cfg.Secret))
	if err != nil {
		return
	}

	refreshToken, err = generateRefreshToken()
	if err != nil {
		return
	}

	var (
		authKey         = GetAuthKey(accessToken)
		tokenKey        = GetTokenKey(AppMember, authKey)
		bindKey         = GetBindKey(AppMember, user.Id)
		refreshKey      = GetRefreshKey(AppMember, refreshToken)
		refreshBindKey  = GetRefreshBindKey(AppMember, user.Id)
		duration        = time.Second * gconv.Duration(cfg.Expires)
		refreshDuration = time.Second * gconv.Duration(cfg.RefreshExpires)
	)

	if !cfg.MultiLogin {
		kickOldSession(ctx, AppMember, bindKey, refreshBindKey)
	}

	tokenMeta := &TokenMeta{ExpireAt: expireAt.Unix()}
	if err = cache.Instance().Set(ctx, tokenKey, tokenMeta, duration); err != nil {
		return
	}
	if err = cache.Instance().Set(ctx, bindKey, tokenKey, duration); err != nil {
		return
	}

	refreshMeta := &RefreshMeta{
		UserId:   user.Id,
		TokenKey: tokenKey,
		ExpireAt: now.Add(time.Duration(cfg.RefreshExpires) * time.Second).Unix(),
	}
	if err = cache.Instance().Set(ctx, refreshKey, refreshMeta, refreshDuration); err != nil {
		return
	}
	if err = cache.Instance().Set(ctx, refreshBindKey, refreshKey, refreshDuration); err != nil {
		return
	}

	expiresIn = cfg.Expires
	refreshExpiresIn = cfg.RefreshExpires
	return
}

// ==================== Parse ====================

// Parse 解析并校验 accessToken（后台管理员专用，保持兼容）
func Parse(ctx context.Context, accessToken string) (*model.AuthUser, error) {
	return ParseAdmin(ctx, accessToken)
}

// ParseAdmin 解析后台管理员 accessToken
func ParseAdmin(ctx context.Context, accessToken string) (*model.AuthUser, error) {
	cfg := getConfig(ctx)
	parsed, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, err
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims.AuthUser == nil {
		return nil, jwt.ErrTokenMalformed
	}

	authKey := GetAuthKey(accessToken)
	tokenKey := GetTokenKey(AppAdmin, authKey)
	if err = validateTokenMeta(ctx, tokenKey); err != nil {
		return nil, err
	}
	return claims.AuthUser, nil
}

// ParseMember 解析前台会员 accessToken
func ParseMember(ctx context.Context, accessToken string) (*model.MemberUser, error) {
	cfg := getConfig(ctx)
	parsed, err := jwt.ParseWithClaims(accessToken, &MemberClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, err
	}

	claims, ok := parsed.Claims.(*MemberClaims)
	if !ok || claims.MemberUser == nil {
		return nil, jwt.ErrTokenMalformed
	}

	authKey := GetAuthKey(accessToken)
	tokenKey := GetTokenKey(AppMember, authKey)
	if err = validateTokenMeta(ctx, tokenKey); err != nil {
		return nil, err
	}
	return claims.MemberUser, nil
}

func validateTokenMeta(ctx context.Context, tokenKey string) error {
	tk, err := cache.Instance().Get(ctx, tokenKey)
	if err != nil {
		g.Log().Debugf(ctx, "get tokenKey err:%+v", err)
		return jwt.ErrTokenInvalidAudience
	}
	if tk.IsEmpty() {
		g.Log().Debug(ctx, "token isEmpty")
		return jwt.ErrTokenInvalidAudience
	}

	var tokenMeta *TokenMeta
	if err = tk.Scan(&tokenMeta); err != nil {
		g.Log().Debugf(ctx, "token scan err:%+v", err)
		return jwt.ErrTokenInvalidAudience
	}
	if tokenMeta == nil {
		return jwt.ErrTokenInvalidAudience
	}
	if tokenMeta.Kicked {
		return ErrTokenKicked
	}
	if tokenMeta.ExpireAt < time.Now().Unix() {
		return jwt.ErrTokenExpired
	}
	return nil
}

// ==================== Refresh ====================

// ValidateRefreshToken 验证 refreshToken 是否有效，返回关联的 userId
func ValidateRefreshToken(ctx context.Context, appName string, refreshTokenStr string) (userId uint64, err error) {
	refreshKey := GetRefreshKey(appName, refreshTokenStr)
	val, err := cache.Instance().Get(ctx, refreshKey)
	if err != nil || val.IsEmpty() {
		return 0, jwt.ErrTokenExpired
	}
	var meta *RefreshMeta
	if err = val.Scan(&meta); err != nil || meta == nil {
		return 0, jwt.ErrTokenMalformed
	}
	if meta.ExpireAt < time.Now().Unix() {
		return 0, jwt.ErrTokenExpired
	}
	return meta.UserId, nil
}

// RefreshAccessAdmin 用 refreshToken 签发新的 accessToken（后台管理员）
func RefreshAccessAdmin(ctx context.Context, refreshTokenStr string, user model.AuthUser) (accessToken string, expiresIn int64, err error) {
	cfg := getConfig(ctx)
	refreshKey := GetRefreshKey(AppAdmin, refreshTokenStr)

	val, err := cache.Instance().Get(ctx, refreshKey)
	if err != nil || val.IsEmpty() {
		return "", 0, jwt.ErrTokenExpired
	}
	var refreshMeta *RefreshMeta
	if err = val.Scan(&refreshMeta); err != nil || refreshMeta == nil {
		return "", 0, jwt.ErrTokenMalformed
	}
	if refreshMeta.ExpireAt < time.Now().Unix() {
		return "", 0, jwt.ErrTokenExpired
	}

	// 删除旧 accessToken 缓存
	if refreshMeta.TokenKey != "" {
		_, _ = cache.Instance().Remove(ctx, refreshMeta.TokenKey)
	}

	// 签发新 accessToken
	now := time.Now()
	expireAt := now.Add(time.Duration(cfg.Expires) * time.Second)
	claims := &Claims{
		AuthUser: &user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = tok.SignedString([]byte(cfg.Secret))
	if err != nil {
		return
	}

	authKey := GetAuthKey(accessToken)
	tokenKey := GetTokenKey(AppAdmin, authKey)
	bindKey := GetBindKey(AppAdmin, user.Id)
	duration := time.Second * gconv.Duration(cfg.Expires)

	tokenMeta := &TokenMeta{ExpireAt: expireAt.Unix()}
	if err = cache.Instance().Set(ctx, tokenKey, tokenMeta, duration); err != nil {
		return
	}
	if err = cache.Instance().Set(ctx, bindKey, tokenKey, duration); err != nil {
		return
	}

	// 更新 RefreshMeta（保持原始过期时间不变）
	remaining := time.Duration(refreshMeta.ExpireAt-now.Unix()) * time.Second
	if remaining <= 0 {
		remaining = time.Second
	}
	refreshMeta.TokenKey = tokenKey
	_ = cache.Instance().Set(ctx, refreshKey, refreshMeta, remaining)

	expiresIn = cfg.Expires
	return
}

// RefreshAccessMember 用 refreshToken 签发新的 accessToken（前台会员）
func RefreshAccessMember(ctx context.Context, refreshTokenStr string, user model.MemberUser) (accessToken string, expiresIn int64, err error) {
	cfg := getConfig(ctx)
	refreshKey := GetRefreshKey(AppMember, refreshTokenStr)

	val, err := cache.Instance().Get(ctx, refreshKey)
	if err != nil || val.IsEmpty() {
		return "", 0, jwt.ErrTokenExpired
	}
	var refreshMeta *RefreshMeta
	if err = val.Scan(&refreshMeta); err != nil || refreshMeta == nil {
		return "", 0, jwt.ErrTokenMalformed
	}
	if refreshMeta.ExpireAt < time.Now().Unix() {
		return "", 0, jwt.ErrTokenExpired
	}

	if refreshMeta.TokenKey != "" {
		_, _ = cache.Instance().Remove(ctx, refreshMeta.TokenKey)
	}

	now := time.Now()
	expireAt := now.Add(time.Duration(cfg.Expires) * time.Second)
	claims := &MemberClaims{
		MemberUser: &user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = tok.SignedString([]byte(cfg.Secret))
	if err != nil {
		return
	}

	authKey := GetAuthKey(accessToken)
	tokenKey := GetTokenKey(AppMember, authKey)
	bindKey := GetBindKey(AppMember, user.Id)
	duration := time.Second * gconv.Duration(cfg.Expires)

	tokenMeta := &TokenMeta{ExpireAt: expireAt.Unix()}
	if err = cache.Instance().Set(ctx, tokenKey, tokenMeta, duration); err != nil {
		return
	}
	if err = cache.Instance().Set(ctx, bindKey, tokenKey, duration); err != nil {
		return
	}

	remaining := time.Duration(refreshMeta.ExpireAt-now.Unix()) * time.Second
	if remaining <= 0 {
		remaining = time.Second
	}
	refreshMeta.TokenKey = tokenKey
	_ = cache.Instance().Set(ctx, refreshKey, refreshMeta, remaining)

	expiresIn = cfg.Expires
	return
}

// ==================== Delete ====================

// Delete 删除 accessToken（退出登录，后台管理员专用，保持兼容）
func Delete(ctx context.Context, accessToken string) error {
	return DeleteByApp(ctx, AppAdmin, accessToken)
}

// DeleteByApp 按应用类型删除 accessToken
func DeleteByApp(ctx context.Context, appName string, accessToken string) error {
	authKey := GetAuthKey(accessToken)
	tokenKey := GetTokenKey(appName, authKey)
	_, err := cache.Instance().Remove(ctx, tokenKey)
	return err
}

// DeleteSession 删除用户的完整会话（accessToken + refreshToken），退出登录时使用
func DeleteSession(ctx context.Context, appName string, accessToken string, userId uint64) error {
	_ = DeleteByApp(ctx, appName, accessToken)

	refreshBindKey := GetRefreshBindKey(appName, userId)
	oldRefreshKey, err := cache.Instance().Get(ctx, refreshBindKey)
	if err == nil && !oldRefreshKey.IsEmpty() {
		_, _ = cache.Instance().Remove(ctx, oldRefreshKey.String())
	}
	_, _ = cache.Instance().Remove(ctx, refreshBindKey)

	bindKey := GetBindKey(appName, userId)
	_, _ = cache.Instance().Remove(ctx, bindKey)
	return nil
}

// ==================== SSO 踢人 ====================

// kickOldSession 踢掉旧会话（SSO单点登录核心逻辑）
// 标记旧 accessToken 为 kicked，同时删除旧 refreshToken
func kickOldSession(ctx context.Context, appName string, bindKey string, refreshBindKey string) {
	// 踢旧 accessToken
	oldTokenKey, err := cache.Instance().Get(ctx, bindKey)
	if err == nil && !oldTokenKey.IsEmpty() {
		oldKey := oldTokenKey.String()
		oldMeta, getErr := cache.Instance().Get(ctx, oldKey)
		if getErr == nil && !oldMeta.IsEmpty() {
			var meta *TokenMeta
			if scanErr := oldMeta.Scan(&meta); scanErr == nil && meta != nil {
				meta.Kicked = true
				_ = cache.Instance().Set(ctx, oldKey, meta, 30*time.Second)
			}
		}
	}

	// 删除旧 refreshToken（使被踢设备无法通过刷新绕过 SSO）
	if refreshBindKey != "" {
		oldRefreshKey, getErr := cache.Instance().Get(ctx, refreshBindKey)
		if getErr == nil && !oldRefreshKey.IsEmpty() {
			_, _ = cache.Instance().Remove(ctx, oldRefreshKey.String())
			_, _ = cache.Instance().Remove(ctx, refreshBindKey)
		}
	}
}

// KickByUserId 按用户ID踢人（管理员强制下线）
func KickByUserId(ctx context.Context, appName string, userId uint64) error {
	bindKey := GetBindKey(appName, userId)
	refreshBindKey := GetRefreshBindKey(appName, userId)
	kickOldSession(ctx, appName, bindKey, refreshBindKey)
	_, _ = cache.Instance().Remove(ctx, bindKey)

	notifyKickViaWs(appName, userId)
	return nil
}

// ==================== WebSocket 踢人通知 ====================

var WsKickNotifier func(userType string, userId uint64)

func notifyKickViaWs(appName string, userId uint64) {
	if WsKickNotifier != nil {
		userType := "admin"
		if appName != "admin" {
			userType = "member"
		}
		WsKickNotifier(userType, userId)
	}
}
