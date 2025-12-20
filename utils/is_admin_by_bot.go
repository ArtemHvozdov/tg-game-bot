package utils

func IsAdminByBot(listAdmin []int64, userID int64) bool {
	for _, adminID := range listAdmin {
		if userID == adminID {
			return true
		}
	}
	return false
}