package game

// trap 计算当前的接雨水量 (LeetCode 42 完美版)
// 这是游戏的“记分员”，负责判断当前局面是多少分
func trap(height []int) int {
	n := len(height)
	if n == 0 {
		return 0
	}
	num := 0

	// 1. 准备两个本子
	leftMax := make([]int, n)
	rightMax := make([]int, n)

	// 2. 从左往右，记录左边最高 (继承冠军逻辑)
	leftMax[0] = height[0]
	for i := 1; i < n; i++ {
		if height[i] > leftMax[i-1] {
			leftMax[i] = height[i]
		} else {
			leftMax[i] = leftMax[i-1]
		}
	}

	// 3. 从右往左，记录右边最高 (倒叙循环)
	rightMax[n-1] = height[n-1]
	for j := n - 2; j >= 0; j-- {
		if height[j] > rightMax[j+1] {
			rightMax[j] = height[j]
		} else {
			rightMax[j] = rightMax[j+1]
		}
	}

	// 4. 计算总水量
	for k := 0; k < n; k++ {
		minMax := rightMax[k]
		if leftMax[k] < rightMax[k] {
			minMax = leftMax[k]
		}

		if minMax > height[k] {
			num += minMax - height[k]
		}
	}
	return num
}

// solveMaxWater 暴力破解器 (Solver)
// 在后台算出这组数据的理论最高分，作为通关目标
func solveMaxWater(arr []int) int {
	maxWater := 0

	// 递归全排列函数
	var permute func([]int, int)
	permute = func(a []int, k int) {
		if k == len(a) {
			w := trap(a)
			if w > maxWater {
				maxWater = w
			}
			return
		}

		for i := k; i < len(a); i++ {
			a[k], a[i] = a[i], a[k] // 交换
			permute(a, k+1)         // 递归
			a[k], a[i] = a[i], a[k] // 回溯
		}
	}

	temp := make([]int, len(arr))
	copy(temp, arr)
	permute(temp, 0)

	return maxWater
}
