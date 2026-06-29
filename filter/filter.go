package filter

import (
	"strings"

	"github.com/AnnaTarantina/CommentServise/models"
)

var forbiddenWords = map[string]bool{
	"qwerty": true,
	"йцукен": true,
	"zxvbnm": true,
}

// CheckComment проверяет текст комментария на наличие запрещённых слов
func CheckComment(text string) models.FilterResult {
	result := models.FilterResult{IsApproved: true}

	// Приводим текст к нижнему регистру ОДИН раз перед циклом
	textLower := strings.ToLower(text)

	for word := range forbiddenWords {
		if strings.Contains(textLower, strings.ToLower(word)) {
			result.IsApproved = false
			break
		}
	}
	return result
}
