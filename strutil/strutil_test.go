package strutil

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	const maxLength = 6
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "return input if input is less than max length",
			input:    "A",
			expected: "A",
		},

		{
			name:     "return truncated text with an ellipses if input is > than max length",
			input:    strings.Repeat("A", maxLength+1),
			expected: strings.Repeat("A", maxLength) + "...",
		},

		{
			name:     "return input if length of input equals maxLength",
			input:    strings.Repeat("A", maxLength),
			expected: strings.Repeat("A", maxLength),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Truncate(tc.input, maxLength)
			if got != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, got)
				return
			}
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "return input if input is empty",
			input:    "",
			expected: "",
		},
		{
			name:     "return input if input is safe",
			input:    "A",
			expected: "a",
		},
		{
			name:     "return santized output if input is unsafe",
			input:    "This is a Test Article 1!  With Spaces & Special----Characters?",
			expected: "this-is-a-test-article-1-with-spaces-special-characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeURL(tc.input)
			if got != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, got)
				return
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal.txt", "normal.txt"},
		{"file with spaces.jpg", "file_with_spaces.jpg"},
		{"../../../etc/passwd", "......etcpasswd"},
		{"file:with:colons.png", "filewithcolons.png"},
		{"file*with?invalid\"chars.xxx", "filewithinvalidchars.xxx"},
	}

	for _, tc := range testCases {
		result := FormatFilename(tc.input)
		if result != tc.expected {
			t.Errorf("For input '%s', expected '%s' but got '%s'", tc.input, tc.expected, result)
		}
	}
}

func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "dami@dami.me", true},
		{"valid subdomain", "user@mail.example.co.uk", true},
		{"valid plus", "hello+test@domain.io", true},
		{"valid hyphen", "first-last@domain.com", true},
		{"valid underscore", "user_name@sub.domain.org", true},

		{"missing @", "noatsymbol.com", false},
		{"no domain", "user@", false},
		{"consecutive dots", "user..name@domain.com", false},
		{"missing dot in domain", "user@domain", false},
		{"ending dot", "user@domain.com.", false},
		{"leading dot", ".user@domain.com", false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValidEmail(tc.input)
			if got != tc.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic tags",
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello World",
		},
		{
			name:     "nested tags with attributes",
			input:    `<div class="intro"><span style="color:red;">Red Text</span></div>`,
			expected: "Red Text",
		},
		{
			name:     "html entities",
			input:    "5 &lt; 10 &amp;&amp; 10 &gt; 5",
			expected: "5 < 10 && 10 > 5",
		},
		{
			name:     "mixed tags and entities",
			input:    `<p>Go is &quot;awesome&quot; &nbsp; &gt; JS</p>`,
			expected: `Go is "awesome"   > JS`,
		},
		{
			name:     "no tags",
			input:    "Just plain text",
			expected: "Just plain text",
		},
		{
			name:     "self-closing tags",
			input:    "Line break<br/>Next line",
			expected: "Line breakNext line",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "malformed tags",
			input:    "<<div>Oops</div>",
			expected: "Oops",
		},
		{
			name:     "no HTML",
			input:    "Just some plain txt",
			expected: "Just some plain txt",
		},
		{
			name:     "Multiline tags",
			input:    "<div\nclass=\"test\">Line</div>",
			expected: "Line",
		},
		{
			name:     "extra space",
			input:    " <p>Hello&nbsp;World</p> ",
			expected: "Hello World",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := StripHTML(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"All GOOD boys deserve FANTA", "all-good-boys-deserve-fanta"},
		{" N is Awesome i guess!", "n-is-awesome-i-guess"},
		{"  Leading and trailing spaces  ", "leading-and-trailing-spaces"},
		{"Special@#$%^&*()Characters", "specialcharacters"},
	}

	for _, tc := range testCases {
		result := Slugify(tc.input)
		if result != tc.expected {
			t.Errorf("For input '%s', expected '%s' but got '%s'", tc.input, tc.expected, result)
		}
	}
}
