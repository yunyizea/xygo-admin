package token

import (
	"context"
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

// 应用名常量
const (
	AppAdmin  = "admin"  // 后台管理员
	AppMember = "member" // 前台会员
)

// Config JWT 配置
type Config struct {
	Secret     string
	Expires    int64 // 秒
	MultiLogin bool  // 是否允许多端登录（false=单点登录，新登录踢掉旧会话）
}

// Claims JWT 载荷（包含完整用户信息，对齐 HotGo）
type Claims struct {
	*model.AuthUser // 嵌入完整用户信息
	jwt.RegisteredClaims
}

// MemberClaims 会员 JWT 载荷
type MemberClaims struct {
	*model.MemberUser
	jwt.RegisteredClaims
}

// getConfig 从配置文件读取 jwt 配置
func getConfig(ctx context.Context) Config {
	secret := g.Cfg().MustGet(ctx, "auth.jwt.secret").String()
	if secret == "" {
		secret = "xygo-secret-key"
	}
	expires := g.Cfg().MustGet(ctx, "auth.jwt.expires").Int64()
	if expires <= 0 {
		expires = 7200
	}
	multiLogin := g.Cfg().MustGet(ctx, "auth.jwt.multiLogin").Bool()
	return Config{
		Secret:     secret,
		Expires:    expires,
		MultiLogin: multiLogin,
	}
}

// TokenMeta Token 元数据（存储在缓存中）
type TokenMeta struct {
	ExpireAt     int64 `json:"exp"`    // token 过期时间
	RefreshAt    int64 `json:"ra"`     // 刷新时间
	RefreshCount int64 `json:"rc"`     // 刷新次数
	Kicked       bool  `json:"kicked"` // 是否被踢下线
}

// ErrTokenKicked 自定义错误：Token 被踢下线
var ErrTokenKicked = fmt.Errorf("token kicked")

// Generate 登录成功后生成 token（后台管理员专用，保持兼容）
func Generate(ctx context.Context, user model.AuthUser) (accessToken string, expiresIn int64, err error) {
	return GenerateAdmin(ctx, user)
}

// GenerateAdmin 后台管理员登录生成 token
func GenerateAdmin(ctx context.Context, user model.AuthUser) (accessToken string, expiresIn int64, err error) {
	cfg := getConfig(ctx)

	now := time.Now()
	expireAt := now.Add(time.Duration(cfg.Expires) * time.Second)

	// ✨ Claims 包含完整用户信息（对齐 HotGo）
	claims := &Claims{
		AuthUser: &user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", 0, err
	}

	// ✨ 使用缓存存储 Token 元数据
	var (
		authKey  = GetAuthKey(accessToken)
		tokenKey = GetTokenKey(AppAdmin, authKey)
		bindKey  = GetBindKey(AppAdmin, user.Id)
		duration = time.Second * gconv.Duration(cfg.Expires)
	)

	// ✨ 单点登录模式：踢掉旧会话（仅在 multiLogin=false 时生效）
	if !cfg.MultiLogin {
		kickOldSession(ctx, AppAdmin, bindKey)
	}

	tokenMeta := &TokenMeta{
		ExpireAt:     expireAt.Unix(),
		RefreshAt:    now.Unix(),
		RefreshCount: 0,
	}

	if err = cache.Instance().Set(ctx, tokenKey, tokenMeta, duration); err != nil {
		return "", 0, err
	}

	if err = cache.Instance().Set(ctx, bindKey, tokenKey, duration); err != nil {
		return "", 0, err
	}

	return accessToken, cfg.Expires, nil
}

// GenerateMember 前台会员登录生成 token
func GenerateMember(ctx context.Context, user model.MemberUser) (accessToken string, expiresIn int64, err error) {
	cfg := getConfig(ctx)

	now := time.Now()
	expireAt := now.Add(time.Duration(cfg.Expires) * time.Second)

	// 会员 Claims
	claims := &MemberClaims{
		MemberUser: &user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", 0, err
	}

	// 使用缓存存储 Token 元数据
	var (
		authKey  = GetAuthKey(accessToken)
		tokenKey = GetTokenKey(AppMember, authKey)
		bindKey  = GetBindKey(AppMember, user.Id)
		duration = time.Second * gconv.Duration(cfg.Expires)
	)

	// ✨ 单点登录模式：踢掉旧会话（仅在 multiLogin=false 时生效）
	if !cfg.MultiLogin {
		kickOldSession(ctx, AppMember, bindKey)
	}

	tokenMeta := &TokenMeta{
		ExpireAt:     expireAt.Unix(),
		RefreshAt:    now.Unix(),
		RefreshCount: 0,
	}

	if err = cache.Instance().Set(ctx, tokenKey, tokenMeta, duration); err != nil {
		return "", 0, err
	}

	if err = cache.Instance().Set(ctx, bindKey, tokenKey, duration); err != nil {
		return "", 0, err
	}

	return accessToken, cfg.Expires, nil
}

// GetAuthKey 生成认证 key（MD5 哈希）
func GetAuthKey(token string) string {
	return gmd5.MustEncryptString("xygo" + token)
}

// GetTokenKey 获取 Token 缓存 key
func GetTokenKey(appName, authKey string) string {
	return fmt.Sprintf("%s:token:%s:%s", consts.CachePrefix, appName, authKey)
}

// GetBindKey 获取用户绑定 key（单点登录）
func GetBindKey(appName string, userId uint64) string {
	return fmt.Sprintf("%s:token:bind:%s:%d", consts.CachePrefix, appName, userId)
}

// Delete 删除 Token（退出登录，后台管理员专用，保持兼容）
func Delete(ctx context.Context, accessToken string) error {
	return DeleteByApp(ctx, AppAdmin, accessToken)
}

// DeleteByApp 按应用类型删除 Token
func DeleteByApp(ctx context.Context, appName string, accessToken string) error {
	var (
		authKey  = GetAuthKey(accessToken)
		tokenKey = GetTokenKey(appName, authKey)
	)

	// 从缓存中删除 Token
	_, err := cache.Instance().Remove(ctx, tokenKey)
	return err
}

// Parse 解析并校验 token（后台管理员专用，保持兼容）
func Parse(ctx context.Context, accessToken string) (*model.AuthUser, error) {
	return ParseAdmin(ctx, accessToken)
}

// ParseAdmin 解析后台管理员 Token
func ParseAdmin(ctx context.Context, accessToken string) (*model.AuthUser, error) {
	cfg := getConfig(ctx)

	// 解析 JWT Token
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

	// 从缓存检查 Token 是否有效
	var (
		authKey  = GetAuthKey(accessToken)
		tokenKey = GetTokenKey(AppAdmin, authKey)
	)

	if err = validateTokenMeta(ctx, tokenKey); err != nil {
		return nil, err
	}

	return claims.AuthUser, nil
}

// ParseMember 解析前台会员 Token
func ParseMember(ctx context.Context, accessToken string) (*model.MemberUser, error) {
	cfg := getConfig(ctx)

	// 解析 JWT Token
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

	// 从缓存检查 Token 是否有效
	var (
		authKey  = GetAuthKey(accessToken)
		tokenKey = GetTokenKey(AppMember, authKey)
	)

	if err = validateTokenMeta(ctx, tokenKey); err != nil {
		return nil, err
	}

	return claims.MemberUser, nil
}

// validateTokenMeta 验证 Token 元数据（缓存中是否存在且未过期）
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

	// ✨ 检查是否被踢下线
	if tokenMeta.Kicked {
		return ErrTokenKicked
	}

	// 检查是否过期
	now := time.Now()
	if tokenMeta.ExpireAt < now.Unix() {
		return jwt.ErrTokenExpired
	}

	return nil
}

