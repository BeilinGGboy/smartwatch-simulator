package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// 测试服务器连接（仅发送 3 条数据，用于验证数据库写入）
func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	
	fmt.Println("========================================")
	fmt.Println("   服务器连接测试")
	fmt.Println("========================================")
	fmt.Printf("服务器地址: %s\n\n", serverURL)
	
	// 1. 测试健康检查
	fmt.Println("1. 测试健康检查接口...")
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 健康检查通过")
	} else {
		fmt.Printf("⚠️  健康检查返回状态码: %d\n", resp.StatusCode)
	}
	fmt.Println()
	
	// 2. 测试批量上传接口（仅 3 条数据，用于验证数据库写入）
	fmt.Println("2. 测试批量上传接口（3 条测试数据）...")
	now := time.Now()
	sleepStart := now.Add(-8 * time.Hour)
	sleepEnd := now.Add(-2 * time.Hour)
	testData := []map[string]interface{}{
		{
			"device_id": "test-device-001",
			"user_id":   1001,
			"timestamp": now.Format(time.RFC3339),
			"heart_rate": map[string]interface{}{
				"bpm":       75,
				"timestamp": now.Format(time.RFC3339),
			},
		},
		{
			"device_id": "test-device-001",
			"user_id":   1001,
			"timestamp": now.Format(time.RFC3339),
			"steps": map[string]interface{}{
				"steps":     1000,
				"distance":  700.0,
				"calories":  30,
				"timestamp": now.Format(time.RFC3339),
			},
		},
		{
			"device_id": "test-device-001",
			"user_id":   1001,
			"timestamp": now.Format(time.RFC3339),
			"sleep": map[string]interface{}{
				"sleep_start":   sleepStart.Format(time.RFC3339),
				"sleep_end":     sleepEnd.Format(time.RFC3339),
				"duration":      360,
				"deep_sleep":    120,
				"light_sleep":   240,
				"sleep_quality": 85,
			},
		},
	}
	
	jsonData, err := json.Marshal(testData)
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		return
	}
	
	req, err := http.NewRequest("POST", serverURL+"/api/v1/data/batch", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 批量上传测试成功")
		
		// 读取响应
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		fmt.Printf("   响应: %+v\n", result)
	} else {
		fmt.Printf("⚠️  上传返回状态码: %d\n", resp.StatusCode)
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		fmt.Printf("   响应内容: %s\n", string(body[:n]))
	}
	fmt.Println()
	
	// 3. 测试统计接口
	fmt.Println("3. 测试统计接口...")
	resp, err = http.Get(serverURL + "/api/v1/stats")
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == 200 {
		var stats map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&stats)
		fmt.Println("✅ 统计接口正常")
		fmt.Printf("   统计信息: %+v\n", stats)
	} else {
		fmt.Printf("⚠️  统计接口返回状态码: %d\n", resp.StatusCode)
	}
	
	fmt.Println("\n========================================")
	fmt.Println("   测试完成")
	fmt.Println("========================================")
}
