package game

func GenerateSessionID() string {
	return generateSessionID()
}

func InitLevel(state *GameState) *GameState {
	return initLevel(state)
}

func Trap(height []int) int {
	return trap(height)
}
