package utils

import (
	"github.com/google/uuid"
)

// GenerateDeviceID 生成设备ID
func GenerateDeviceID() string {
	return "device-" + uuid.New().String()[:8]
}

// GenerateUserID 生成用户ID（1-1000000）
func GenerateUserID() int {
	// 简化实现，实际可以使用更复杂的逻辑
	return int(uuid.New().ID() % 1000000)
}
