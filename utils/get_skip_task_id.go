package utils

import (
	"fmt"
	"strconv"
	"strings"
)
// GetSkipTaskID extracts the task ID from a status string that starts with "\fskip_".
func GetSkipTaskID(status string) (int, error) {
	if !strings.HasPrefix(status, "\fskip_") {
		return 0, fmt.Errorf("status does not start with 'skip_'")
	}

	parts := strings.Split(status, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid status format")
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid task ID: %v", err)
	}

	return id, nil
}