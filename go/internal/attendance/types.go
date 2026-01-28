package attendance

// GameStats corresponds to per-game statistics.
type GameStats struct {
	Total           int
	Succeeded       int
	AlreadyAttended int
	Failed          int
}

// ExecutionStats corresponds to overall execution statistics.
type ExecutionStats struct {
	Accounts struct {
		Total         int
		Successful    int
		Skipped       int
		Failed        int
		FailedIndexes []int
	}
	CharactersByGame map[int]*GameStats
}

// Result is the result of a full attendance run.
type Result struct {
	Result string          `json:"result"` // "success" or "failed"
	Stats  ExecutionStats  `json:"stats"`
}

