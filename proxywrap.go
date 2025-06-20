package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s claude chat\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  PROXY_HOST (default: 127.0.0.1)\n")
		fmt.Fprintf(os.Stderr, "  PROXY_PORT (default: 7890)\n")
		fmt.Fprintf(os.Stderr, "  PROXY_TYPE (default: http)\n")
		os.Exit(1)
	}

	// 获取代理配置
	proxyHost := getEnvWithDefault("PROXY_HOST", "127.0.0.1")
	proxyPort := getEnvWithDefault("PROXY_PORT", "7890")
	proxyType := getEnvWithDefault("PROXY_TYPE", "http")
	
	proxyURL := fmt.Sprintf("%s://%s:%s", proxyType, proxyHost, proxyPort)

	// 设置所有可能的代理环境变量
	proxyEnvs := map[string]string{
		"http_proxy":  proxyURL,
		"https_proxy": proxyURL,
		"HTTP_PROXY":  proxyURL,
		"HTTPS_PROXY": proxyURL,
		"ALL_PROXY":   proxyURL,
		"ftp_proxy":   proxyURL,
		"FTP_PROXY":   proxyURL,
	}

	// 准备执行的命令
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	// 查找命令的完整路径
	cmdPath, err := exec.LookPath(cmdName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: command '%s' not found: %v\n", cmdName, err)
		os.Exit(1)
	}

	// 准备环境变量
	env := os.Environ()
	
	// 添加代理环境变量
	for key, value := range proxyEnvs {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// 为 Node.js 程序添加特殊支持
	if isNodeProgram(cmdPath) {
		// 尝试创建临时的代理初始化文件
		proxyInitFile, err := createProxyInitFile()
		if err == nil {
			// 检查是否已有 NODE_OPTIONS
			nodeOptions := ""
			for _, envVar := range env {
				if strings.HasPrefix(envVar, "NODE_OPTIONS=") {
					nodeOptions = strings.TrimPrefix(envVar, "NODE_OPTIONS=")
					break
				}
			}
			
			// 添加 require 钩子
			if nodeOptions != "" {
				nodeOptions += " "
			}
			nodeOptions += "--require " + proxyInitFile
			
			// 更新 NODE_OPTIONS
			env = updateOrAddEnv(env, "NODE_OPTIONS", nodeOptions)
		}
	}

	fmt.Printf("🔗 Using proxy: %s\n", proxyURL)
	fmt.Printf("🚀 Executing: %s %s\n", cmdName, strings.Join(cmdArgs, " "))

	// 执行命令
	err = syscall.Exec(cmdPath, append([]string{cmdName}, cmdArgs...), env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func isNodeProgram(cmdPath string) bool {
	// 读取文件的前几行来检测是否是 Node.js 程序
	if strings.Contains(cmdPath, "node") {
		return true
	}
	
	// 检查是否是 npm 全局安装的程序
	if strings.Contains(cmdPath, ".nvm") || strings.Contains(cmdPath, "node_modules") {
		return true
	}
	
	// 读取文件开头检查 shebang
	file, err := os.Open(cmdPath)
	if err != nil {
		return false
	}
	defer file.Close()
	
	buffer := make([]byte, 100)
	n, err := file.Read(buffer)
	if err != nil {
		return false
	}
	
	content := string(buffer[:n])
	return strings.Contains(content, "#!/usr/bin/env node") || 
		   strings.Contains(content, "#!/usr/bin/node") ||
		   strings.Contains(content, "node")
}

func updateOrAddEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, envVar := range env {
		if strings.HasPrefix(envVar, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func createProxyInitFile() (string, error) {
	// 创建临时文件
	tmpDir := os.TempDir()
	proxyInitContent := "// Auto-generated proxy initialization\n" +
		"try {\n" +
		"  // 方法1: 尝试使用 global-agent (如果已安装)\n" +
		"  require('global-agent').bootstrap();\n" +
		"  console.log('[proxywrap] global-agent enabled');\n" +
		"} catch (e1) {\n" +
		"  try {\n" +
		"    // 方法2: 尝试使用 undici ProxyAgent\n" +
		"    const { setGlobalDispatcher, ProxyAgent } = require('undici');\n" +
		"    const proxyUrl = process.env.http_proxy || process.env.HTTP_PROXY;\n" +
		"    if (proxyUrl) {\n" +
		"      setGlobalDispatcher(new ProxyAgent(proxyUrl));\n" +
		"      console.log('[proxywrap] undici ProxyAgent enabled');\n" +
		"    }\n" +
		"  } catch (e2) {\n" +
		"    // 方法3: 劫持 http/https 模块\n" +
		"    try {\n" +
		"      const http = require('http');\n" +
		"      const https = require('https');\n" +
		"      const url = require('url');\n" +
		"      \n" +
		"      const proxyUrl = process.env.http_proxy || process.env.HTTP_PROXY;\n" +
		"      if (proxyUrl) {\n" +
		"        const proxy = url.parse(proxyUrl);\n" +
		"        \n" +
		"        // 劫持 http.request\n" +
		"        const originalHttpRequest = http.request;\n" +
		"        http.request = function(options, callback) {\n" +
		"          if (typeof options === 'string') {\n" +
		"            options = url.parse(options);\n" +
		"          }\n" +
		"          options = Object.assign({}, options);\n" +
		"          options.host = proxy.hostname;\n" +
		"          options.port = proxy.port;\n" +
		"          options.path = 'http://' + (options.hostname || options.host) + ':' + (options.port || 80) + (options.path || '/');\n" +
		"          options.headers = options.headers || {};\n" +
		"          options.headers['Host'] = (options.hostname || options.host) + ':' + (options.port || 80);\n" +
		"          return originalHttpRequest(options, callback);\n" +
		"        };\n" +
		"        \n" +
		"        console.log('[proxywrap] HTTP module hijack enabled');\n" +
		"      }\n" +
		"    } catch (e3) {\n" +
		"      console.log('[proxywrap] No proxy method available, relying on environment variables');\n" +
		"    }\n" +
		"  }\n" +
		"}\n"
	
	proxyInitFile := filepath.Join(tmpDir, "proxywrap-init.js")
	err := ioutil.WriteFile(proxyInitFile, []byte(proxyInitContent), 0644)
	if err != nil {
		return "", err
	}
	
	return proxyInitFile, nil
}
