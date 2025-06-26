# Command Proxy Wrapper

一个用于为任意命令设置代理的轻量级工具，专门解决 ClaudeCode 和 GeminiCode 等工具的网络代理问题。
![image](https://github.com/user-attachments/assets/138a763b-ad46-45b6-851f-fc3a078478f5)

## 项目目标

解决 ClaudeCode、GeminiCode 及其他命令行工具在受限网络环境下的连接问题，通过自动设置代理环境变量和 Node.js 程序的特殊处理，确保命令能够通过代理正常访问网络。

## 安装

### 远程安装（推荐）

```bash
go install github.com/yunfengsay/command-proxy-wrapper@latest
```

### 本地构建

```bash
go build -o proxywrap proxywrap.go
```

## 用法

### 基本用法

```bash
./proxywrap <command> [args...]
```

### 示例

```bash
# 使用代理运行 ClaudeCode 
./proxywrap claude

# 使用代理运行 Gemini  
./proxywrap gemini

# 使用代理运行其他命令
./proxywrap curl https://api.anthropic.com
./proxywrap npm install
```

### 环境变量配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PROXY_HOST` | 127.0.0.1 | 代理服务器地址 |
| `PROXY_PORT` | 7890 | 代理服务器端口 |
| `PROXY_TYPE` | http | 代理类型 (http/socks5) |

### 示例配置

```bash
export PROXY_HOST=127.0.0.1
export PROXY_PORT=7890
export PROXY_TYPE=http

./proxywrap claude 
```

## 特性

- 自动设置所有常见的代理环境变量 (http_proxy, https_proxy, etc.)
- 对 Node.js 程序提供特殊支持，自动注入代理配置
- 支持多种代理类型 (HTTP, SOCKS5)
- 轻量级，单个二进制文件
- 跨平台支持
