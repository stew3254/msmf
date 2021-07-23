package games

import (
	"strconv"
	"strings"
)

const MC_DEFAULT_PORT uint16 = 25565

// MCIsVersion checks if the string is actually a valid Minecraft version
func MCIsVersion(v string) bool {
	s := strings.Split(v, ".")
	if len(s) < 2 || len(s) > 3 {
		return false
	}
	for i, part := range s {
		n, err := strconv.Atoi(part)
		if err != nil {
			return false
		}
		if i == 0 && n < 1 {
			return false
		}
	}
	return true
}

// MCVersionCompare checks mainstream Minecraft versions, does not support snapshots
// Check to see if these are actually versions first before running this function
func MCVersionCompare(a, b string) int {
	aSlice := strings.Split(a, ".")
	bSlice := strings.Split(b, ".")

	// Check major version (This probably will always be the same)
	if aSlice[0] > bSlice[0] {
		return 1
	} else if aSlice[0] < bSlice[0] {
		return -1
	}

	// Check minor version
	if aSlice[1] > bSlice[1] {
		return 1
	} else if aSlice[1] < bSlice[1] {
		return -1
	}

	// First see if there is a patch number
	if len(aSlice) == 3 && len(bSlice) == 3 {
		// Check patch version
		if aSlice[2] > bSlice[2] {
			return 1
		} else if aSlice[2] < bSlice[2] {
			return -1
		}
	} else if len(aSlice) == 3 && len(bSlice) == 2 {
		return 1
	} else if len(aSlice) == 2 && len(bSlice) == 3 {
		return -1
	}

	// Everything must be the same
	return 0
}
