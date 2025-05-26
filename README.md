<div align="center">

# 代理管理器 - Nobiggie社区特供版

[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg?style=flat&logo=go&logoColor=white)](https://golang.org)
[![Wails](https://img.shields.io/badge/Wails-v2.10-FF6B35.svg?style=flat&logo=go&logoColor=white)](https://wails.io)
[![Windows](https://img.shields.io/badge/Windows-10%2B-0078D6.svg?style=flat&logo=windows&logoColor=white)](https://www.microsoft.com/windows)
[![macOS](https://img.shields.io/badge/macOS-10.15%2B-000000.svg?style=flat&logo=apple&logoColor=white)](https://www.apple.com/macos)
[![License](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)

  <strong>作者：Devilflasher</strong>：<span title="No Biggie Community Founder"></span>
  [![X](https://img.shields.io/badge/X-1DA1F2.svg?style=flat&logo=x&logoColor=white)](https://x.com/DevilflasherX)
[![微信](https://img.shields.io/badge/微信-7BB32A.svg?style=flat&logo=wechat&logoColor=white)](https://x.com/DevilflasherX/status/1781563666485448736 "Devilflasherx")
 [![Telegram](https://img.shields.io/badge/Telegram-0A74DA.svg?style=flat&logo=telegram&logoColor=white)](https://t.me/devilflasher0) （欢迎加入微信群交流）

</div>

> ## ⚠️ 免责声明
> 
> 1. **本软件为开源项目，仅供学习交流使用，不得用于任何闭源商业用途**
> 2. **使用者应遵守当地法律法规，禁止用于任何非法用途**
> 3. **开发者不对因使用本软件导致的直接/间接损失承担任何责任**
> 4. **使用本软件即表示您已阅读并同意本免责声明**

## 工具介绍

代理管理器是一款专门为 `NoBiggie社区` 准备的代理协议转换和管理工具。**核心解决浏览器代理兼容性问题**：当您购买的代理IP是带认证的SOCKS5协议时，由于Chrome浏览器不支持SOCKS5认证，无法直接使用。本工具可以将带认证的SOCKS5代理转换为无认证的本地代理，同时支持HTTP与SOCKS5协议间的相互转换，让您在不同代理类型间自由切换。

**适用场景**：浏览器多开代理配置、爬虫开发调试、批量账号管理、代理池统一管理、网络环境测试等需要灵活使用不同协议代理的场景。

## ❇️ 功能特性

-  高性能架构：基于 Go 语言开发，利用 goroutine 实现高并发代理处理
-  多协议支持：支持 HTTP 和 SOCKS5 代理协议，满足不同使用场景
-  IPv6 完全支持：全面支持 IPv6 地址，包括监听 IPv6、连接 IPv6 上游代理、代理到 IPv6 目标服务器
-  智能管理：支持代理配置的增删改查，实时状态监控
-  状态持久化：自动保存代理运行状态，重启应用后恢复之前的运行状态
-  批量操作：支持批量启动、停止、删除代理，提高管理效率
-  配置导入导出：支持 CSV 格式的配置文件导入导出，便于备份和迁移
-  灵活配置：支持上游代理认证（用户名/密码），本地监听端口自定义
-  实时统计：显示代理总数、运行中数量等统计信息
-  现代界面：采用现代化 Web UI，操作简洁直观
-  响应式设计：界面自适应，支持各种屏幕尺寸

## ❇️ 技术架构

- **后端框架**：Go + Wails v2
- **前端技术**：HTML5 + CSS3 + JavaScript
- **代理协议**：HTTP/HTTPS + SOCKS5
- **数据存储**：JSON 配置文件
- **并发模型**：Go goroutines + channels
- **跨平台**：支持 Windows 10+ 和 macOS 10.15+

## ❇️ 环境要求

### Windows 系统
- Windows 10/11 (64-bit)
- Go 1.22+ (开发环境)
- Node.js 16+ (开发环境)
- 已安装 [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### macOS 系统  
- macOS 10.15+ (Catalina 或更高版本)
- Intel Mac 或 Apple Silicon
- Go 1.22+ (开发环境)
- Node.js 16+ (开发环境)
- 已安装 [Wails CLI](https://wails.io/docs/gettingstarted/installation)

## ⚠️ Windows 使用说明

### Windows Defender 警告处理
首次运行时可能出现"Windows 已保护你的电脑"警告，这是正常现象：

1. **原因**：程序暂未进行代码签名，Windows 会对未知发行者显示警告
2. **解决**：点击"更多信息" → "仍要运行"
3. **安全性**：本程序开源透明，代码可在 GitHub 查看

### 杀毒软件误报
某些杀毒软件可能误报，请添加程序到白名单或信任列表。

## ❇️ 如何编译使用

### Windows 系统构建

1. **安装开发环境**
   ```bash
   # 安装 Go 
      从 https://golang.org/dl/ 下载
   # 安装 Node.js 
      从 https://nodejs.org/ 下载
   # 安装 Wails CLI
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

2. **安装依赖**
   ```bash
   命令提示符进入程序文件夹后
   # 安装 Go 依赖
   go mod tidy
   
   # 安装前端依赖
   cd frontend
   npm install
   cd ..
   ```

3. **编译生成可执行文件**
   ```bash
   # 生产模式编译
   wails build
   
   # 编译后的文件在 build/bin/ 目录下，可直接复制到电脑中其它位置单独使用
   ```

### macOS 系统构建


1. **安装环境和依赖**

   ```bash
   # 安装 Go (如果未安装)
   下载安装：https://go.dev/dl/


   
   # 安装 Node.js (如果未安装)
   下载安装：https://nodejs.org/zh-cn

   #在程序所在目录打开终端窗口
   程序文件夹上点击右键-服务-新建位于文件夹位置的终端窗口
   
   # 运行命令安装 Wails CLI
   go install github.com/wailsapp/wails/v2/cmd/wails@latest

   若提示命令不可用，输入以下两个命令添加环境变量：
   echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.zprofile
   source ~/.zprofile

   若提示网络问题无法下载，请在代理工具中复制环境变量，然后在终端中运行

   # 安装 Go 依赖
   go mod tidy
   
   # 安装前端依赖（依次输入）
   cd frontend
   npm install
   cd ..

   # ⚠️ 如果在构建时遇到vite权限错误，运行以下命令修复：
   chmod +x frontend/node_modules/.bin/vite
   chmod +x frontend/node_modules/.bin/*
   ```

2. **构建应用**
   ```bash
   # 标准构建
   wails build
   
   # 或指定特定架构
   wails build -platform darwin/amd64    # Intel Mac
   wails build -platform darwin/arm64    # Apple Silicon
   ```

4. **查看输出**
   构建完成后，应用程序将在 `build/bin/` 目录中生成 `.app` 文件。可直接复制到电脑中其它位置单独使用



## ❇️ 使用说明

### 基本操作

1. **添加代理**
   - 点击右上角"添加代理"按钮
   - 填写代理名称、上游代理信息（协议、地址、认证信息）
   - 配置本地监听信息（协议、IP、端口）
   - 点击"保存"完成添加

2. **启动/停止代理**
   - 单个代理：点击代理项目右侧的"启动"或"停止"按钮
   - 批量操作：勾选多个代理，使用工具栏的"启动选中"或"停用选中"

3. **编辑代理**
   - 点击代理项目右侧的"编辑"按钮
   - 修改配置信息后保存

4. **删除代理**
   - 单个删除：点击代理项目右侧的"删除"按钮
   - 批量删除：勾选多个代理，点击工具栏的"删除选中"

### 配置管理

1. **导出配置**
   - 点击"📤 导出配置"按钮
   - 选择保存位置，配置将以CSV格式导出

2. **批量导入**
   - 点击"📥 批量导入"按钮
   - 选择CSV配置文件进行批量导入
   - CSV格式：名称,上游协议,上游地址,用户名,密码,本地协议,本地IP,本地端口
   - 若第一次倒入没有模板，可以导出一次获取模板

3. **CSV文件格式示例**
   ```csv
   名称,上游协议,上游地址,用户名,密码,本地协议,本地IP,本地端口
   测试代理1,http,127.0.0.1:8080,user1,pass1,http,127.0.0.1,9001
   测试代理2,socks5,127.0.0.1:1080,,,socks5,127.0.0.1,9002
   IPv6代理,socks5,[2001:db8::1]:1080,user,pass,http,::1,9003
   ```

### 状态持久化

- **自动保存**：程序会自动记住每个代理的运行状态
- **启动恢复**：重新启动程序时，之前运行的代理会自动恢复运行
- **智能管理**：只有手动启动的代理才会在下次启动时自动运行

## ❇️ 配置文件说明

程序使用JSON格式存储配置，默认文件为 `config.json`：

```json
{
  "proxies": [
    {
      "id": "unique-id",
      "name": "代理名称",
      "upstream": {
        "protocol": "http",
        "address": "127.0.0.1:8080",
        "username": "用户名",
        "password": "密码"
      },
      "local": {
        "protocol": "http",
        "listen_ip": "127.0.0.1",
        "listen_port": 9001
      },
      "enabled": true,
      "auto_start": false
    }
  ],
  "proxy_order": ["unique-id"]
}
```

### IPv6 配置示例

程序完全支持IPv6地址配置，以下是IPv6使用示例：

```json
{
  "proxies": [
    {
      "id": "ipv6-example",
      "name": "IPv6代理示例",
      "upstream": {
        "protocol": "socks5",
        "address": "[2001:db8::1]:1080",
        "username": "user",
        "password": "pass"
      },
      "local": {
        "protocol": "http",
        "listen_ip": "::1",
        "listen_port": 8080
      },
      "enabled": true,
      "auto_start": false
    }
  ]
}
```

**IPv6地址格式说明**：
- 监听IPv6：`listen_ip` 使用 `::1`（本地）或 `::`（所有接口）
- 上游IPv6代理：`address` 使用 `[IPv6地址]:端口` 格式，如 `[2001:db8::1]:8080`
- 支持IPv4/IPv6混合环境，可以任意组合使用

## ❇️ 性能特性

- **高并发**：基于 Go goroutines，支持数百个并发代理连接
- **低延迟**：使用 `io.Copy` 实现高效数据转发
- **内存优化**：智能内存管理，长时间运行稳定
- **线程安全**：使用 `sync.RWMutex` 保证并发安全
- **非阻塞设计**：UI操作不会因代理启停而卡顿

## 常见问题

1. **程序无法启动❓️**
   - 检查是否有足够的权限运行程序
   - 确保没有杀毒软件误报拦截
   - 检查系统版本是否满足程序要求

2. **代理无法启动❓️**
   - 检查本地监听端口是否被占用
   - 验证上游代理地址和认证信息是否正确

3. **配置文件损坏❓️**
   - 删除 `config.json` 文件，程序会自动创建新的配置
   - 检查JSON格式是否正确
   - 使用导出功能备份配置以防数据丢失

4. **批量导入失败❓️**
   - 检查CSV文件格式是否正确
   - 确保CSV文件编码为UTF-8
   - 验证端口号范围是否正确（1-65535）

5. **macOS构建时vite权限错误❓️**
   - 错误信息：`Permission denied: vite`
   - 解决方案：运行 `chmod +x frontend/node_modules/.bin/*`
   - 原因：npm安装的二进制文件缺少执行权限

## ❇️ 更新日志

### v1.0 （2025-5-27） 🆕
 **✨ 首次发布**
  - 支持 HTTP 和 SOCKS5 代理协议
  - 实现代理配置的增删改查功能
  - 支持批量操作和配置导入导出
  - 基于 Wails 框架的现代化界面
  - 状态持久化和自动恢复功能
  - 高性能并发代理处理
  - 完整的错误处理和日志记录

## 许可证

本项目采用 GPL-3.0 License，保留所有权利。使用本代码需明确标注来源，禁止闭源商业使用。


