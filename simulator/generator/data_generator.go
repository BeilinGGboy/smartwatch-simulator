package generator

import (
	"math/rand"
	"time"
)

// DataGenerator 数据生成器
type DataGenerator struct {
	rand *rand.Rand
}

// NewDataGenerator 创建数据生成器
func NewDataGenerator() *DataGenerator {
	return &DataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// HeartRateData 心率数据
type HeartRateData struct {
	BPM       int       `json:"bpm"`
	Timestamp time.Time `json:"timestamp"`
}

// StepsData 步数数据
type StepsData struct {
	Steps     int       `json:"steps"`
	Distance  float64   `json:"distance"`  // 米
	Calories  int       `json:"calories"`
	Timestamp time.Time `json:"timestamp"`
}

// SleepData 睡眠数据
type SleepData struct {
	SleepStart  time.Time `json:"sleep_start"`
	SleepEnd    time.Time `json:"sleep_end"`
	Duration    int       `json:"duration"`     // 分钟
	DeepSleep   int       `json:"deep_sleep"`   // 分钟
	LightSleep  int       `json:"light_sleep"`  // 分钟
	SleepQuality int      `json:"sleep_quality"` // 1-100
}

// SportData 运动数据
type SportData struct {
	SportType   string    `json:"sport_type"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    int       `json:"duration"`     // 秒
	Distance    float64   `json:"distance"`     // 米
	Calories    int       `json:"calories"`
	AvgHeartRate int      `json:"avg_heart_rate"`
	MaxHeartRate int      `json:"max_heart_rate"`
}

// DeviceData 设备数据包
type DeviceData struct {
	DeviceID   string          `json:"device_id"`
	UserID     int             `json:"user_id"`
	Timestamp  time.Time       `json:"timestamp"`
	HeartRate  *HeartRateData  `json:"heart_rate,omitempty"`
	Steps      *StepsData      `json:"steps,omitempty"`
	Sleep      *SleepData      `json:"sleep,omitempty"`
	Sport      *SportData      `json:"sport,omitempty"`
}

// GenerateHeartRate 生成心率数据
// baseRate: 基础心率（60-100），mode: 模式（normal/exercise/sleep）
func (g *DataGenerator) GenerateHeartRate(baseRate int, mode string) *HeartRateData {
	var bpm int
	
	switch mode {
	case "exercise":
		// 运动时心率：120-180
		bpm = 120 + g.rand.Intn(60)
	case "sleep":
		// 睡眠时心率：50-70
		bpm = 50 + g.rand.Intn(20)
	default:
		// 正常心率：在基础心率附近波动 ±10
		bpm = baseRate + g.rand.Intn(21) - 10
		if bpm < 60 {
			bpm = 60
		}
		if bpm > 100 {
			bpm = 100
		}
	}
	
	return &HeartRateData{
		BPM:       bpm,
		Timestamp: time.Now(),
	}
}

// GenerateSteps 生成步数数据
// 模拟一天的步数分布：早上少，中午多，晚上少
func (g *DataGenerator) GenerateSteps(hour int, dailyTarget int) *StepsData {
	// 根据时间计算步数比例
	var ratio float64
	if hour >= 6 && hour < 9 {
		ratio = 0.1 // 早上
	} else if hour >= 9 && hour < 12 {
		ratio = 0.15 // 上午
	} else if hour >= 12 && hour < 14 {
		ratio = 0.1 // 中午
	} else if hour >= 14 && hour < 18 {
		ratio = 0.25 // 下午
	} else if hour >= 18 && hour < 22 {
		ratio = 0.3 // 晚上
	} else {
		ratio = 0.1 // 深夜
	}
	
	// 添加随机波动
	ratio += (g.rand.Float64() - 0.5) * 0.2
	if ratio < 0 {
		ratio = 0.05
	}
	
	steps := int(float64(dailyTarget) * ratio / 24) // 每小时步数
	steps += g.rand.Intn(100) - 50 // 随机波动
	
	if steps < 0 {
		steps = 0
	}
	
	// 计算距离（假设每步0.7米）
	distance := float64(steps) * 0.7
	
	// 计算卡路里（假设每1000步消耗30卡路里）
	calories := steps * 30 / 1000
	
	return &StepsData{
		Steps:     steps,
		Distance:  distance,
		Calories:  calories,
		Timestamp: time.Now(),
	}
}

// GenerateSleep 生成睡眠数据
// 模拟夜间睡眠
func (g *DataGenerator) GenerateSleep() *SleepData {
	now := time.Now()
	
	// 睡眠开始时间：晚上 22:00 - 23:30
	sleepHour := 22 + g.rand.Intn(2)
	sleepMinute := g.rand.Intn(60)
	sleepStart := time.Date(now.Year(), now.Month(), now.Day()-1, sleepHour, sleepMinute, 0, 0, now.Location())
	
	// 睡眠时长：6-9小时
	duration := 360 + g.rand.Intn(180) // 6-9小时（分钟）
	sleepEnd := sleepStart.Add(time.Duration(duration) * time.Minute)
	
	// 深睡：占总睡眠的 20-30%
	deepSleep := duration * (20 + g.rand.Intn(11)) / 100
	// 浅睡：剩余部分
	lightSleep := duration - deepSleep
	
	// 睡眠质量：60-100
	sleepQuality := 60 + g.rand.Intn(41)
	
	return &SleepData{
		SleepStart:  sleepStart,
		SleepEnd:    sleepEnd,
		Duration:    duration,
		DeepSleep:   deepSleep,
		LightSleep:  lightSleep,
		SleepQuality: sleepQuality,
	}
}

// GenerateSport 生成运动数据
func (g *DataGenerator) GenerateSport(sportType string) *SportData {
	sportTypes := []string{"running", "cycling", "walking", "swimming"}
	if sportType == "" {
		sportType = sportTypes[g.rand.Intn(len(sportTypes))]
	}
	
	now := time.Now()
	// 运动时长：30分钟 - 2小时
	duration := 1800 + g.rand.Intn(6300) // 30分钟 - 2小时（秒）
	startTime := now.Add(-time.Duration(duration) * time.Second)
	endTime := now
	
	// 根据运动类型计算距离和卡路里
	var distance float64
	var calories int
	
	switch sportType {
	case "running":
		// 跑步：配速 5-8分钟/公里
		speed := 5.0 + g.rand.Float64()*3.0 // 分钟/公里
		distance = float64(duration) / 60.0 / speed * 1000 // 米
		calories = int(distance / 1000.0 * 60) // 每公里60卡路里
	case "cycling":
		// 骑行：速度 15-25 km/h
		speed := 15.0 + g.rand.Float64()*10.0 // km/h
		distance = speed * float64(duration) / 3600.0 * 1000 // 米
		calories = int(distance / 1000.0 * 30) // 每公里30卡路里
	case "walking":
		// 步行：速度 4-6 km/h
		speed := 4.0 + g.rand.Float64()*2.0 // km/h
		distance = speed * float64(duration) / 3600.0 * 1000 // 米
		calories = int(distance / 1000.0 * 50) // 每公里50卡路里
	default:
		distance = float64(g.rand.Intn(10000)) // 0-10km
		calories = g.rand.Intn(500) // 0-500卡路里
	}
	
	// 运动时心率：120-180
	avgHeartRate := 120 + g.rand.Intn(60)
	maxHeartRate := avgHeartRate + g.rand.Intn(20)
	
	return &SportData{
		SportType:   sportType,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    duration,
		Distance:    distance,
		Calories:    calories,
		AvgHeartRate: avgHeartRate,
		MaxHeartRate: maxHeartRate,
	}
}

// GenerateDeviceData 生成设备数据包
func (g *DataGenerator) GenerateDeviceData(deviceID string, userID int, dataType string) *DeviceData {
	data := &DeviceData{
		DeviceID:  deviceID,
		UserID:    userID,
		Timestamp: time.Now(),
	}
	
	// 根据类型生成对应数据
	switch dataType {
	case "heartrate":
		baseRate := 70 + g.rand.Intn(20) // 基础心率 70-90
		mode := []string{"normal", "exercise", "sleep"}[g.rand.Intn(3)]
		data.HeartRate = g.GenerateHeartRate(baseRate, mode)
	case "steps":
		hour := time.Now().Hour()
		dailyTarget := 5000 + g.rand.Intn(10000) // 每日目标 5000-15000步
		data.Steps = g.GenerateSteps(hour, dailyTarget)
	case "sleep":
		data.Sleep = g.GenerateSleep()
	case "sport":
		data.Sport = g.GenerateSport("")
	default:
		// 随机生成一种类型
		types := []string{"heartrate", "steps", "sleep", "sport"}
		randomType := types[g.rand.Intn(len(types))]
		return g.GenerateDeviceData(deviceID, userID, randomType)
	}
	
	return data
}
