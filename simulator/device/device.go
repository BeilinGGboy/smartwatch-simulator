package device

import (
	"math/rand"
	"sync"
	"time"
	
	"smartwatch-simulator/simulator/generator"
)

// UploadMode 上传模式
type UploadMode string

const (
	ModeRealTime UploadMode = "realtime" // 实时上传
	ModeBatch    UploadMode = "batch"    // 批量上传
	ModeTimer    UploadMode = "timer"    // 定时上传
)

// Device 模拟设备
type Device struct {
	ID          string
	UserID      int
	Mode        UploadMode
	Generator   *generator.DataGenerator
	DataBuffer  []*generator.DeviceData
	BufferMutex sync.Mutex
	
	// 设备状态
	BaseHeartRate int
	DailyStepTarget int
	LastUploadTime time.Time
	
	// 统计信息
	DataGenerated int64
	DataUploaded  int64
	
	// 控制通道
	StopChan     chan struct{}
	UploadChan   chan []*generator.DeviceData
}

// NewDevice 创建新设备
func NewDevice(deviceID string, userID int, mode UploadMode) *Device {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	return &Device{
		ID:            deviceID,
		UserID:        userID,
		Mode:          mode,
		Generator:     generator.NewDataGenerator(),
		DataBuffer:    make([]*generator.DeviceData, 0),
		BaseHeartRate: 70 + rand.Intn(20), // 70-90
		DailyStepTarget: 5000 + rand.Intn(10000), // 5000-15000
		StopChan:      make(chan struct{}),
		UploadChan:    make(chan []*generator.DeviceData, 100),
	}
}

// Start 启动设备
func (d *Device) Start(dataInterval time.Duration, uploadInterval time.Duration, batchSize int) {
	// 根据模式启动不同的数据生成和上传逻辑
	switch d.Mode {
	case ModeRealTime:
		go d.realTimeMode(dataInterval)
	case ModeBatch:
		go d.batchMode(dataInterval, uploadInterval, batchSize)
	case ModeTimer:
		go d.timerMode(dataInterval, uploadInterval)
	}
}

// Stop 停止设备
func (d *Device) Stop() {
	close(d.StopChan)
	// 上传剩余数据
	d.flushBuffer()
}

// realTimeMode 实时上传模式
func (d *Device) realTimeMode(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-d.StopChan:
			return
		case <-ticker.C:
			// 生成数据并立即上传
			dataType := d.getRandomDataType()
			data := d.Generator.GenerateDeviceData(d.ID, d.UserID, dataType)
			d.DataGenerated++
			
			// 立即上传
			d.UploadChan <- []*generator.DeviceData{data}
			d.DataUploaded++
		}
	}
}

// batchMode 批量上传模式
func (d *Device) batchMode(dataInterval, uploadInterval time.Duration, batchSize int) {
	dataTicker := time.NewTicker(dataInterval)
	uploadTicker := time.NewTicker(uploadInterval)
	defer dataTicker.Stop()
	defer uploadTicker.Stop()
	
	for {
		select {
		case <-d.StopChan:
			return
		case <-dataTicker.C:
			// 生成数据并缓存
			dataType := d.getRandomDataType()
			data := d.Generator.GenerateDeviceData(d.ID, d.UserID, dataType)
			d.DataGenerated++
			
			d.BufferMutex.Lock()
			d.DataBuffer = append(d.DataBuffer, data)
			d.BufferMutex.Unlock()
			
		case <-uploadTicker.C:
			// 批量上传
			d.flushBuffer()
		}
	}
}

// timerMode 定时上传模式
func (d *Device) timerMode(dataInterval, uploadInterval time.Duration) {
	dataTicker := time.NewTicker(dataInterval)
	uploadTicker := time.NewTicker(uploadInterval)
	defer dataTicker.Stop()
	defer uploadTicker.Stop()
	
	for {
		select {
		case <-d.StopChan:
			return
		case <-dataTicker.C:
			// 生成数据并缓存
			dataType := d.getRandomDataType()
			data := d.Generator.GenerateDeviceData(d.ID, d.UserID, dataType)
			d.DataGenerated++
			
			d.BufferMutex.Lock()
			d.DataBuffer = append(d.DataBuffer, data)
			d.BufferMutex.Unlock()
			
		case <-uploadTicker.C:
			// 定时上传（上传所有缓存数据）
			d.flushBuffer()
		}
	}
}

// flushBuffer 上传缓冲区数据
func (d *Device) flushBuffer() {
	d.BufferMutex.Lock()
	if len(d.DataBuffer) == 0 {
		d.BufferMutex.Unlock()
		return
	}
	
	// 复制数据
	dataToUpload := make([]*generator.DeviceData, len(d.DataBuffer))
	copy(dataToUpload, d.DataBuffer)
	d.DataBuffer = d.DataBuffer[:0] // 清空缓冲区
	d.BufferMutex.Unlock()
	
	// 发送到上传通道
	if len(dataToUpload) > 0 {
		d.UploadChan <- dataToUpload
		d.DataUploaded += int64(len(dataToUpload))
	}
}

// getRandomDataType 获取随机数据类型
func (d *Device) getRandomDataType() string {
	types := []string{"heartrate", "steps", "sleep", "sport"}
	// 使用设备的生成器的随机数，避免每次创建新的 rand
	// 注意：这里需要访问 Generator 的 rand，但由于它是私有的，我们使用简单的方法
	// 实际可以使用 time 和 device ID 生成一个稳定的随机数
	seed := time.Now().UnixNano() + int64(len(d.ID))
	r := rand.New(rand.NewSource(seed))
	return types[r.Intn(len(types))]
}

// GetStats 获取统计信息
func (d *Device) GetStats() map[string]interface{} {
	d.BufferMutex.Lock()
	bufferSize := len(d.DataBuffer)
	d.BufferMutex.Unlock()
	
	return map[string]interface{}{
		"device_id":      d.ID,
		"user_id":        d.UserID,
		"mode":           string(d.Mode),
		"data_generated": d.DataGenerated,
		"data_uploaded":  d.DataUploaded,
		"buffer_size":    bufferSize,
	}
}
