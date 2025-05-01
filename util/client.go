package util

import "github.com/zelenin/go-tdlib/client"

// CopyFormattedText копирует FormattedText
func CopyFormattedText(formattedText *client.FormattedText) *client.FormattedText {
	result := *formattedText
	return &result
}
