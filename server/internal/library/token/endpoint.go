package token

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/golang-jwt/jwt/v5"

	"xygo/internal/consts"
	"xygo/internal/library/cache"
	"xygo/internal/library/contexts"
)

// EndpointConfig 扩展认证端点配置
type EndpointConfig struct {
	Name       string   // 端点唯一标识，如 "supplier"
	Secret     string   // JWT 签名密钥，留空则复用系统 auth.jwt.secret
	Expires    int64    // 过期时间（秒），0 则复用系统默认
	MultiLogin bool     // 是否允许多端登录
	HeaderKey  string   // Token 请求头名称，默认 "Authorization"
	HeaderType string   // 请求头前缀，默认 "Bearer"（设为空则直接读取整个 header 值）
	LoginPaths []string // 免鉴权路径（如 "/supplier/auth/login"）
}

// EndpointClaims 通用端点 JWT 载荷
type EndpointClaims struct {
	Endpoint string         `json:"ep"`
	UserId   uint64         `json:"uid"`
	Data     map[string]any `json:"data,omitempty"`
	jwt.RegisteredClaims
}

// Endpoint 认证端点实例
type Endpoint struct {
	config EndpointConfig
}

// NewEndpoint 创建认证端点（扩展在 init() 中调用）
func NewEndpoint(cfg EndpointConfig) *Endpoint {
	if cfg.Name == "" {
		panic("token.NewEndpoint: Name is required")
	}
	if cfg.HeaderKey == "" {
		cfg.HeaderKey = "Authorization"
	}
	if cfg.HeaderType == "" && cfg.HeaderKey == "Authorization" {
		cfg.HeaderType = "Bearer"
	}
	return &Endpoint{config: cfg}
}

// getEffectiveConfig 获取实际生效的配置（合并系统默认值）
func (e *Endpoint) getEffectiveConfig(ctx context.Context) (secret string, expires int64, multiLogin bool) {
	sysCfg := getConfig(ctx)

	secret = e.config.Secret
	if secret == "" {
		secret = sysCfg.Secret
	}
	expires = e.config.Expires
	if expires <= 0 {
		expires = sysCfg.Expires
	}
	multiLogin = e.config.MultiLogin
	return
}

// Generate 生成 Token
func (e *Endpoint) Generate(ctx context.Context, userId uint64, data map[string]any) (accessToken string, expiresIn int64, err error) {
	secret, expires, multiLogin := e.getEffectiveConfig(ctx)

	now := time.Now()
	expireAt := now.Add(time.Duration(expires) * time.Second)

	claims := &EndpointClaims{
		Endpoint: e.config.Name,
		UserId:   userId,
		Data:     data,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = tok.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}

	var (
		authKey  = GetAuthKey(accessToken)
		tokenKey = GetTokenKey(e.config.Name, authKey)
		bindKey  = GetBindKey(e.config.Name, userId)
		duration = time.Second * gconv.Duration(expires)
	)

	if !multiLogin {
		kickOldSession(ctx, e.config.Name, bindKey)
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

	return accessToken, expires, nil
}

// Parse 解析并验证 Token，返回用户ID和附加数据
func (e *Endpoint) Parse(ctx context.Context, accessToken string) (userId uint64, data map[string]any, err error) {
	secret, _, _ := e.getEffectiveConfig(ctx)

	parsed, parseErr := jwt.ParseWithClaims(accessToken, &EndpointClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if parseErr != nil || !parsed.Valid {
		return 0, nil, parseErr
	}

	claims, ok := parsed.Claims.(*EndpointClaims)
	if !ok {
		return 0, nil, jwt.ErrTokenMalformed
	}

	if claims.Endpoint != e.config.Name {
		return 0, nil, jwt.ErrTokenMalformed
	}

	var (
		authKey  = GetAuthKey(accessToken)
		tokenKey = GetTokenKey(e.config.Name, authKey)
	)

	if err = validateTokenMeta(ctx, tokenKey); err != nil {
		return 0, nil, err
	}

	return claims.UserId, claims.Data, nil
}

// Delete 删除 Token（退出登录）
func (e *Endpoint) Delete(ctx context.Context, accessToken string) error {
	return DeleteByApp(ctx, e.config.Name, accessToken)
}

// KickByUserId 按用户ID踢人
func (e *Endpoint) KickByUserId(ctx context.Context, userId uint64) error {
	bindKey := GetBindKey(e.config.Name, userId)
	kickOldSession(ctx, e.config.Name, bindKey)
	_, err := cache.Instance().Remove(ctx, bindKey)
	return err
}

// Middleware 生成鉴权中间件
func (e *Endpoint) Middleware() func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		path := r.URL.Path

		customCtx := &contexts.Context{
			Module: e.config.Name,
		}
		contexts.Init(r, customCtx)

		// 检查免鉴权路径
		for _, p := range e.config.LoginPaths {
			if path == p {
				r.Middleware.Next()
				return
			}
		}

		// 提取 Token
		tokenStr := e.extractToken(r)
		if tokenStr == "" {
			r.SetError(gerror.NewCode(consts.CodeNotAuthorized, "未登录"))
			return
		}

		// 解析 Token
		userId, data, err := e.Parse(r.Context(), tokenStr)
		if err != nil {
			if errors.Is(err, ErrTokenKicked) {
				r.SetError(gerror.NewCode(consts.CodeKickedOut, "您的账号已在其他设备登录，请重新登录"))
				return
			}
			r.SetError(gerror.NewCode(consts.CodeNotAuthorized, "登录已失效，请重新登录"))
			return
		}

		// 注入用户信息到上下文
		contexts.SetEndpointUser(r.Context(), e.config.Name, &contexts.EndpointUser{
			Id:   userId,
			Data: data,
		})

		r.Middleware.Next()
	}
}

// extractToken 从请求中提取 Token
func (e *Endpoint) extractToken(r *ghttp.Request) string {
	headerVal := r.Header.Get(e.config.HeaderKey)
	if headerVal == "" {
		return ""
	}
	if e.config.HeaderType != "" {
		return strings.TrimSpace(strings.TrimPrefix(headerVal, e.config.HeaderType+" "))
	}
	return headerVal
}

// Name 返回端点名称
func (e *Endpoint) Name() string {
	return e.config.Name
}
