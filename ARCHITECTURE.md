# 智能手表数据模拟器 - 架构说明

## 📁 整体项目结构

```
smartwatch-simulator/
├── main.go                    # 主程序入口（控制中心）
├── test_server.go             # 测试服务器（接收数据）
├── go.mod                     # Go 模块依赖管理
├── go.sum                     # 依赖校验文件
│
├── simulator/                 # 模拟器核心模块
│   ├── config/               # 配置管理模块
│   │   └── config.go        # 配置结构定义
│   │
│   ├── generator/            # 数据生成模块
│   │   └── data_generator.go # 数据生成器（生成各种健康数据）
│   │
│   ├── device/               # 设备模拟模块
│   │   └── device.go        # 设备模拟器（模拟单个手表）
│   │
│   └── uploader/             # 数据上传模块
│       └── uploader.go      # 批量上传器（并发上传数据）
│
└── utils/                     # 工具函数
    └── uuid.go              # UUID 生成工具
```

## 🏗️ 架构流程图

```
┌─────────────────────────────────────────────────────────────┐
│                        main.go                               │
│                    (主程序控制中心)                           │
│                                                              │
│  1. 加载配置 (config)                                        │
│  2. 创建上传器 (uploader)                                    │
│  3. 创建 N 个设备 (device)                                   │
│  4. 启动所有设备                                             │
│  5. 监控统计信息                                             │
└────────────┬────────────────────────────────────────────────┘
             │
             ├─────────────────┬─────────────────┐
             │                 │                 │
             ▼                 ▼                 ▼
    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │   Device 1   │  │   Device 2   │  │  Device N    │
    │  (设备1)      │  │  (设备2)      │  │  (设备N)      │
    └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
           │                  │                 │
           │  生成数据        │  生成数据        │  生成数据
           │  (generator)     │  (generator)     │  (generator)
           │                  │                 │
           └──────────┬───────┴─────────────────┘
                      │
                      │  数据上传通道
                      ▼
           ┌──────────────────────┐
           │     Uploader         │
           │   (批量上传器)        │
           │                      │
           │  - 接收设备数据       │
           │  - 批量打包          │
           │  - 并发上传到服务器   │
           └──────────┬───────────┘
                      │
                      │  HTTP POST
                      ▼
           ┌──────────────────────┐
           │   test_server.go     │
           │   (测试服务器)       │
           └──────────────────────┘
```

## 📄 各文件详细说明

### 1. main.go - 主程序入口

**作用：** 整个模拟器的控制中心，负责协调各个模块

**主要功能：**
- 加载配置参数
- 创建上传器实例
- 创建和管理所有设备
- 启动设备（分批启动，避免同时启动）
- 监控和统计
- 优雅退出（Ctrl+C）

**关键代码结构：**
```go
func main() {
    // 1. 加载配置
    cfg := config.DefaultConfig()
    
    // 2. 创建上传器
    uploaderInstance := uploader.NewUploader(...)
    
    // 3. 创建设备
    devices := make([]*device.Device, 0, cfg.DeviceCount)
    
    // 4. 启动设备（分批启动）
    for i, dev := range devices {
        go func() {
            time.Sleep(time.Duration(idx) * cfg.DeviceStartDelay)
            dev.Start(...)
        }()
    }
    
    // 5. 连接设备上传通道到上传器
    go func() {
        for data := range dev.UploadChan {
            uploaderInstance.Upload(data)
        }
    }()
    
    // 6. 监控统计
    go printStats(...)
    
    // 7. 等待退出信号
    <-sigChan
}
```

---

### 2. simulator/config/config.go - 配置管理

**作用：** 定义所有可配置的参数

**主要结构：**
```go
type Config struct {
    // 设备配置
    DeviceCount      int           // 设备数量（默认10万）
    DeviceStartDelay time.Duration // 设备启动延迟
    
    // 数据生成配置
    DataInterval     time.Duration // 数据生成间隔（1秒）
    BatchSize        int           // 批量上传大小（100条）
    UploadInterval   time.Duration // 上传间隔（5秒）
    
    // 服务器配置
    ServerURL        string        // 服务器地址
    APIEndpoint      string        // API端点
    ConcurrentUpload int           // 并发上传数（100）
    
    // 上传模式比例
    RealTimeRatio    float64       // 实时上传比例（20%）
    BatchRatio       float64       // 批量上传比例（70%）
    TimerRatio       float64       // 定时上传比例（10%）
    
    // 运行时间
    Duration         time.Duration // 运行时长（0=持续运行）
}
```

**关键函数：**
- `DefaultConfig()` - 返回默认配置

---

### 3. simulator/generator/data_generator.go - 数据生成器

