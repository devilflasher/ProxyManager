package main

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"proxy-manager-desktop/internal/config"
	"proxy-manager-desktop/internal/server"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct - 代理管理器应用
type App struct {
	ctx           context.Context
	configManager *config.ConfigManager
	proxyManager  *server.ProxyManager
}

// ProxyConfig 代理配置结构 - 前端接口
type ProxyConfig struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Upstream    UpstreamProxy `json:"upstream"`
	Local       LocalProxy    `json:"local"`
	Enabled     bool          `json:"enabled"`
	Description string        `json:"description,omitempty"`
}

type UpstreamProxy struct {
	Protocol   string `json:"protocol"` // "http" or "socks5"
	Address    string `json:"address"`  // IP:Port
	Username   string `json:"username"`
	Password   string `json:"password"`
	AuthMethod string `json:"auth_method,omitempty"` // "basic", "digest", "ntlm"
}

type LocalProxy struct {
	Protocol   string `json:"protocol"`  // "http" or "socks5"
	ListenIP   string `json:"listen_ip"` // "127.0.0.1" or "::1"
	ListenPort int    `json:"listen_port"`
}

// ProxyStatus 代理状态
type ProxyStatus struct {
	ID      string `json:"id"`
	Running bool   `json:"running"`
	Error   string `json:"error,omitempty"`
}

// ProxyWithStatus 带状态的代理
type ProxyWithStatus struct {
	ProxyConfig
	Running bool `json:"running"`
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 配置文件路径 - 优先使用当前目录
	configPath := "config.json"

	// 如果当前目录没有配置文件，尝试可执行文件目录
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		execDir, err := os.Executable()
		if err != nil {
			log.Printf("获取可执行文件路径失败: %v", err)
		} else {
			configPath = filepath.Join(filepath.Dir(execDir), "config.json")
		}
	}

	log.Printf("使用配置文件路径: %s", configPath)

	// 初始化配置管理器
	a.configManager = config.NewConfigManager(configPath)

	// 初始化代理管理器
	a.proxyManager = server.NewProxyManager(a.configManager)

	// 自动启动标记为AutoStart的代理
	errors := a.startAutoStartProxies()
	if len(errors) > 0 {
		// 只记录实际的错误，减少日志输出
		log.Printf("自动启动代理时有 %d 个错误", len(errors))
	} else {
		log.Println("所有标记为自动启动的代理已启动")
	}

	log.Println("🚀 代理管理器桌面应用已启动")
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	log.Println("正在保存代理状态...")

	// 获取所有代理的当前运行状态
	proxies := a.configManager.GetAllProxies()
	for _, proxy := range proxies {
		isRunning := a.proxyManager.IsProxyRunning(proxy.ID)
		a.configManager.UpdateProxyAutoStart(proxy.ID, isRunning)
	}

	// 保存配置
	if err := a.configManager.SaveConfig(); err != nil {
		log.Printf("保存配置失败: %v", err)
	} else {
		log.Println("代理状态已保存")
	}

	log.Println("应用正在关闭...")
}

// startAutoStartProxies 启动所有标记为自动启动的代理
func (a *App) startAutoStartProxies() []error {
	var errors []error
	proxies := a.configManager.GetAllProxies()

	for _, proxyConfig := range proxies {
		if !proxyConfig.Enabled || !proxyConfig.AutoStart {
			continue
		}

		if err := a.proxyManager.StartProxy(proxyConfig.ID); err != nil {
			errors = append(errors, fmt.Errorf("启动代理 %s 失败: %w", proxyConfig.ID, err))
		}
	}

	return errors
}

// GetAllProxies 获取所有代理配置
func (a *App) GetAllProxies() ([]*ProxyWithStatus, error) {
	proxies := a.configManager.GetAllProxies()
	var result []*ProxyWithStatus

	for _, proxy := range proxies {
		status := a.proxyManager.IsProxyRunning(proxy.ID)
		result = append(result, &ProxyWithStatus{
			ProxyConfig: ProxyConfig{
				ID:          proxy.ID,
				Name:        proxy.Name,
				Upstream:    UpstreamProxy(proxy.Upstream),
				Local:       LocalProxy(proxy.Local),
				Enabled:     proxy.Enabled,
				Description: "",
			},
			Running: status,
		})
	}

	return result, nil
}

