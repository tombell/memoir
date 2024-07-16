package valid

import (
	"cmp"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Empty checks if the given string is empty.
func Empty(value string) bool {
	return len(value) == 0
}

// NotEmpty checks if the given string is not empty.
func NotEmpty(value string) bool {
	return len(value) != 0
}

// StartsWith checks if the given string starts with the given prefix.
func StartsWith(value, prefix string) bool {
	return strings.HasPrefix(value, prefix)
}

// StartsWith checks if the given string ends with the given suffix.
func EndsWith(value, suffix string) bool {
	return strings.HasSuffix(value, suffix)
}

// MaxLength checks if the given string's UTF8 length is equal or less than the
// given max.
func MaxLength(value string, max int) bool {
	return utf8.RuneCountInString(value) <= max
}

// MinLength checks if the given string's UTF8 length is equal or greater than
// the given min.
func MinLength(value string, min int) bool {
	return min <= utf8.RuneCountInString(value)
}

// RangeLength checks if the given string's UTF8 length is within the given
// range.
func RangeLength(value string, min, max int) bool {
	return min <= utf8.RuneCountInString(value) && utf8.RuneCountInString(value) <= max
}

// Matches checks if the given string matches the given regular expression.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// IsNumber checks if the given string is all numbers.
func IsNumber(value string) bool {
	if _, err := strconv.Atoi(value); err != nil {
		return false
	}

	return true
}

// Max checks if the given value is equal or less than the given max.
func Max[T cmp.Ordered](value T, max T) bool {
	return cmp.Compare(value, max) == -1 || cmp.Compare(value, max) == 0
}

// Min checks if the given value is equal or greater than the given min.
func Min[T cmp.Ordered](value T, min T) bool {
	return cmp.Compare(value, min) == 1 || cmp.Compare(value, min) == 0
}

// Range checks if the given value is inside the given min and max.
func Range[T cmp.Ordered](value T, min, max T) bool {
	return (cmp.Compare(value, min) == 1 || cmp.Compare(value, min) == 0) &&
		(cmp.Compare(value, max) == -1 || cmp.Compare(value, max) == 0)
}

// Unique checks if the given slice of values are all unique.
func Unique[T cmp.Ordered](values []T) bool {
	s := slices.Clone(values)
	slices.Sort(s)

	if len(slices.Compact(s)) != len(values) {
		return false
	}

	return true
}

// In checks if the given value is contained in the given list of arguments.
func In[T comparable](value T, list ...T) bool {
	return slices.Contains(list, value)
}
