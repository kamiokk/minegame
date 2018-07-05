package background

// InitBackground init background task
func InitBackground() {
	go HandleBalanceLogQueue()
}