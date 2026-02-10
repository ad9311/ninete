package prog

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func ToLowerCamel(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})

	if len(words) == 0 {
		return ""
	}

	words[0] = strings.ToLower(words[0])

	for i := 1; i < len(words); i++ {
		words[i] = Capitalize(words[i])
	}

	return strings.Join(words, "")
}

func Capitalize(s string) string {
	if s == "" {
		return s
	}

	r, size := utf8.DecodeRuneInString(s)

	return string(unicode.ToUpper(r)) + strings.ToLower(s[size:])
}

func StringToUnixDate(date string) (int64, error) {
	parsedDate, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return 0, fmt.Errorf("failed to parse date, %w", err)
	}

	return parsedDate.Unix(), nil
}

func ParseID(idStr string) (int, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse id of value %s, %w", idStr, err)
	}

	return int(id), nil
}

// func UnixToStringDate(date int64) string {
// 	normalDate := time.Unix(date, 0)

// 	return FormatTime(normalDate)
// }

// func FormatTime(date time.Time) string {
// 	return date.Format(time.RFC3339)
// }

// func WithTimeout(d time.Duration, fn func(context.Context) error) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), d)
// 	defer cancel()

// 	return fn(ctx)
// }
