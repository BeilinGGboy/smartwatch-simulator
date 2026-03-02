package config

import (
	"time"
)

// Config 模拟器配置
type Config struct {
	// 设备配置
	DeviceCount      int           `json:"device_count"`       // 设备数量
	DeviceStartDelay time.Duration `json:"device_start_delay"` // 设备启动延迟（避免同时启动）
	
	// 数据生成配置
	DataInterval     time.Duration `json:"data_interval"`     // 数据生成间隔
	BatchSize        int           `json:"batch_size"`        // 批量上传大小
	UploadInterval   time.Duration `json:"upload_interval"`   // 上传间隔
	
	// 服务器配置
	ServerURL        string        `json:"server_url"`        // 服务器地址
	APIEndpoint      string        `json:"api_endpoint"`      // API 端点
	ConcurrentUpload int           `json:"concurrent_upload"` // 并发上传数
	
	// 数据生成策略
	RealTimeRatio    float64       `json:"realtime_ratio"`   // 实时上传比例（0-1）
	BatchRatio       float64       `json:"batch_ratio"`      // 批量上传比例（0-1）
	TimerRatio       float64       `json:"timer_ratio"`      // 定时上传比例（0-1）
	
	// 运行时间
	Duration         time.Duration `json:"duration"`          // 运行时长（0表示持续运行）
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DeviceCount:      100,                       // 测试用：100个设备（可调整为100000）
		DeviceStartDelay: 10 * time.Millisecond,    // 每10ms启动一个设备
		DataInterval:     1 * time.Second,           // 每秒生成一次数据
		BatchSize:        100,                      // 每批100条数据
		UploadInterval:   5 * time.Second,          // 每5秒上传一次
		ServerURL:        "http://192.168.0.102:8080", // 服务器地址
		APIEndpoint:      "/api/v1/data/batch",     // 批量上传端点
		ConcurrentUpload: 10,                       // 测试用：10个并发上传（可调整为100）
		RealTimeRatio:    0.2,                       // 20%实时上传
		BatchRatio:       0.7,                       // 70%批量上传
		TimerRatio:       0.1,                       // 10%定时上传
		Duration:         0,                         // 持续运行
	}
}
