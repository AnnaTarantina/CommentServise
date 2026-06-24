package filter

import (
	"strings"

	"../models"
)

var forbiddenWords = map[string]bool{
	"qwerty": true,
	"йцукен": true,
	"zxvbnm": true,
}

func CheckComment(text string) models.FilterResult {
	result := models.FilterResult{IsApproved: true}

	for word := range forbiddenWords {
		if containsIgnoreCase(text, word) {
			result.IsApproved = false
			break
		}
	}

	return result
}

func containsIgnoreCase(str, substr string) bool {
	strLower := strings.ToLower(str)
	substrLower := strings.ToLower(substr)
	return strings.Contains(strLower, substrLower)
}
