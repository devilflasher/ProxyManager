package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// AppSettings 应用级设置
type AppSettings struct {
	// 托盘设置
	MinimizeToTray bool `json:"minimize_to_tray"` // 关闭时是否最小化到托盘
	ShowTrayIcon   bool `json:"show_tray_icon"`   // 是否显示托盘图标

	// 首次关闭提示设置
	FirstCloseAsked bool `json:"first_close_asked"` // 是否已经询问过首次关闭行为
}

// AppSettingsManager 应用设置管理器
type AppSettingsManager struct {
	settings *AppSettings
	mu       sync.RWMutex
	filePath string
}

// NewAppSettingsManager 创建应用设置管理器
func NewAppSettingsManager(configDir string) *AppSettingsManager {
	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("警告: 无法创建配置目录: %v\n", err)
	}

	settingsPath := filepath.Join(configDir, "app_settings.json")

	asm := &AppSettingsManager{
		settings: &AppSettings{
			MinimizeToTray:  true,  // 默认最小化到托盘
			ShowTrayIcon:    true,  // 默认显示托盘图标
			FirstCloseAsked: false, // 默认未询问过
		},
		filePath: settingsPath,
	}

	// 尝试加载现有设置
	if err := asm.LoadSettings(); err != nil {
		fmt.Printf("警告: 无法加载应用设置: %v，使用默认设置\n", err)
		// 保存默认设置
		if err := asm.SaveSettings(); err != nil {
			fmt.Printf("警告: 无法保存默认设置: %v\n", err)
		}
	}

	return asm
}

// LoadSettings 加载应用设置
func (asm *AppSettingsManager) LoadSettings() error {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	data, err := os.ReadFile(asm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，使用默认设置
			return nil
		}
		return fmt.Errorf("无法读取应用设置文件: %w", err)
	}

	if err := json.Unmarshal(data, asm.settings); err != nil {
		return fmt.Errorf("无法解析应用设置: %w", err)
	}

	return nil
}

// SaveSettings 保存应用设置
func (asm *AppSettingsManager) SaveSettings() error {
	asm.mu.RLock()
	defer asm.mu.RUnlock()

	data, err := json.MarshalIndent(asm.settings, "", "  ")
	if err != nil {
		return fmt.Errorf("无法序列化应用设置: %w", err)
	}

	if err := os.WriteFile(asm.filePath, data, 0644); err != nil {
		return fmt.Errorf("无法写入应用设置文件: %w", err)
	}

	return nil
}

// GetSettings 获取当前设置
func (asm *AppSettingsManager) GetSettings() AppSettings {
	asm.mu.RLock()
	defer asm.mu.RUnlock()
	return *asm.settings
}

// SetMinimizeToTray 设置关闭时行为
func (asm *AppSettingsManager) SetMinimizeToTray(minimize bool) error {
	asm.mu.Lock()
	asm.settings.MinimizeToTray = minimize
	asm.mu.Unlock()

	return asm.SaveSettings()
}

// SetShowTrayIcon 设置是否显示托盘图标
func (asm *AppSettingsManager) SetShowTrayIcon(show bool) error {
	asm.mu.Lock()
	asm.settings.ShowTrayIcon = show
	asm.mu.Unlock()

	return asm.SaveSettings()
}

// SetFirstCloseAsked 设置是否已询问过首次关闭行为
func (asm *AppSettingsManager) SetFirstCloseAsked(asked bool) error {
	asm.mu.Lock()
	asm.settings.FirstCloseAsked = asked
	asm.mu.Unlock()

	return asm.SaveSettings()
}

// IsFirstClose 检查是否是首次关闭
func (asm *AppSettingsManager) IsFirstClose() bool {
	asm.mu.RLock()
	defer asm.mu.RUnlock()
	return !asm.settings.FirstCloseAsked
}
