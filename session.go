package main

import (
	"fmt"
	"math/rand"
	"sync"
)

// ==========================================
// 游戏配置
// ==========================================
const (
	InitialSteps = 15
	StepsReward  = 5
)

// GameStatus 游戏状态
type GameStatus string

const (
	StatusPlaying  GameStatus = "playing"
	StatusWonLevel GameStatus = "won_level" // 通过一关，等待下一关
	StatusGameOver GameStatus = "game_over" // 步数耗尽，游戏结束
)

// GameState 单个玩家的完整游戏状态
type GameState struct {
	PlayerName        string     `json:"player_name"`
	Level             int        `json:"level"`
	CurrentSteps      int        `json:"current_steps"`
	N                 int        `json:"n"`
	Heights           []int      `json:"heights"`     // 始终传输，前端决定是否显示
	TargetScore       int        `json:"-"`           // 服务端私有，不发送给客户端
	CurrentWater      int        `json:"current_water"`
	LastWater         int        `json:"-"`
	Status            GameStatus `json:"status"`
	Feedback          string     `json:"feedback"`
	Revealed          bool       `json:"revealed"`    // true 时前端显示真实高度
	TargetScoreReveal int        `json:"target_score_reveal"` // 只在结算时有意义
}

// SessionManager 管理所有玩家会话
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*GameState
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: make(map[string]*GameState)}
}

func (sm *SessionManager) Get(id string) (*GameState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	s, ok := sm.sessions[id]
	return s, ok
}

func (sm *SessionManager) Set(id string, s *GameState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[id] = s
}

func (sm *SessionManager) Delete(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, id)
}

// generateSessionID 生成随机会话 ID
func generateSessionID() string {
	return fmt.Sprintf("%016x%016x", rand.Int63(), rand.Int63())
}

// NewGameState 创建新游戏会话（从第 1 关开始）
func NewGameState(playerName string) *GameState {
	state := &GameState{
		PlayerName:   playerName,
		Level:        1,
		CurrentSteps: InitialSteps,
		N:            5,
	}
	return initLevel(state)
}

// initLevel 初始化当前关卡（随机高度 + 计算目标）
func initLevel(state *GameState) *GameState {
	// 难度递增：每关增加一根柱子，从 5 根开始，最高 10 根
	state.N = 4 + state.Level
	if state.N > 10 {
		state.N = 10
	}

	heights := make([]int, state.N)
	for i := 0; i < state.N; i++ {
		heights[i] = rand.Intn(6) + 1
	}

	targetScore := solveMaxWater(heights)

	rand.Shuffle(len(heights), func(i, j int) {
		heights[i], heights[j] = heights[j], heights[i]
	})

	state.Heights = heights
	state.TargetScore = targetScore
	state.LastWater = trap(heights)
	state.CurrentWater = state.LastWater
	state.Feedback = ""
	state.Revealed = false
	state.Status = StatusPlaying
	state.TargetScoreReveal = 0

	// 极少概率：初始排列已是最大值
	if state.CurrentWater == state.TargetScore {
		state.Status = StatusWonLevel
		state.Revealed = true
		state.TargetScoreReveal = state.TargetScore
	}

	return state
}