**作用：** 生成符合真实规律的模拟数据

**主要数据结构：**
```go
// 心率数据
type HeartRateData struct {
    BPM       int
    Timestamp time.Time
}

// 步数数据
type StepsData struct {
    Steps     int
    Distance  float64
    Calories  int
    Timestamp time.Time
}

// 睡眠数据
type SleepData struct {
    SleepStart  time.Time
    SleepEnd    time.Time
    Duration    int
    DeepSleep   int
    LightSleep  int
    SleepQuality int
}

// 运动数据
type SportData struct {
    SportType   string
    StartTime   time.Time
    EndTime     time.Time
    Duration    int
    Distance    float64
    Calories    int
    AvgHeartRate int
    MaxHeartRate int
}

// 设备数据包（包含以上任意一种数据）
type DeviceData struct {
    DeviceID   string
    UserID     int
    Timestamp  time.Time
    HeartRate  *HeartRateData
    Steps      *StepsData
    Sleep      *SleepData
    Sport      *SportData
}
```

**关键函数：**
- `GenerateHeartRate(baseRate, mode)` - 生成心率数据
  - mode: "normal" / "exercise" / "sleep"
- `GenerateSteps(hour, dailyTarget)` - 生成步数数据
  - 根据时间分布（早上少，下午多）
- `GenerateSleep()` - 生成睡眠数据
  - 模拟夜间睡眠周期
- `GenerateSport(sportType)` - 生成运动数据
  - sportType: "running" / "cycling" / "walking" / "swimming"
- `GenerateDeviceData(deviceID, userID, dataType)` - 生成完整设备数据包

**数据生成策略：**
- 心率：遵循生理规律（运动时高，睡眠时低）
- 步数：符合时间分布（白天多，晚上少）
- 睡眠：模拟真实睡眠周期（深睡/浅睡交替）
- 运动：根据运动类型计算距离和卡路里

---

### 4. simulator/device/device.go - 设备模拟器

**作用：** 模拟单个智能手表设备的行为

**主要结构：**
```go
type Device struct {
    ID          string              // 设备ID
    UserID      int                 // 用户ID
    Mode        UploadMode          // 上传模式
    Generator   *DataGenerator      // 数据生成器
    DataBuffer  []*DeviceData       // 数据缓冲区
    BufferMutex sync.Mutex          // 缓冲区锁
    
    // 设备状态
    BaseHeartRate   int             // 基础心率
    DailyStepTarget  int             // 每日步数目标
    
    // 统计信息
    DataGenerated int64              // 生成的数据量
    DataUploaded  int64              // 上传的数据量
    
    // 控制通道
    StopChan   chan struct{}        // 停止信号
    UploadChan chan []*DeviceData   // 上传通道
}
```

**上传模式：**
```go
const (
    ModeRealTime UploadMode = "realtime"  // 实时上传
    ModeBatch    UploadMode = "batch"     // 批量上传
    ModeTimer    UploadMode = "timer"     // 定时上传
)
```

**关键函数：**
- `NewDevice(deviceID, userID, mode)` - 创建新设备
- `Start(dataInterval, uploadInterval, batchSize)` - 启动设备
  - 根据模式启动不同的协程
- `realTimeMode()` - 实时上传模式
  - 生成数据后立即上传
- `batchMode()` - 批量上传模式
  - 生成数据后缓存，定时批量上传
- `timerMode()` - 定时上传模式
  - 定时上传所有缓存数据
- `flushBuffer()` - 上传缓冲区数据
- `Stop()` - 停止设备
- `GetStats()` - 获取统计信息

**工作流程：**
```
设备启动
  ↓
根据模式启动协程
  ↓
定时生成数据 (ticker)
  ↓
根据模式处理数据:
  - 实时模式: 立即发送到 UploadChan
  - 批量模式: 存入 DataBuffer，定时批量发送
  - 定时模式: 存入 DataBuffer，定时全部发送
  ↓
数据发送到 UploadChan
  ↓
主程序接收并转发给 Uploader
```

---

### 5. simulator/uploader/uploader.go - 批量上传器

**作用：** 接收设备数据并批量上传到服务器

**主要结构：**
```go
type Uploader struct {
    ServerURL   string                    // 服务器地址
    APIEndpoint string                    // API端点
    Client      *http.Client              // HTTP客户端
    Concurrency int                       // 并发数
    
    // 统计信息
    TotalUploaded int64                   // 总上传数
    TotalSuccess  int64                   // 成功数
    TotalFailed   int64                   // 失败数
    TotalBytes    int64                   // 总字节数
    Mutex         sync.RWMutex            // 统计锁
    
    // 上传队列
    UploadQueue chan []*DeviceData        // 上传队列
    Workers     []*Worker                 // 工作协程
}

type Worker struct {
    ID       int
    Uploader *Uploader
    StopChan chan struct{}
}
```

