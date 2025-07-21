package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// Version #2 of GetWaitingTaskID function
func GetWaitingTaskID(status string) (int, error) {
    // Remove the \f prefix if it exists
    cleanStatus := strings.TrimPrefix(status, "\f")
    
    // Check that the status starts with "waiting_"
    if !strings.HasPrefix(cleanStatus, "waiting_") {
        return 535, fmt.Errorf("status does not start with 'waiting_'")
    }
    
    // Separate by "_"
    parts := strings.Split(cleanStatus, "_")
    if len(parts) != 2 {
        return 545, fmt.Errorf("invalid status format")
    }
    
    // Convert ID to number
    id, err := strconv.Atoi(parts[1])
    if err != nil {
        return 555, fmt.Errorf("invalid task ID: %v", err)
    }
    
    return id, nil
}

// Version #1 of GetWaitingTaskID function
// func GetWaitingTaskID(status string) (int, error) {
// 	if !strings.HasPrefix(status, "\fwaiting") {
// 		return 535, fmt.Errorf("status does not start with 'waiting'")
// 	}
// 	parts := strings.Split(status, "_")
// 	if len(parts) != 2 {
// 		return 545, fmt.Errorf("invalid status format")
// 	}
// 	id, err := strconv.Atoi(parts[1])
// 	if err != nil {
// 		return 555, fmt.Errorf("invalid task ID: %v", err)
// 	}
// 	return id, nil
// }