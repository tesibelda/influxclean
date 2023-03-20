// sliceplus package provides helper functions for slices
//
// Author: Tesifonte Belda
// License: The MIT License (MIT)

package sliceplus

import (
	"strings"
)

// ChunkSlice returns chunks of max size for the given slice
func ChunkSlice(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// Difference returns items unique to slice1
func Difference(slice1, slice2 []string) []string {
	var diff []string
outer:
	for _, v1 := range slice1 {
		for _, v2 := range slice2 {
			if v1 == v2 {
				continue outer
			}
		}
		diff = append(diff, v1)
	}
	return diff
}

// Split2Dims returns two slices splitting elements with the given separator
func Split2Dims(ch []string, sep string) ([]string, []string) {
	var vals, vals1, vals2 []string
	for _, dims := range ch {
		vals = strings.Split(dims, sep)
		if len(vals) == 2 {
			vals1 = append(vals1, vals[0])
			vals2 = append(vals2, vals[1])
		}
	}
	return vals1, vals2
}