// kickOldSession 踢掉旧会话（SSO单点登录核心逻辑）
// 通过 bindKey 找到旧的 tokenKey，将其标记为 kicked
func kickOldSession(ctx context.Context, appName string, bindKey string) {
	oldTokenKey, err := cache.Instance().Get(ctx, bindKey)
	if err != nil || oldTokenKey.IsEmpty() {
		return // 无旧会话
	}

	oldKey := oldTokenKey.String()

	// 将旧 token 标记为"被踢"（而非删除，这样前端可区分"过期"和"被踢"）
	oldMeta, err := cache.Instance().Get(ctx, oldKey)
	if err != nil || oldMeta.IsEmpty() {
		return
	}

	var meta *TokenMeta
	if err = oldMeta.Scan(&meta); err != nil || meta == nil {
		return
	}

	meta.Kicked = true
	// 保留 30 秒，让前端有机会识别"被踢"状态，之后自动清除
	_ = cache.Instance().Set(ctx, oldKey, meta, 30*time.Second)
}

// KickByUserId 按用户ID踢人（管理员强制下线）
func KickByUserId(ctx context.Context, appName string, userId uint64) error {
	bindKey := GetBindKey(appName, userId)
	kickOldSession(ctx, appName, bindKey)
	// 删除 bindKey
	_, err := cache.Instance().Remove(ctx, bindKey)

	// ✨ 通过 WebSocket 实时推送踢人通知
	notifyKickViaWs(appName, userId)

	return err
}

// notifyKickViaWs 通过 WebSocket 推送被踢通知（延迟导入避免循环依赖）
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
