package tonality

import (
	"errors"
	"strings"
)

// KeyNotation is the type representing a type of key notation.
type KeyNotation int

// Key notatations are the available notations for converting to and from.
const (
	CamelotKeys KeyNotation = iota
	OpenKey
	Musical
	MusicalAlt
	Beatport
)

var (
	// NotationCamelotKeys are the keys in camelot key notation.
	NotationCamelotKeys = []string{
		"1A", "1B",
		"2A", "2B",
		"3A", "3B",
		"4A", "4B",
		"5A", "5B",
		"6A", "6B",
		"7A", "7B",
		"8A", "8B",
		"9A", "9B",
		"10A", "10B",
		"11A", "11B",
		"12A", "12B",
	}

	// NotationOpenKey are the keys in open key notation.
	NotationOpenKey = []string{
		"6M", "6D",
		"7M", "7D",
		"8M", "8D",
		"9M", "9D",
		"10M", "10D",
		"11M", "11D",
		"12M", "12D",
		"1M", "1D",
		"2M", "2D",
		"3M", "3D",
		"4M", "4D",
		"5M", "5D",
	}

	// NotationMusical are the keys in musical notation.
	NotationMusical = []string{
		"Abm", "B",
		"Ebm", "Gb",
		"Bbm", "Db",
		"Fm", "Ab",
		"Cm", "Eb",
		"Gm", "Bb",
		"Dm", "F",
		"Am", "C",
		"Em", "G",
		"Bm", "D",
		"Gbm", "A",
		"Dbm", "E",
	}

	// NotationMusicalAlt are the keys in alternate musical notation.
	NotationMusicalAlt = []string{
		"G#m", "B",
		"Ebm", "F#",
		"A#m", "Db",
		"Fm", "G#",
		"Cm", "D#",
		"Gm", "Bb",
		"Dm", "F",
		"Am", "C",
		"Em", "G",
		"Bm", "D",
		"F#m", "A",
		"C#m", "E",
	}

	// NotationBeatport are the keys in beatport notation.
	NotationBeatport = []string{
		"G#m", "Bmaj",
		"Ebm", "Gb",
		"Bbm", "Db",
		"Fmin", "Ab",
		"Cmin", "Eb",
		"Gmin", "Bb",
		"Dmin", "Fmaj",
		"Amin", "Cmaj",
		"Emin", "Gmaj",
		"Bmin", "Dmaj",
		"F#m", "Amaj",
		"C#m", "Emaj",
	}
)

var (
	notationToKeys = map[KeyNotation][]string{
		CamelotKeys: NotationCamelotKeys,
		OpenKey:     NotationOpenKey,
		Musical:     NotationMusical,
		MusicalAlt:  NotationMusicalAlt,
		Beatport:    NotationBeatport,
	}

	keyToNotationMap  = map[string]KeyNotation{}
	notationToKeysMap = map[KeyNotation][]string{}
)

func init() {
	for notation, keys := range notationToKeys {
		if _, ok := notationToKeysMap[notation]; !ok {
			notationToKeysMap[notation] = make([]string, 0)
		}

		for _, key := range keys {
			normalizedKey := normalizeKey(key)

			notationToKeysMap[notation] = append(notationToKeysMap[notation], normalizedKey)

			if _, ok := keyToNotationMap[normalizedKey]; !ok {
				keyToNotationMap[normalizedKey] = notation
			}
		}
	}
}

// ConvertKeyToNotation converts a key from its detected key notation to the
// specified key notation.
func ConvertKeyToNotation(key string, notation KeyNotation) (string, error) {
	idx := getKeyIndex(key)
	if idx == -1 {
		return "", errors.New("invalid key")
	}

	return notationToKeys[notation][idx], nil
}

func normalizeKey(key string) string {
	return strings.TrimLeft(strings.ToLower(key), "0")
}

func getKeyIndex(key string) int {
	normalizedKey := normalizeKey(key)

	if _, ok := keyToNotationMap[normalizedKey]; !ok {
		return -1
	}

	notation := getNotation(key)
	keys := notationToKeysMap[notation]

	found := -1

	for idx, k := range keys {
		if normalizedKey == k {
			found = idx
			break
		}
	}

	return found
}

func getNotation(key string) KeyNotation {
	normalizedKey := normalizeKey(key)
	return keyToNotationMap[normalizedKey]
}
