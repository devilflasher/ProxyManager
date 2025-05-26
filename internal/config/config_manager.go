package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// ConfigManager 负责配置的加载、保存和管理
type ConfigManager struct {
	proxies    map[string]*ProxyConfig
	proxyOrder []string // 维护代理的顺序
	mu         sync.RWMutex
	filePath   string
}

// NewConfigManager 创建一个新的配置管理器
func NewConfigManager(filePath string) *ConfigManager {
	// 确保配置目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("警告: 无法创建配置目录: %v\n", err)
	}

	cm := &ConfigManager{
		proxies:    make(map[string]*ProxyConfig),
		proxyOrder: make([]string, 0),
		filePath:   filePath,
	}

	if err := cm.LoadConfig(); err != nil {
		fmt.Printf("警告: 无法加载初始配置: %v\n", err)
		// 如果加载失败或文件不存在，则初始化为空配置
		cm.proxies = make(map[string]*ProxyConfig)
	}

	return cm
}

// LoadConfig 从文件加载配置
func (cm *ConfigManager) LoadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			cm.proxies = make(map[string]*ProxyConfig)
			cm.proxyOrder = make([]string, 0)
			return nil // 文件不存在，全新开始
		}
		return fmt.Errorf("无法读取配置文件: %w", err)
	}

	var proxies []*ProxyConfig

	if strings.HasSuffix(cm.filePath, ".yaml") || strings.HasSuffix(cm.filePath, ".yml") {
		if err := yaml.Unmarshal(data, &proxies); err != nil {
			return fmt.Errorf("无法解析YAML配置: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &proxies); err != nil {
			return fmt.Errorf("无法解析JSON配置: %w", err)
		}
	}

	// 将数组转换为映射表，并维护顺序
	newProxies := make(map[string]*ProxyConfig)
	newProxyOrder := make([]string, 0, len(proxies))
	for _, proxy := range proxies {
		newProxies[proxy.ID] = proxy
		newProxyOrder = append(newProxyOrder, proxy.ID)
	}
	cm.proxies = newProxies
	cm.proxyOrder = newProxyOrder

	return nil
}

// SaveConfig 保存配置到文件
func (cm *ConfigManager) SaveConfig() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 将映射表转换为数组，按照维护的顺序
	var proxies []*ProxyConfig
	for _, id := range cm.proxyOrder {
		if proxy, exists := cm.proxies[id]; exists {
			proxies = append(proxies, proxy)
		}
	}

	var (
		data []byte
		err  error
	)

	if strings.HasSuffix(cm.filePath, ".yaml") || strings.HasSuffix(cm.filePath, ".yml") {
		data, err = yaml.Marshal(proxies)
		if err != nil {
			return fmt.Errorf("无法序列化为YAML: %w", err)
		}
	} else {
		data, err = json.MarshalIndent(proxies, "", "  ")
		if err != nil {
			return fmt.Errorf("无法序列化为JSON: %w", err)
		}
	}

	if err := os.WriteFile(cm.filePath, data, 0644); err != nil {
		return fmt.Errorf("无法写入配置文件: %w", err)
	}

	return nil
}

// AddProxy 添加一个新的代理配置
func (cm *ConfigManager) AddProxy(proxy *ProxyConfig) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if proxy.ID == "" {
		id, err := generateUniqueID()
		if err != nil {
			return "", fmt.Errorf("无法生成唯一ID: %w", err)
		}
		proxy.ID = id
	} else if _, exists := cm.proxies[proxy.ID]; exists {
		return "", fmt.Errorf("ID '%s' 已存在", proxy.ID)
	}

	cm.proxies[proxy.ID] = proxy
	cm.proxyOrder = append(cm.proxyOrder, proxy.ID)
	return proxy.ID, nil
}

// UpdateProxy 更新代理配置
func (cm *ConfigManager) UpdateProxy(proxy *ProxyConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.proxies[proxy.ID]; !exists {
		return fmt.Errorf("ID '%s' 不存在", proxy.ID)
	}

	cm.proxies[proxy.ID] = proxy
	return nil
}

// DeleteProxy 删除代理配置
func (cm *ConfigManager) DeleteProxy(id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.proxies[id]; !exists {
		return fmt.Errorf("ID '%s' 不存在", id)
	}

	delete(cm.proxies, id)
	// 从顺序数组中移除
	for i, existingID := range cm.proxyOrder {
		if existingID == id {
			cm.proxyOrder = append(cm.proxyOrder[:i], cm.proxyOrder[i+1:]...)
			break
		}
	}
	return nil
}

// GetProxy 获取代理配置
func (cm *ConfigManager) GetProxy(id string) (*ProxyConfig, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	proxy, exists := cm.proxies[id]
	if !exists {
		return nil, fmt.Errorf("ID '%s' 不存在", id)
	}

	return proxy, nil
}

// GetAllProxies 获取所有代理配置，按照添加顺序返回
func (cm *ConfigManager) GetAllProxies() []*ProxyConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var proxies []*ProxyConfig
	for _, id := range cm.proxyOrder {
		if proxy, exists := cm.proxies[id]; exists {
			proxies = append(proxies, proxy)
		}
	}

	return proxies
}

// UpdateProxyAutoStart 更新代理的自动启动状态
func (cm *ConfigManager) UpdateProxyAutoStart(id string, autoStart bool) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	proxy, exists := cm.proxies[id]
	if !exists {
		return fmt.Errorf("ID '%s' 不存在", id)
	}

	proxy.AutoStart = autoStart
	return nil
}

// SaveProxyStates 批量保存代理的运行状态
func (cm *ConfigManager) SaveProxyStates(states map[string]bool) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for id, autoStart := range states {
		if proxy, exists := cm.proxies[id]; exists {
			proxy.AutoStart = autoStart
		}
	}

	return nil
}

// generateUniqueID 生成唯一ID
func generateUniqueID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
