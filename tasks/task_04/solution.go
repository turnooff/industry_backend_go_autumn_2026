package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}

	var results Stats

	results.Min = nums[1] - nums[0]
	results.Max = nums[1] - nums[0]

	results.Count = len(nums) - 1

	for i := 1; i < len(nums); i++ {
		diff := nums[i] - nums[i-1]

		results.Sum += diff
		results.Min = min(diff, results.Min)
		results.Max = max(diff, results.Max)
	}

	return results
}
