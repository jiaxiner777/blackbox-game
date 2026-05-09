package store

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

const leaderboardFile = "leaderboard.json"
const maxLeaderboardSize = 5 // TOP 5

// LeaderboardEntry 排行榜单条记录
type LeaderboardEntry struct {
	Name     string `json:"name"`
	MaxLevel int    `json:"max_level"`
	Date     string `json:"date"`
}

// Leaderboard 排行榜，线程安全
type Leaderboard struct {
	mu      sync.RWMutex
	Entries []LeaderboardEntry `json:"entries"`
}

// LoadLeaderboard 从文件加载排行榜
func LoadLeaderboard() *Leaderboard {
	lb := &Leaderboard{}
	data, err := os.ReadFile(leaderboardFile)
	if err != nil {
		return lb
	}
	_ = json.Unmarshal(data, lb)
	return lb
}

func (lb *Leaderboard) save() {
	data, err := json.MarshalIndent(lb, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(leaderboardFile, data, 0644)
}

// Submit 提交成绩（levelsCompleted = 实际通过的关数）
func (lb *Leaderboard) Submit(name string, levelsCompleted int) {
	if levelsCompleted <= 0 {
		return
	}
	lb.mu.Lock()
	defer lb.mu.Unlock()

	found := false
	for i, e := range lb.Entries {
		if e.Name == name {
			found = true
			if levelsCompleted > e.MaxLevel {
				lb.Entries[i].MaxLevel = levelsCompleted
				lb.Entries[i].Date = time.Now().Format("2006-01-02")
			}
			break
		}
	}
	if !found {
		lb.Entries = append(lb.Entries, LeaderboardEntry{
			Name:     name,
			MaxLevel: levelsCompleted,
			Date:     time.Now().Format("2006-01-02"),
		})
	}

	sort.Slice(lb.Entries, func(i, j int) bool {
		return lb.Entries[i].MaxLevel > lb.Entries[j].MaxLevel
	})

	if len(lb.Entries) > maxLeaderboardSize {
		lb.Entries = lb.Entries[:maxLeaderboardSize]
	}

	lb.save()
}

// Top5 返回 TOP 5 排名（线程安全）
func (lb *Leaderboard) Top5() []LeaderboardEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]LeaderboardEntry, len(lb.Entries))
	copy(result, lb.Entries)
	return result
}
