/*
TWO-SUM-Problem
Given an array of integers nums and an integer target, return indices of the two numbers such that they add up to target. You may assume that each input would have exactly one solution, and you may not use the same element twice.
You can return the answer in any order.

*/


package main

import (
	"fmt"
)

func twoSum(nums []int, target int) []int {
	// Create a map to store the index of each element
	numMap := make(map[int]int)

	// Iterate through the nums array
	for i, num := range nums {
		// Calculate the complement that would sum to the target
		complement := target - num
		
		// Check if the complement is already in the map
		if index, found := numMap[complement]; found {
			// Return the indices of the two numbers that add up to the target
			return []int{index, i}
		}
		
		// Store the current number's index in the map
		numMap[num] = i
	}

	// If no solution is found, return an empty slice (should not happen as per the problem constraints)
	return []int{}
}

func main() {
	// Example usage
	nums := []int{2, 7, 11, 15}
	target := 9
	result := twoSum(nums, target)
	fmt.Println(result) // Output: [0, 1] or [1, 0]
}