// AddProxy 添加新代理
func (a *App) AddProxy(proxy ProxyConfig) (string, error) {
	// 转换为内部配置格式
	internalProxy := &config.ProxyConfig{
		ID:        proxy.ID,
		Name:      proxy.Name,
		Upstream:  config.UpstreamProxy(proxy.Upstream),
		Local:     config.LocalProxy(proxy.Local),
		Enabled:   proxy.Enabled,
		AutoStart: false, // 新添加的代理默认不自动启动
	}

	id, err := a.configManager.AddProxy(internalProxy)
	if err != nil {
		return "", err
	}

	// 保存配置
	if err := a.configManager.SaveConfig(); err != nil {
		return "", fmt.Errorf("保存配置失败: %w", err)
	}

	// 如果代理设置为启用状态，立即启动
	if proxy.Enabled {
		if startErr := a.proxyManager.StartProxy(id); startErr != nil {
			log.Printf("启动新添加的代理失败: %v", startErr)
			// 不返回错误，因为代理已成功添加，只是启动失败
		} else {
			// 启动成功，更新AutoStart状态
			a.configManager.UpdateProxyAutoStart(id, true)
			a.configManager.SaveConfig()
		}
	}

	return id, nil
}

// UpdateProxy 更新代理配置
func (a *App) UpdateProxy(proxy ProxyConfig) error {
	// 获取当前代理配置以保留AutoStart状态
	currentProxy, err := a.configManager.GetProxy(proxy.ID)
	if err != nil {
		return fmt.Errorf("获取当前代理配置失败: %w", err)
	}

	internalProxy := &config.ProxyConfig{
		ID:        proxy.ID,
		Name:      proxy.Name,
		Upstream:  config.UpstreamProxy(proxy.Upstream),
		Local:     config.LocalProxy(proxy.Local),
		Enabled:   proxy.Enabled,
		AutoStart: currentProxy.AutoStart, // 保留原有的AutoStart状态
	}

	if err := a.configManager.UpdateProxy(internalProxy); err != nil {
		return err
	}

	return a.configManager.SaveConfig()
}

// DeleteProxy 删除代理
func (a *App) DeleteProxy(id string) error {
	// 先停止代理
	a.proxyManager.StopProxy(id)

	if err := a.configManager.DeleteProxy(id); err != nil {
		return err
	}

	return a.configManager.SaveConfig()
}

// StartProxy 启动代理
func (a *App) StartProxy(id string) error {
	err := a.proxyManager.StartProxy(id)
	if err == nil {
		// 启动成功，更新AutoStart状态
		a.configManager.UpdateProxyAutoStart(id, true)
		a.configManager.SaveConfig()
	}
	return err
}

// StopProxy 停止代理
func (a *App) StopProxy(id string) error {
	err := a.proxyManager.StopProxy(id)
	if err == nil {
		// 停止成功，更新AutoStart状态
		a.configManager.UpdateProxyAutoStart(id, false)
		a.configManager.SaveConfig()
	}
	return err
}

// StartAllProxies 启动所有代理
func (a *App) StartAllProxies() []string {
	var errorMessages []string
	proxies := a.configManager.GetAllProxies()

	for _, proxy := range proxies {
		if !proxy.Enabled {
			continue
		}

		if err := a.proxyManager.StartProxy(proxy.ID); err != nil {
			errorMessages = append(errorMessages, err.Error())
		} else {
			// 启动成功，更新AutoStart状态
			a.configManager.UpdateProxyAutoStart(proxy.ID, true)
		}
	}

	// 保存配置
	a.configManager.SaveConfig()
	return errorMessages
}

// StopAllProxies 停止所有代理
func (a *App) StopAllProxies() []string {
	var errorMessages []string
	runningProxies := a.proxyManager.GetRunningProxies()

	for _, proxyID := range runningProxies {
		if err := a.proxyManager.StopProxy(proxyID); err != nil {
			errorMessages = append(errorMessages, err.Error())
		} else {
			// 停止成功，更新AutoStart状态
			a.configManager.UpdateProxyAutoStart(proxyID, false)
		}
	}

	// 保存配置
	a.configManager.SaveConfig()
	return errorMessages
}

// GetProxyStatus 获取代理状态
func (a *App) GetProxyStatus(id string) ProxyStatus {
	running := a.proxyManager.IsProxyRunning(id)
	return ProxyStatus{
		ID:      id,
		Running: running,
		Error:   "",
	}
}

