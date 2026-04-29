package utils

import (
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const base = 62

// Encode converts a numeric ID to a Base62 string
func Encode(id uint64) string {
	if id == 0 {
		return strings.Repeat(string(base62Chars[0]), 8)
	}

	var encodedBuilder strings.Builder
	for id > 0 {
		rem := id % base
		encodedBuilder.WriteByte(base62Chars[rem])
		id = id / base
	}

	// Reverse the string
	encoded := encodedBuilder.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	result := string(runes)
	return result
}

// GenerateHashID creates a non-sequential 8-character combination from a numeric ID
func GenerateHashID(id uint64) string {
	// 1. Hash the ID to scatter the bits (makes it look random instead of sequential)
	// We use a simple multiplication by a large prime and XOR to obfuscate the ID
	// Or simply use the built-in hash/fnv, but doing a quick obfuscation is faster:
	obfuscatedId := (id ^ 0x5bf036354641ce19) * 0x01000193
	
	// 2. Encode the obfuscated ID to Base62
	encoded := Encode(obfuscatedId)

	// 3. Ensure the result is exactly 8 characters long
	if len(encoded) > 8 {
		encoded = encoded[:8]
	} else if len(encoded) < 8 {
		// Pad with leading zeros if it's too short
		encoded = strings.Repeat(string(base62Chars[0]), 8-len(encoded)) + encoded
	}

	return encoded
}

// Decode converts a Base62 string back to a numeric ID
func Decode(encoded string) uint64 {
	var id uint64
	for _, char := range encoded {
		index := strings.IndexRune(base62Chars, char)
		id = id*base + uint64(index)
	}
	return id
}
