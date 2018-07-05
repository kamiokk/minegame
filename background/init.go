package background

func initBackground() {
	go HandleBalanceLogQueue()
}