// ExportConfig 导出配置为CSV格式
func (a *App) ExportConfig() (string, error) {
	proxies := a.configManager.GetAllProxies()

	// 创建CSV字符串，使用UTF-8 BOM
	var csvData strings.Builder

	// 添加UTF-8 BOM头，确保Excel正确识别中文编码
	csvData.WriteString("\ufeff")

	writer := csv.NewWriter(&csvData)

	// 设置CSV分隔符为逗号（默认）
	writer.Comma = ','

	// 写入标题行
	headers := []string{
		"代理名称", "上游协议", "上游地址", "上游用户名", "上游密码",
		"本地协议", "本地IP", "本地端口", "是否启用",
	}
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("写入CSV标题失败: %w", err)
	}

	// 写入数据行
	for _, proxy := range proxies {
		enabledStr := "否"
		if proxy.Enabled {
			enabledStr = "是"
		}

		row := []string{
			proxy.Name,
			proxy.Upstream.Protocol,
			proxy.Upstream.Address,
			proxy.Upstream.Username,
			proxy.Upstream.Password,
			proxy.Local.Protocol,
			proxy.Local.ListenIP,
			strconv.Itoa(proxy.Local.ListenPort),
			enabledStr,
		}

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("写入CSV数据失败: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV写入出错: %w", err)
	}

	// 记录生成的CSV内容用于调试
	result := csvData.String()
	log.Printf("导出CSV内容（前200字符）: %s", result[:min(200, len(result))])
	log.Printf("导出了 %d 个代理配置", len(proxies))

	return result, nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExportConfigToFile 导出配置到用户选择的文件
func (a *App) ExportConfigToFile() error {
	// 获取CSV数据
	csvData, err := a.ExportConfig()
	if err != nil {
		return fmt.Errorf("生成CSV数据失败: %w", err)
	}

	// 显示保存文件对话框
	options := runtime.SaveDialogOptions{
		DefaultDirectory: "",
		DefaultFilename:  "代理配置.csv",
		Title:            "保存代理配置",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "CSV文件 (*.csv)",
				Pattern:     "*.csv",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	}

	filename, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return fmt.Errorf("保存文件对话框错误: %w", err)
	}

	// 用户取消了保存
	if filename == "" {
		return nil
	}

	// 写入文件
	if err := os.WriteFile(filename, []byte(csvData), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// ImportConfig 导入CSV格式配置
func (a *App) ImportConfig(configData string) error {
	// 移除UTF-8 BOM（如果存在）
	configData = strings.TrimPrefix(configData, "\ufeff")

	reader := csv.NewReader(strings.NewReader(configData))

	// 设置CSV分隔符
	reader.Comma = ','

	// 设置允许不等长的字段（容错处理）
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("解析CSV数据失败: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV文件至少需要包含标题行和一行数据")
	}

	// 跳过标题行，从第二行开始处理数据
	for i, record := range records[1:] {
		if len(record) < 9 {
			log.Printf("第%d行数据不完整，跳过", i+2)
			continue
		}

		// 解析本地端口
		localPort, err := strconv.Atoi(record[7])
		if err != nil {
			log.Printf("第%d行本地端口格式错误: %v，跳过", i+2, err)
			continue
		}

		// 解析是否启用
		enabled := record[8] == "是" || strings.ToLower(record[8]) == "true" || record[8] == "1"

		// 生成新的ID
		id, err := generateUniqueID()
		if err != nil {
			return fmt.Errorf("生成ID失败: %w", err)
		}

		proxy := &config.ProxyConfig{
			ID:   id,
			Name: record[0],
			Upstream: config.UpstreamProxy{
				Protocol:   record[1],
				Address:    record[2],
				Username:   record[3],
				Password:   record[4],
				AuthMethod: "basic",
			},
			Local: config.LocalProxy{
				Protocol:   record[5],
				ListenIP:   record[6],
				ListenPort: localPort,
			},
			Enabled: enabled,
		}

		if _, err := a.configManager.AddProxy(proxy); err != nil {
			log.Printf("导入代理 %s 失败: %v", proxy.Name, err)
		}
	}

	return a.configManager.SaveConfig()
}

// ImportConfigFromFile 从用户选择的文件导入配置
func (a *App) ImportConfigFromFile() error {
	// 显示打开文件对话框
	options := runtime.OpenDialogOptions{
		DefaultDirectory: "",
		Title:            "选择代理配置文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "CSV文件 (*.csv)",
				Pattern:     "*.csv",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	}

	filename, err := runtime.OpenFileDialog(a.ctx, options)
	if err != nil {
		return fmt.Errorf("打开文件对话框错误: %w", err)
	}

	// 用户取消了选择
	if filename == "" {
		return nil
	}

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// 导入配置
	return a.ImportConfig(string(data))
}

// GetStats 获取统计信息
func (a *App) GetStats() map[string]int {
	proxies := a.configManager.GetAllProxies()
	total := len(proxies)
	running := 0
	enabled := 0

	for _, proxy := range proxies {
		if proxy.Enabled {
			enabled++
		}
		if a.proxyManager.IsProxyRunning(proxy.ID) {
			running++
		}
	}

	return map[string]int{
		"total":   total,
		"running": running,
		"stopped": total - running,
		"enabled": enabled,
	}
}

// generateUniqueID 生成唯一ID
func generateUniqueID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
