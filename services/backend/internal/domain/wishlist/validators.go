package wishlist

import "unicode/utf8"

const maxNoteLen = 500

func isValidNote(note string) bool {
	return utf8.RuneCountInString(note) <= maxNoteLen
}
