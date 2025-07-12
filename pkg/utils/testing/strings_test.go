package testing

import (
	"testing"

	"kalshi/pkg/utils"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"newline only", "\n", true},
		{"mixed whitespace", " \t\n ", true},
		{"non-empty", "hello", false},
		{"non-empty with spaces", " hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsEmpty(tt.input); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		suffix   string
		expected string
	}{
		{"shorter than limit", "hello", 10, "...", "hello"},
		{"exact length", "hello", 5, "...", "hello"},
		{"longer than limit", "hello world", 5, "...", "he..."},
		{"longer than limit no suffix", "hello world", 5, "", "hello"},
		{"suffix longer than limit", "hello", 2, "...", ".."},
		{"empty string", "", 5, "...", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.Truncate(tt.input, tt.length, tt.suffix); got != tt.expected {
				t.Errorf("Truncate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTruncateWords(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wordCount int
		suffix    string
		expected  string
	}{
		{"fewer words than limit", "hello world", 5, "...", "hello world"},
		{"exact word count", "hello world", 2, "...", "hello world"},
		{"more words than limit", "hello world test", 2, "...", "hello world..."},
		{"empty string", "", 3, "...", ""},
		{"single word", "hello", 1, "...", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.TruncateWords(tt.input, tt.wordCount, tt.suffix); got != tt.expected {
				t.Errorf("TruncateWords() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple camelCase", "camelCase", "camel_case"},
		{"PascalCase", "PascalCase", "pascal_case"},
		{"single word", "hello", "hello"},
		{"empty string", "", ""},
		{"multiple words", "helloWorldTest", "hello_world_test"},
		{"with numbers", "userID123", "user_i_d123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.CamelToSnake(tt.input); got != tt.expected {
				t.Errorf("CamelToSnake() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple snake_case", "snake_case", "snakeCase"},
		{"single word", "hello", "hello"},
		{"empty string", "", ""},
		{"multiple underscores", "hello_world_test", "helloWorldTest"},
		{"with numbers", "user_id_123", "userId123"},
		{"empty parts", "hello__world", "helloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.SnakeToCamel(tt.input); got != tt.expected {
				t.Errorf("SnakeToCamel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple words", "hello world", "Hello World"},
		{"single word", "hello", "Hello"},
		{"empty string", "", ""},
		{"already title case", "Hello World", "Hello World"},
		{"mixed case", "hElLo WoRlD", "Hello World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.TitleCase(tt.input); got != tt.expected {
				t.Errorf("TitleCase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCleanString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple text", "hello world", "hello world"},
		{"with special chars", "hello@world#test", "helloworldtest"},
		{"multiple spaces", "hello   world", "hello world"},
		{"with punctuation", "hello, world! test?", "hello, world! test?"},
		{"empty string", "", ""},
		{"only special chars", "@#$%^&*()", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.CleanString(tt.input); got != tt.expected {
				t.Errorf("CleanString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"email", "user@example.com", "example.com"},
		{"email with subdomain", "user@sub.example.com", "sub.example.com"},
		{"http url", "http://example.com/path", "example.com"},
		{"https url", "https://example.com:8080/path", "example.com"},
		{"url with subdomain", "https://api.example.com", "api.example.com"},
		{"no domain", "just text", "just text"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.ExtractDomain(tt.input); got != tt.expected {
				t.Errorf("ExtractDomain() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple text", "Hello World", "hello-world"},
		{"with underscores", "hello_world", "hello-world"},
		{"with special chars", "Hello@World#Test", "helloworldtest"},
		{"multiple spaces", "hello   world", "hello-world"},
		{"multiple hyphens", "hello--world", "hello-world"},
		{"leading/trailing hyphens", "-hello-world-", "hello-world"},
		{"empty string", "", ""},
		{"only special chars", "@#$%^&*()", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.Slugify(tt.input); got != tt.expected {
				t.Errorf("Slugify() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		visibleChars int
		maskChar     rune
		expected     string
	}{
		{"normal masking", "1234567890", 3, '*', "123****890"},
		{"short string", "123", 2, '*', "***"},
		{"empty string", "", 2, '*', ""},
		{"zero visible chars", "123456", 0, '*', "******"},
		{"custom mask char", "1234567890", 2, '#', "12######90"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.Mask(tt.input, tt.visibleChars, tt.maskChar); got != tt.expected {
				t.Errorf("Mask() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal email", "user@example.com", "u**r@example.com"},
		{"short username", "ab@example.com", "**@example.com"},
		{"single char username", "a@example.com", "*@example.com"},
		{"not an email", "notanemail", "no******il"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.MaskEmail(tt.input); got != tt.expected {
				t.Errorf("MaskEmail() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"hello", "world", "test"}

	tests := []struct {
		name          string
		slice         []string
		item          string
		caseSensitive bool
		expected      bool
	}{
		{"case sensitive match", slice, "hello", true, true},
		{"case sensitive no match", slice, "Hello", true, false},
		{"case insensitive match", slice, "HELLO", false, true},
		{"case insensitive no match", slice, "missing", false, false},
		{"empty slice", []string{}, "test", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.Contains(tt.slice, tt.item, tt.caseSensitive); got != tt.expected {
				t.Errorf("Contains() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"normal slice", []string{"hello", "", "world", "  ", "test"}, []string{"hello", "world", "test"}},
		{"no empty strings", []string{"hello", "world"}, []string{"hello", "world"}},
		{"all empty strings", []string{"", "  ", "\t"}, []string{}},
		{"empty slice", []string{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.RemoveEmptyStrings(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("RemoveEmptyStrings() length = %d, want %d", len(result), len(tt.expected))
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("RemoveEmptyStrings()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"duplicates", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"no duplicates", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"empty slice", []string{}, []string{}},
		{"single element", []string{"a"}, []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.UniqueStrings(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("UniqueStrings() length = %d, want %d", len(result), len(tt.expected))
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("UniqueStrings()[%d] = %v, want %v", i, v, tt.expected[i])
				}
			}
		})
	}
}
