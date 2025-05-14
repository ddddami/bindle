package strutil

import (
	"regexp"
	"strings"
)

// Truncate truncates a string to a specified length and adds an ellipsis if it exceeds that length.
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[:max] + "..."
}

func SanitizeURL(input string) string {
	output := strings.ToLower(input)
	output = strings.ReplaceAll(output, " ", "-")

	reg := regexp.MustCompile("[^a-z0-9-]")
	output = reg.ReplaceAllString(output, "")

	// Replace multiple hyphens with a single hyphen
	reg = regexp.MustCompile("-+")
	output = reg.ReplaceAllString(output, "-")

	output = strings.Trim(output, "-")
	return output
}

func IsValidEmail(email string) bool {
	emailRX := regexp.MustCompile(
		"(?i)^[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@" +
			"[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+$",
	)

	return emailRX.MatchString(email)
}

// StripHTML removes HTML tags from a string
func StripHTML(input string) string {
	// This is a simple implementation that doesn't handle all edge cases
	// For production, you should probably use a HTML parser
	tagRegex := regexp.MustCompile(`(?s)<[^>]*>`)
	output := tagRegex.ReplaceAllString(input, "")

	replacements := map[string]string{
		"&nbsp;":  " ",
		"&amp;":   "&",
		"&lt;":    "<",
		"&gt;":    ">",
		"&quot;":  `"`,
		"&#39;":   "'",
		"&mdash;": "—",
	}

	for entity, replacement := range replacements {
		output = strings.ReplaceAll(output, entity, replacement)
	}

	return strings.TrimSpace(output)
}

func FormatFilename(input string) string {
	output := strings.ToLower(input)
	output = strings.ReplaceAll(output, " ", "_")
	output = regexp.MustCompile(`[\\/:*?"<>|]`).ReplaceAllString(output, "")

	// Replace multiple underscores with a single one
	output = regexp.MustCompile(`_+`).ReplaceAllString(output, "_")
	output = strings.Trim(output, "_")

	if output == "" {
		output = "file"
	}

	return output
}
