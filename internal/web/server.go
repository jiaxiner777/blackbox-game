package web

import (
	"blackbox-game/internal/game"
	"blackbox-game/internal/store"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type SessionManager = game.SessionManager
type GameState = game.GameState
type Leaderboard = store.Leaderboard

const (
	StatusPlaying  = game.StatusPlaying
	StatusWonLevel = game.StatusWonLevel
	StatusGameOver = game.StatusGameOver
	StepsReward    = game.StepsReward
)

var (
	NewSessionManager = game.NewSessionManager
	NewGameState      = game.NewGameState
	generateSessionID = game.GenerateSessionID
	trap              = game.Trap
	initLevel         = game.InitLevel
)

// Server HTTP 服务器，持有会话管理器和排行榜
type Server struct {
	sessions *SessionManager
	lb       *Leaderboard
}

func NewServer(lb *Leaderboard) *Server {
	return &Server{
		sessions: NewSessionManager(),
		lb:       lb,
	}
}

// Handler 注册所有路由
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 静态文件
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", s.handleIndex)

	// Game API
	mux.HandleFunc("/api/start", s.handleStart)
	mux.HandleFunc("/api/swap", s.handleSwap)
	mux.HandleFunc("/api/next", s.handleNext)
	mux.HandleFunc("/api/quit", s.handleQuit)
	mux.HandleFunc("/api/leaderboard", s.handleLeaderboard)

	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

// ─────────── /api/start ───────────

type StartRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "无名侠"
	}

	// 检查是否有旧 session（重开时先提交旧成绩）
	if cookie, err := r.Cookie("session_id"); err == nil {
		if old, ok := s.sessions.Get(cookie.Value); ok {
			s.lb.Submit(old.PlayerName, old.Level-1)
			s.sessions.Delete(cookie.Value)
		}
	}

	sessionID := generateSessionID()
	state := NewGameState(name)
	s.sessions.Set(sessionID, state)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})

	writeJSON(w, state)
}

// ─────────── /api/swap ───────────

type SwapRequest struct {
	Idx1 int `json:"idx1"`
	Idx2 int `json:"idx2"`
}

func (s *Server) handleSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state, ok := s.getSession(r)
	if !ok {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}

	// 只有 playing 状态才接受操作
	if state.Status != StatusPlaying {
		writeJSON(w, state)
		return
	}

	var req SwapRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	idx1, idx2 := req.Idx1, req.Idx2
	n := state.N

	if idx1 < 0 || idx1 >= n || idx2 < 0 || idx2 >= n || idx1 == idx2 {
		state.Feedback = "⚠️ 选择无效！"
		writeJSON(w, state)
		return
	}

	// 计算步数消耗（简化机制：每次交换固定消耗 1 步）
	cost := 1

	state.Heights[idx1], state.Heights[idx2] = state.Heights[idx2], state.Heights[idx1]
	state.CurrentSteps -= cost

	newWater := trap(state.Heights)

	if newWater == state.TargetScore {
		// ✅ 通关！
		state.Status = StatusWonLevel
		state.CurrentWater = newWater
		state.Revealed = true
		state.TargetScoreReveal = state.TargetScore
		state.Feedback = "🎯 目标达成！"
		if state.CurrentSteps < 0 {
			state.CurrentSteps = 0
		}
	} else if state.CurrentSteps <= 0 {
		// ❌ 步数耗尽
		state.Status = StatusGameOver
		state.CurrentWater = newWater
		state.Revealed = true
		state.TargetScoreReveal = state.TargetScore
		state.Feedback = "💀 步数耗尽！"
		s.lb.Submit(state.PlayerName, state.Level-1)
	} else {
		// 继续游戏，给出涨跌反馈
		if newWater > state.LastWater {
			state.Feedback = "🔥 涨了！"
		} else if newWater < state.LastWater {
			state.Feedback = "❄️ 跌了！"
		} else {
			state.Feedback = "⚪ 没变化"
		}
		state.LastWater = newWater
		state.CurrentWater = newWater
	}

	writeJSON(w, state)
}

// ─────────── /api/next ───────────

func (s *Server) handleNext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state, ok := s.getSession(r)
	if !ok {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}

	if state.Status != StatusWonLevel {
		writeJSON(w, state)
		return
	}

	// 进入下一关（增加奖励步数，initLevel 会处理难度递增）
	state.CurrentSteps += StepsReward
	state.Level++
	initLevel(state)

	writeJSON(w, state)
}

// ─────────── /api/quit ───────────

func (s *Server) handleQuit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state, ok := s.getSession(r)
	if !ok {
		http.Error(w, "no session", http.StatusUnauthorized)
		return
	}

	// 提交当前成绩再清除 session
	s.lb.Submit(state.PlayerName, state.Level-1)

	if cookie, err := r.Cookie("session_id"); err == nil {
		s.sessions.Delete(cookie.Value)
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// ─────────── /api/leaderboard ───────────

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	entries := s.lb.Top5()
	writeJSON(w, entries)
}

// ─────────── helpers ───────────

func (s *Server) getSession(r *http.Request) (*GameState, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, false
	}
	return s.sessions.Get(cookie.Value)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}
