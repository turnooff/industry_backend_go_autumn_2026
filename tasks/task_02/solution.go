package main

func rotateRunes(s string, shift int) string {
	runes := []rune(s)
	if len(runes) == 0 || shift%len(runes) == 0 {
		return string(runes)
	}
	shift = shift % len(runes)

	if shift > 0 {
		return string(runes[shift:]) + string(runes[:shift])
	}
	return string(runes[len(runes)+shift:]) + string(runes[:len(runes)+shift])
}
