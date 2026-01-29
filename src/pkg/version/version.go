package version

import (
	"strconv"
	"strings"
)

// IsNewer はlatestがcurrentより新しいかを判定する
// Semantic Versioning (major.minor.patch) に基づいて比較
func IsNewer(latest, current string) bool {
	latestParts := Parse(latest)
	currentParts := Parse(current)

	// 各部分を比較
	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	// latestの方がパート数が多い場合は新しい（例: 1.0.1 > 1.0）
	return len(latestParts) > len(currentParts)
}

// Parse はバージョン文字列を整数のスライスにパースする
func Parse(version string) []int {
	parts := strings.Split(version, ".")
	result := make([]int, len(parts))

	for i, part := range parts {
		// 数字以外の文字を除去（例: "1.0.0-beta" -> "1.0.0"）
		numStr := ""
		for _, c := range part {
			if c >= '0' && c <= '9' {
				numStr += string(c)
			} else {
				break
			}
		}
		if numStr != "" {
			result[i], _ = strconv.Atoi(numStr)
		}
	}

	return result
}
