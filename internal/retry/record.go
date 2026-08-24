package retry

// ShouldRetry 判断某次失败后是否还允许重试。attempt 是已经完成的
// 尝试次数，小于总上限时继续重试，否则转死信。
func (s *Scheduler) ShouldRetry(attempt int) bool {
	return attempt < s.policy.TotalAttempts()
}

// Exhausted 判断重试是否已经耗尽。
func (s *Scheduler) Exhausted(attempt int) bool {
	return attempt >= s.policy.TotalAttempts()
}
