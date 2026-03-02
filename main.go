package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	
	"smartwatch-simulator/simulator/config"
	"smartwatch-simulator/simulator/device"
	"smartwatch-simulator/simulator/uploader"
	"smartwatch-simulator/utils"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("   智能手表数据模拟器")
	fmt.Println("========================================")
	
	// 加载配置
	cfg := config.DefaultConfig()
	
	// 可以根据命令行参数调整配置
	if len(os.Args) > 1 {
		// 简化处理，实际可以使用 flag 包
		fmt.Printf("使用自定义配置（当前使用默认配置）\n")
	}
	
	fmt.Printf("\n配置信息:\n")
	fmt.Printf("  设备数量: %d\n", cfg.DeviceCount)
	fmt.Printf("  数据间隔: %v\n", cfg.DataInterval)
	fmt.Printf("  批量大小: %d\n", cfg.BatchSize)
	fmt.Printf("  上传间隔: %v\n", cfg.UploadInterval)
	fmt.Printf("  服务器地址: %s\n", cfg.ServerURL)
	fmt.Printf("  并发上传数: %d\n", cfg.ConcurrentUpload)
	fmt.Printf("  实时上传比例: %.1f%%\n", cfg.RealTimeRatio*100)
	fmt.Printf("  批量上传比例: %.1f%%\n", cfg.BatchRatio*100)
	fmt.Printf("  定时上传比例: %.1f%%\n", cfg.TimerRatio*100)
	fmt.Println()
	
	// 创建上传器
	uploaderInstance := uploader.NewUploader(
		cfg.ServerURL,
		cfg.APIEndpoint,
		cfg.ConcurrentUpload,
	)
	defer uploaderInstance.Stop()
	
	// 创建设备
	devices := make([]*device.Device, 0, cfg.DeviceCount)
	
	// 计算各模式设备数量
	realtimeCount := int(float64(cfg.DeviceCount) * cfg.RealTimeRatio)
	batchCount := int(float64(cfg.DeviceCount) * cfg.BatchRatio)
	timerCount := cfg.DeviceCount - realtimeCount - batchCount
	
	fmt.Printf("正在创建 %d 个设备...\n", cfg.DeviceCount)
	fmt.Printf("  实时上传设备: %d\n", realtimeCount)
	fmt.Printf("  批量上传设备: %d\n", batchCount)
	fmt.Printf("  定时上传设备: %d\n", timerCount)
	fmt.Println()
	
	// 创建设备并分配模式
	deviceIndex := 0
	for i := 0; i < realtimeCount; i++ {
		dev := device.NewDevice(
			utils.GenerateDeviceID(),
			utils.GenerateUserID(),
			device.ModeRealTime,
		)
		devices = append(devices, dev)
		deviceIndex++
	}
	
	for i := 0; i < batchCount; i++ {
		dev := device.NewDevice(
			utils.GenerateDeviceID(),
			utils.GenerateUserID(),
			device.ModeBatch,
		)
		devices = append(devices, dev)
		deviceIndex++
	}
	
	for i := 0; i < timerCount; i++ {
		dev := device.NewDevice(
			utils.GenerateDeviceID(),
			utils.GenerateUserID(),
			device.ModeTimer,
		)
		devices = append(devices, dev)
		deviceIndex++
	}
	
	// 启动设备（分批启动，避免同时启动造成压力）
	fmt.Printf("正在启动设备...\n")
	startTime := time.Now()
	
	var wg sync.WaitGroup
	for i, dev := range devices {
		wg.Add(1)
		go func(idx int, d *device.Device) {
			defer wg.Done()
			// 延迟启动
			time.Sleep(time.Duration(idx) * cfg.DeviceStartDelay)
			
			// 根据模式启动
			d.Start(cfg.DataInterval, cfg.UploadInterval, cfg.BatchSize)
			
			// 连接上传通道
			go func() {
				for data := range d.UploadChan {
					uploaderInstance.Upload(data)
				}
			}()
		}(i, dev)
	}
	
	wg.Wait()
	startDuration := time.Since(startTime)
	fmt.Printf("所有设备启动完成，耗时: %v\n", startDuration)
	fmt.Println()
	
	// 启动统计监控
	stopStats := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-stopStats:
				return
			case <-ticker.C:
				printStats(devices, uploaderInstance)
			}
		}
	}()
	
	// 处理退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	fmt.Println("模拟器运行中... 按 Ctrl+C 停止")
	fmt.Println()
	
	// 如果设置了运行时长，启动定时器
	if cfg.Duration > 0 {
		go func() {
			time.Sleep(cfg.Duration)
			fmt.Println("\n运行时长到达，正在停止...")
			sigChan <- syscall.SIGTERM
		}()
	}
	
	// 等待退出信号
	<-sigChan
	fmt.Println("\n正在停止模拟器...")
	
	// 停止统计
	close(stopStats)
	time.Sleep(1 * time.Second)
	
	// 停止所有设备
	fmt.Println("正在停止设备...")
	for _, dev := range devices {
		dev.Stop()
	}
	
	// 等待上传完成
	fmt.Println("等待数据上传完成...")
	time.Sleep(5 * time.Second)
	
	// 打印最终统计
	fmt.Println("\n========================================")
	fmt.Println("   最终统计")
	fmt.Println("========================================")
	printStats(devices, uploaderInstance)
	fmt.Println()
}

// printStats 打印统计信息
func printStats(devices []*device.Device, uploader *uploader.Uploader) {
	// 设备统计
	totalGenerated := int64(0)
	totalUploaded := int64(0)
	
	for _, dev := range devices {
		stats := dev.GetStats()
		totalGenerated += stats["data_generated"].(int64)
		totalUploaded += stats["data_uploaded"].(int64)
	}
	
	// 上传器统计
	uploadStats := uploader.GetStats()
	
	fmt.Println("----------------------------------------")
	fmt.Printf("时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("\n设备统计:\n")
	fmt.Printf("  总设备数: %d\n", len(devices))
	fmt.Printf("  总生成数据: %d 条\n", totalGenerated)
	fmt.Printf("  总上传数据: %d 条\n", totalUploaded)
	fmt.Printf("\n上传统计:\n")
	fmt.Printf("  成功上传: %v 条\n", uploadStats["total_success"])
	fmt.Printf("  失败次数: %v 次\n", uploadStats["total_failed"])
	fmt.Printf("  成功率: %v\n", uploadStats["success_rate"])
	fmt.Printf("  总数据量: %.2f MB\n", float64(uploadStats["total_bytes"].(int64))/1024/1024)
	fmt.Printf("  队列大小: %v\n", uploadStats["queue_size"])
	fmt.Println("----------------------------------------")
}
