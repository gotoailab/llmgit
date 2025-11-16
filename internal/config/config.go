package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 项目配置
type Config struct {
	Provider     string `json:"provider"`
	APIKey       string `json:"api_key"`
	Model        string `json:"model"`
	CommitPrompt string `json:"commit_prompt,omitempty"` // 自定义 commit prompt 模板
}

const configDirName = ".llmgit"
const configFileName = "config.json"

// getConfigDir 获取配置目录路径
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户主目录: %w", err)
	}
	return filepath.Join(homeDir, configDirName), nil
}

// getConfigPath 获取配置文件路径
func getConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

// Load 加载配置
func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件不存在，请先运行 'llmgit init'")
		}
		return nil, fmt.Errorf("无法读取配置文件: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("无法解析配置文件: %w", err)
	}

	return &cfg, nil
}

// Save 保存配置
func Save(cfg *Config) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("无法创建配置目录: %w", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("无法序列化配置: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("无法写入配置文件: %w", err)
	}

	return nil
}

