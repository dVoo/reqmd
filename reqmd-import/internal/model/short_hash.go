package model

import "fmt"

// ShortHash produces a stable 7-hex-char FNV-1a-64 hash over (pkg, name),
// masked to 28 bits. It exists as a single shared utility so the cobra
// `extract` command and any future entry point produce byte-identical
// IDs for the same source. Two callers previously duplicated this
// function; keeping it here removes the maintenance hazard.
//
// The FNV-1a-64 offset basis and prime are the standard constants
// (0xcbf29ce484222325 and 0x100000001b3, respectively). The 28-bit
// mask is "7 hex chars" — long enough to make accidental collisions
// negligible for any realistic source tree, short enough to keep IDs
// readable in generated .md files.
//
// A single XOR of the ':' separator is mixed in between the pkg and
// name halves: adding the separator as a normal hashed byte would
// self-cancel under XOR, so the explicit XOR is what disambiguates
// ("x:y") from ("x::y") and similar.
func ShortHash(pkg, name string) string {
	h := uint64(14695981039346656037) // FNV-1a-64 offset basis
	for i := 0; i < len(pkg); i++ {
		h ^= uint64(pkg[i])
		h *= 1099511628211 // FNV-1a-64 prime
	}
	h ^= uint64(':')
	for i := 0; i < len(name); i++ {
		h ^= uint64(name[i])
		h *= 1099511628211
	}
	return fmt.Sprintf("%07x", h&0xfffffff)
}