**关键函数：**
- `NewUploader(serverURL, apiEndpoint, concurrency)` - 创建上传器
  - 创建 HTTP 客户端（连接池）
  - 启动 N 个工作协程
- `Upload(data)` - 上传数据（异步）
  - 将数据放入队列
- `upload(data)` - 实际上传逻辑（Worker 执行）
  - 序列化 JSON
  - 发送 HTTP POST 请求
  - 记录成功/失败
- `GetStats()` - 获取统计信息
- `Stop()` - 停止上传器

**工作流程：**
```
设备数据 → UploadChan
  ↓
main.go 接收并调用 Uploader.Upload()
  ↓
数据放入 UploadQueue（缓冲队列）
  ↓
Worker 协程从队列取数据
  ↓
序列化为 JSON
  ↓
HTTP POST 到服务器
  ↓
记录统计信息（成功/失败）
```

**并发控制：**
- 使用 Worker 池模式
- 多个 Worker 并发处理上传队列
- 使用 HTTP 连接池复用连接

---

### 6. utils/uuid.go - 工具函数

**作用：** 提供 UUID 生成等工具函数

**关键函数：**
- `GenerateDeviceID()` - 生成设备ID
  - 格式: "device-xxxxxxxx"
- `GenerateUserID()` - 生成用户ID
  - 范围: 1-1000000

---

### 7. test_server.go - 测试服务器

**作用：** 简单的 HTTP 服务器，用于接收和统计模拟器发送的数据

**主要功能：**
- 接收批量数据 POST 请求
- 统计接收的数据量
- 提供统计查询接口
- 健康检查接口

**接口：**
- `POST /api/v1/data/batch` - 接收批量数据
- `GET /stats` - 查看统计信息
- `GET /health` - 健康检查

---

## 🔄 完整数据流

```
1. main.go 启动
   ↓
2. 加载配置（设备数量、上传模式等）
   ↓
3. 创建 Uploader（上传器）
   ↓
4. 创建 N 个 Device（设备）
   ↓
5. 分批启动设备（每10ms启动一个）
   ↓
6. 每个设备启动后：
   - 启动数据生成协程（定时器）
   - 启动上传协程（根据模式）
   ↓
7. 数据生成流程：
   Device → Generator.GenerateDeviceData()
   ↓
   生成心率/步数/睡眠/运动数据
   ↓
   包装成 DeviceData
   ↓
8. 数据上传流程（根据模式）：
   
   实时模式：
   DeviceData → UploadChan → Uploader.Upload() → UploadQueue → Worker → HTTP POST
   
   批量模式：
   DeviceData → DataBuffer → 定时 flushBuffer() → UploadChan → ...
   
   定时模式：
   DeviceData → DataBuffer → 定时 flushBuffer() → UploadChan → ...
   ↓
9. Uploader 处理：
   - 从队列取数据
   - 批量打包（JSON）
   - 并发 HTTP POST
   - 记录统计
   ↓
10. 服务器接收数据
    ↓
11. 主程序每5秒打印统计信息
```

## 🎯 关键设计点

### 1. **并发控制**
- 设备分批启动（避免同时启动压力）
- Worker 池模式（控制并发上传数）
- 使用 channel 进行协程间通信

### 2. **内存优化**
- 数据流式处理（不一次性生成所有数据）
- 批量上传（减少内存占用）
- 及时释放缓冲区

### 3. **性能优化**
- HTTP 连接池复用
- 批量请求（减少网络开销）
- 异步处理（不阻塞数据生成）

### 4. **可扩展性**
- 配置化设计（易于调整参数）
- 模块化设计（易于添加新功能）
- 统计监控（便于性能分析）

## 📊 运行时的协程结构

```
main goroutine
  ├── 设备1 goroutine (数据生成)
  │   └── 设备1 goroutine (数据上传)
  ├── 设备2 goroutine (数据生成)
  │   └── 设备2 goroutine (数据上传)
  ├── ...
  ├── 设备N goroutine (数据生成)
  │   └── 设备N goroutine (数据上传)
  ├── Uploader Worker 1
  ├── Uploader Worker 2
  ├── ...
  ├── Uploader Worker N
  └── 统计监控 goroutine
```

## 🔍 如何理解代码

1. **从 main.go 开始**：看整体流程
2. **理解 Device**：看单个设备如何工作
3. **理解 Generator**：看数据如何生成
4. **理解 Uploader**：看数据如何上传
5. **理解 Config**：看如何配置参数

希望这个文档能帮助你理解整个模拟器的结构！
