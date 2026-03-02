package uploader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	
	"smartwatch-simulator/simulator/generator"
)

// Uploader 数据上传器
type Uploader struct {
	ServerURL   string
	APIEndpoint string
	Client      *http.Client
	Concurrency int
	
	// 统计信息
	TotalUploaded int64
	TotalSuccess  int64
	TotalFailed   int64
	TotalBytes    int64
	Mutex         sync.RWMutex
	
	// 上传队列
	UploadQueue chan []*generator.DeviceData
	Workers     []*Worker
}

// Worker 上传工作协程
type Worker struct {
	ID       int
	Uploader *Uploader
	StopChan chan struct{}
}

// NewUploader 创建上传器
func NewUploader(serverURL, apiEndpoint string, concurrency int) *Uploader {
	uploader := &Uploader{
		ServerURL:   serverURL,
		APIEndpoint: apiEndpoint,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        concurrency,
				MaxIdleConnsPerHost: concurrency,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		Concurrency: concurrency,
		UploadQueue: make(chan []*generator.DeviceData, 10000), // 缓冲队列
		Workers:     make([]*Worker, concurrency),
	}
	
	// 启动工作协程
	for i := 0; i < concurrency; i++ {
		worker := &Worker{
			ID:       i,
			Uploader: uploader,
			StopChan: make(chan struct{}),
		}
		uploader.Workers[i] = worker
		go worker.Start()
	}
	
	return uploader
}

// Start 启动工作协程
func (w *Worker) Start() {
	for {
		select {
		case <-w.StopChan:
			return
		case data := <-w.Uploader.UploadQueue:
			w.upload(data)
		}
	}
}

// Stop 停止工作协程
func (w *Worker) Stop() {
	close(w.StopChan)
}

// upload 上传数据
func (w *Worker) upload(data []*generator.DeviceData) {
	if len(data) == 0 {
		return
	}
	
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		w.Uploader.recordFailure()
		return
	}
	
	// 创建请求
	url := w.Uploader.ServerURL + w.Uploader.APIEndpoint
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		w.Uploader.recordFailure()
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	resp, err := w.Uploader.Client.Do(req)
	if err != nil {
		w.Uploader.recordFailure()
		return
	}
	defer resp.Body.Close()
	
	// 读取响应（避免连接泄漏）
	io.Copy(io.Discard, resp.Body)
	
	// 记录结果
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.Uploader.recordSuccess(len(data), len(jsonData))
	} else {
		w.Uploader.recordFailure()
	}
}

// Upload 上传数据（异步）
func (u *Uploader) Upload(data []*generator.DeviceData) {
	select {
	case u.UploadQueue <- data:
		// 成功加入队列
	default:
		// 队列满了，记录失败
		u.recordFailure()
	}
}

// recordSuccess 记录成功
func (u *Uploader) recordSuccess(count int, bytes int) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	u.TotalUploaded += int64(count)
	u.TotalSuccess += int64(count)
	u.TotalBytes += int64(bytes)
}

// recordFailure 记录失败
func (u *Uploader) recordFailure() {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	u.TotalFailed++
}

// GetStats 获取统计信息
func (u *Uploader) GetStats() map[string]interface{} {
	u.Mutex.RLock()
	defer u.Mutex.RUnlock()
	
	queueSize := len(u.UploadQueue)
	totalRequests := u.TotalSuccess + u.TotalFailed
	successRate := 0.0
	if totalRequests > 0 {
		successRate = float64(u.TotalSuccess) / float64(totalRequests) * 100
	}
	
	return map[string]interface{}{
		"total_uploaded": u.TotalUploaded,
		"total_success":  u.TotalSuccess,
		"total_failed":   u.TotalFailed,
		"success_rate":   fmt.Sprintf("%.2f%%", successRate),
		"total_bytes":    u.TotalBytes,
		"queue_size":     queueSize,
	}
}

// Stop 停止上传器
func (u *Uploader) Stop() {
	for _, worker := range u.Workers {
		worker.Stop()
	}
}
