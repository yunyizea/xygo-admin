package sms

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/service"
)

// Sender 短信驱动接口
type Sender interface {
	Send(ctx context.Context, req *SendRequest) (*SendResult, error)
	DriverName() string
}

// SendRequest 发送请求
type SendRequest struct {
	Phone      string
	TemplateId string
	SignName   string
	Params     map[string]string // 阿里云等 key-value 类服务商使用
	ParamList  []string          // 腾讯云等按序号 {1}{2}{3} 的服务商使用
}

// SendResult 发送结果
type SendResult struct {
	Success   bool
	RequestId string
	Code      string
	Message   string
	Driver    string
}

// Manager 短信管理器，负责驱动选择、发送策略、日志记录
type Manager struct {
	config  *SmsConfig
	drivers map[string]Sender
	mu      sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

// Instance 获取全局短信管理器单例
func Instance(ctx ...context.Context) *Manager {
	once.Do(func() {
		c := context.Background()
		if len(ctx) > 0 && ctx[0] != nil {
			c = ctx[0]
		}
		cfg := LoadConfig(c)
		instance = &Manager{
			config:  cfg,
			drivers: make(map[string]Sender),
		}
		for _, dc := range cfg.Drivers {
			drv, err := NewSender(dc)
			if err != nil {
				g.Log().Warningf(c, "[SMS] init driver %s error: %v", dc.Name, err)
				continue
			}
			instance.drivers[dc.Name] = drv
		}
		if len(instance.drivers) > 0 {
			names := make([]string, 0, len(instance.drivers))
			for k := range instance.drivers {
				names = append(names, k)
			}
			g.Log().Infof(c, "[SMS] initialized with drivers: %s, strategy: %s", strings.Join(names, ","), cfg.Strategy)
		} else {
			g.Log().Warning(c, "[SMS] no driver initialized, sms sending will fail")
		}
	})
	return instance
}

// ResetInstance 重置单例（配置变更时调用）
func ResetInstance() {
	once = sync.Once{}
	instance = nil
}

// Send 按策略选择驱动发送短信
func (m *Manager) Send(ctx context.Context, req *SendRequest) (*SendResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.drivers) == 0 {
		return nil, fmt.Errorf("no sms driver available")
	}

	drv, err := m.selectDriver()
	if err != nil {
		g.Log().Warningf(ctx, "[SMS] selectDriver error: %v, available drivers: %d", err, len(m.drivers))
		return nil, err
	}
	g.Log().Infof(ctx, "[SMS] selected driver: %s, phone=%s tplId=%s", drv.DriverName(), req.Phone, req.TemplateId)

	timeoutSec := m.config.Timeout
	if timeoutSec <= 0 {
		timeoutSec = 5
	}
	sendCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	result, err := drv.Send(sendCtx, req)
	if result != nil {
		result.Driver = drv.DriverName()
	}
	g.Log().Infof(ctx, "[SMS] driver %s send done: result=%+v err=%v", drv.DriverName(), result, err)
	return result, err
}

// selectDriver 按策略选择驱动
func (m *Manager) selectDriver() (Sender, error) {
	enabledConfs := m.config.Drivers
	if len(enabledConfs) == 0 {
		return nil, fmt.Errorf("no enabled sms driver")
	}

	switch m.config.Strategy {
	case "random":
		idx := rand.Intn(len(enabledConfs))
		name := enabledConfs[idx].Name
		if drv, ok := m.drivers[name]; ok {
			return drv, nil
		}
	case "weight":
		totalWeight := 0
		for _, dc := range enabledConfs {
			if _, ok := m.drivers[dc.Name]; ok {
				totalWeight += dc.Weight
			}
		}
		if totalWeight <= 0 {
			break
		}
		r := rand.Intn(totalWeight)
		cumulative := 0
		for _, dc := range enabledConfs {
			if drv, ok := m.drivers[dc.Name]; ok {
				cumulative += dc.Weight
				if r < cumulative {
					return drv, nil
				}
			}
		}
	}

	for _, dc := range enabledConfs {
		if drv, ok := m.drivers[dc.Name]; ok {
			return drv, nil
		}
	}
	return nil, fmt.Errorf("no available sms driver")
}

// GetConfig 获取当前配置
func (m *Manager) GetConfig() *SmsConfig {
	return m.config
}

// NewSender 根据驱动配置创建 Sender 实例
func NewSender(conf DriverConf) (Sender, error) {
	switch strings.ToLower(conf.Name) {
	case "aliyun":
		return NewAliyun(conf)
	case "tencent":
		return NewTencent(conf)
	default:
		return nil, fmt.Errorf("unsupported sms driver: %s", conf.Name)
	}
}

// SmsConfig 短信全局配置
type SmsConfig struct {
	Timeout  int          `json:"timeout"`
	Strategy string       `json:"strategy"`
	Drivers  []DriverConf `json:"drivers"`
}

// DriverConf 服务商配置
type DriverConf struct {
	Name      string            `json:"name"`
	Weight    int               `json:"weight"`
	AccessId  string            `json:"accessId"`
	AccessKey string            `json:"accessKey"`
	SignName  string            `json:"signName"`
	Extra     map[string]string `json:"extra"`
}

// LoadConfig 从 sys_config 表的 sms 分组读取短信配置
func LoadConfig(ctx context.Context) *SmsConfig {
	cfg := &SmsConfig{
		Timeout:  5,
		Strategy: "weight",
	}

	sysConf, err := service.SysConfig().GetConfigByGroup(ctx, "sms")
	if err != nil {
		g.Log().Warningf(ctx, "[SMS] read sms config error: %v, using defaults", err)
		return cfg
	}

	if v := sysConf["sms_timeout"]; v != "" {
		fmt.Sscanf(v, "%d", &cfg.Timeout)
	}
	if v := sysConf["sms_strategy"]; v != "" {
		cfg.Strategy = v
	}

	enabledStr := sysConf["sms_enabled_drivers"]
	if enabledStr == "" {
		return cfg
	}
	enabled := strings.Split(enabledStr, ",")

	for _, name := range enabled {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		dc := DriverConf{
			Name:   name,
			Weight: 10,
			Extra:  make(map[string]string),
		}
		switch name {
		case "aliyun":
			dc.AccessId = sysConf["sms_aliyun_access_key_id"]
			dc.AccessKey = sysConf["sms_aliyun_access_key_secret"]
			dc.SignName = sysConf["sms_aliyun_sign_name"]
		case "tencent":
			dc.AccessId = sysConf["sms_tencent_secret_id"]
			dc.AccessKey = sysConf["sms_tencent_secret_key"]
			dc.SignName = sysConf["sms_tencent_sign_name"]
			dc.Extra["appId"] = sysConf["sms_tencent_app_id"]
		}
		cfg.Drivers = append(cfg.Drivers, dc)
	}

	g.Log().Infof(ctx, "[SMS] loaded config: timeout=%d, strategy=%s, drivers=%d", cfg.Timeout, cfg.Strategy, len(cfg.Drivers))
	return cfg
}
