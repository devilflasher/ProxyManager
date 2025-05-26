package server

import (
	"fmt"
	"sync"

	"proxy-manager-desktop/internal/config"
)

type Proxy interface {
	Start() error
	Stop() error
	IsRunning() bool
	GetConfig() *config.ProxyConfig
}

type ProxyManager struct {
	configManager *config.ConfigManager
	proxies       map[string]Proxy
	mu            sync.RWMutex
}

func NewProxyManager(configManager *config.ConfigManager) *ProxyManager {
	return &ProxyManager{
		configManager: configManager,
		proxies:       make(map[string]Proxy),
	}
}

func (m *ProxyManager) StartProxy(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if proxy, exists := m.proxies[id]; exists && proxy.IsRunning() {
		return fmt.Errorf("代理 %s 已在运行", id)
	}
	proxyConfig, err := m.configManager.GetProxy(id)
	if err != nil {
		return fmt.Errorf("无法获取代理配置: %w", err)
	}
	var proxy Proxy
	switch proxyConfig.Local.Protocol {
	case "http":
		proxy, err = NewHTTPProxy(proxyConfig)
	case "socks5":
		proxy, err = NewSOCKS5Proxy(proxyConfig)
	default:
		return fmt.Errorf("不支持的代理协议: %s", proxyConfig.Local.Protocol)
	}
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}
	if err := proxy.Start(); err != nil {
		return fmt.Errorf("启动代理失败: %w", err)
	}
	m.proxies[id] = proxy
	return nil
}

func (m *ProxyManager) StopProxy(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	proxy, exists := m.proxies[id]
	if !exists {
		return fmt.Errorf("代理 %s 不存在", id)
	}

	if !proxy.IsRunning() {
		return fmt.Errorf("代理 %s 未在运行", id)
	}

	if err := proxy.Stop(); err != nil {
		return fmt.Errorf("停止代理失败: %w", err)
	}
	delete(m.proxies, id)
	return nil
}

func (m *ProxyManager) IsProxyRunning(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proxy, exists := m.proxies[id]
	return exists && proxy.IsRunning()
}

func (m *ProxyManager) GetRunningProxies() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var runningProxies []string
	for id, proxy := range m.proxies {
		if proxy.IsRunning() {
			runningProxies = append(runningProxies, id)
		}
	}
	return runningProxies
}

func (m *ProxyManager) StartAllEnabledProxies() []error {
	var errors []error
	proxies := m.configManager.GetAllProxies()

	for _, proxyConfig := range proxies {
		if !proxyConfig.Enabled {
			continue
		}

		if err := m.StartProxy(proxyConfig.ID); err != nil {
			errors = append(errors, fmt.Errorf("启动代理 %s 失败: %w", proxyConfig.ID, err))
		}
	}
	return errors
}

func (m *ProxyManager) StopAllProxies() []error {
	m.mu.Lock()
	proxyIDs := make([]string, 0, len(m.proxies))
	for id := range m.proxies {
		proxyIDs = append(proxyIDs, id)
	}
	m.mu.Unlock()
	var errors []error
	for _, id := range proxyIDs {
		if err := m.StopProxy(id); err != nil {
			errors = append(errors, fmt.Errorf("停止代理 %s 失败: %w", id, err))
		}
	}

	return errors
}

func (m *ProxyManager) GetProxyStatus(id string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	proxy, exists := m.proxies[id]
	if !exists {
		return nil, fmt.Errorf("代理 %s 不存在", id)
	}
	status := map[string]interface{}{
		"id":        id,
		"running":   proxy.IsRunning(),
		"config":    proxy.GetConfig(),
		"protocol":  proxy.GetConfig().Local.Protocol,
		"listen_ip": proxy.GetConfig().Local.ListenIP,
		"port":      proxy.GetConfig().Local.ListenPort,
	}
	return status, nil
}

func (m *ProxyManager) RefreshProxy(id string) error {
	isRunning := m.IsProxyRunning(id)
	if isRunning {
		if err := m.StopProxy(id); err != nil {
			return fmt.Errorf("停止代理失败: %w", err)
		}
		if err := m.StartProxy(id); err != nil {
			return fmt.Errorf("重新启动代理失败: %w", err)
		}
	}
	return nil
}
