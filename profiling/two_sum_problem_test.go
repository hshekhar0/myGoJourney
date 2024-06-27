package main

import (
	"testing"
)

func BenchmarkTwoSum(b *testing.B) {
	nums := []int{2, 7, 11, 15}
	target := 9

	for i := 0; i < b.N; i++ {
		twoSum(nums, target)
	}
}

