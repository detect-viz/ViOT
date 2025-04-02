package global

import (
	"fmt"
	"sync"
	"time"
	"viot-scanner/models"
)

// Scanner 掃描器
type Scanner struct {
	config        *models.Config
	ipRanges      []models.IPRange
	lastScans     map[string]time.Time
	roomSchedules map[string]*RoomScanSchedule
	mu            sync.Mutex
	semaphore     chan struct{}
}

// NewScanner 創建新的掃描器
func NewScanner(config *models.Config) *Scanner {
	return &Scanner{
		config:        config,
		ipRanges:      make([]models.IPRange, 0),
		lastScans:     make(map[string]time.Time),
		roomSchedules: make(map[string]*RoomScanSchedule),
		semaphore:     make(chan struct{}, config.Scanner.MaxConcurrent),
	}
}

// getLastScanTime 獲取房間最後掃描時間
func (s *Scanner) getLastScanTime(room string) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	if lastTime, exists := s.lastScans[room]; exists {
		return lastTime
	}
	return time.Time{}
}

// updateLastScanTime 更新房間最後掃描時間
func (s *Scanner) updateLastScanTime(room string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastScans[room] = time.Now()
}

// UpdateRoomInterval 更新指定房間的掃描間隔
func (s *Scanner) UpdateRoomInterval(room string, interval int, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.roomSchedules[room]
	if !exists {
		return fmt.Errorf("房間 %s 不存在", room)
	}

	return schedule.AddTemporaryTask(room, interval, duration)
}

// shouldScan 檢查是否應該掃描指定房間
func (s *Scanner) shouldScan(room string, interval int) bool {
	lastScanTime := s.getLastScanTime(room)
	return time.Since(lastScanTime) >= time.Duration(interval)*time.Second
}

// GetRoomSchedule 獲取房間掃描排程
func (s *Scanner) GetRoomSchedule(room string) *RoomScanSchedule {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.roomSchedules[room]
	if !exists {
		schedule = NewRoomScanSchedule(s.config.Scanner.ScanInterval)
		s.roomSchedules[room] = schedule
	}
	return schedule
}

// UpdateRoomSchedule 更新房間掃描排程
func (s *Scanner) UpdateRoomSchedule(room string, interval int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roomSchedules[room]; !exists {
		s.roomSchedules[room] = NewRoomScanSchedule(interval)
	}
}

// CleanupExpiredSchedules 清理過期的掃描排程
func (s *Scanner) CleanupExpiredSchedules() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, schedule := range s.roomSchedules {
		schedule.CleanupExpiredTasks()
	}
}

// GetNextScanTime 獲取下一次掃描時間
func (s *Scanner) GetNextScanTime(room string) time.Time {
	lastScan := s.getLastScanTime(room)
	interval := s.GetScanInterval(room)
	return lastScan.Add(time.Duration(interval) * time.Second)
}

// IsScanDue 檢查是否應該進行掃描
func (s *Scanner) IsScanDue(room string) bool {
	schedule := s.GetRoomSchedule(room)
	interval := schedule.GetRoomInterval(room)
	return s.shouldScan(room, interval)
}

// UpdateLastScan 更新最後掃描時間
func (s *Scanner) UpdateLastScan(room string) {
	s.updateLastScanTime(room)
}

// GetScanInterval 獲取掃描間隔
func (s *Scanner) GetScanInterval(room string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.roomSchedules[room]
	if !exists {
		return s.config.Scanner.ScanInterval
	}

	return schedule.GetRoomInterval(room)
}

// TemporaryScanTask 臨時掃描任務
type TemporaryScanTask struct {
	Room      string    `json:"room"`
	Interval  int       `json:"interval"`   // 掃描間隔（秒）
	StartTime time.Time `json:"start_time"` // 開始時間
	EndTime   time.Time `json:"end_time"`   // 結束時間
	IsActive  bool      `json:"is_active"`  // 是否啟用
	CreatedAt time.Time `json:"created_at"` // 創建時間
	UpdatedAt time.Time `json:"updated_at"` // 更新時間
}

// RoomScanSchedule 房間掃描排程
type RoomScanSchedule struct {
	DefaultInterval int                 `json:"default_interval"` // 預設掃描間隔（秒）
	TemporaryTasks  []TemporaryScanTask `json:"temporary_tasks"`  // 臨時任務列表
	mu              sync.RWMutex        // 讀寫鎖
}

// NewRoomScanSchedule 創建新的房間掃描排程
func NewRoomScanSchedule(defaultInterval int) *RoomScanSchedule {
	return &RoomScanSchedule{
		DefaultInterval: defaultInterval,
		TemporaryTasks:  make([]TemporaryScanTask, 0),
	}
}

// AddTemporaryTask 添加臨時掃描任務
func (s *RoomScanSchedule) AddTemporaryTask(room string, interval int, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	task := TemporaryScanTask{
		Room:      room,
		Interval:  interval,
		StartTime: now,
		EndTime:   now.Add(duration),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 檢查是否已存在相同房間的臨時任務
	for i, t := range s.TemporaryTasks {
		if t.Room == room && t.IsActive {
			// 更新現有任務
			s.TemporaryTasks[i] = task
			return nil
		}
	}

	s.TemporaryTasks = append(s.TemporaryTasks, task)
	return nil
}

// GetRoomInterval 獲取房間的當前掃描間隔
func (s *RoomScanSchedule) GetRoomInterval(room string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	for _, task := range s.TemporaryTasks {
		if task.Room == room && task.IsActive {
			if now.After(task.EndTime) {
				// 任務已過期，標記為非活動
				task.IsActive = false
				continue
			}
			return task.Interval
		}
	}

	return s.DefaultInterval
}

// CleanupExpiredTasks 清理過期的臨時任務
func (s *RoomScanSchedule) CleanupExpiredTasks() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	activeTasks := make([]TemporaryScanTask, 0)
	for _, task := range s.TemporaryTasks {
		if task.IsActive && now.Before(task.EndTime) {
			activeTasks = append(activeTasks, task)
		}
	}
	s.TemporaryTasks = activeTasks
}

// AddTemporaryScanTask 添加臨時掃描任務
func (s *Scanner) AddTemporaryScanTask(roomID string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.roomSchedules[roomID]
	if !exists {
		schedule = NewRoomScanSchedule(s.config.Scanner.ScanInterval)
		s.roomSchedules[roomID] = schedule
	}

	schedule.AddTemporaryTask(roomID, s.config.Scanner.ScanInterval, duration)
}

// GetRoomScanInterval 獲取房間掃描間隔
func (s *Scanner) GetRoomScanInterval(roomID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.roomSchedules[roomID]
	if !exists {
		return s.config.Scanner.ScanInterval
	}

	return schedule.GetRoomInterval(roomID)
}
