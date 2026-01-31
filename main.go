package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())

	// 1. 设置难度 (N=8)
	n := 8
	heights := make([]int, n)
	for i := 0; i < n; i++ {
		heights[i] = rand.Intn(7)
	}

	// 2. 电脑上帝视角：计算满分
	fmt.Println("🤖 系统正在暴力计算所有排列组合，生成谜题中...")
	targetScore := solveMaxWater(heights)

	// 3. 打乱数组
	rand.Shuffle(len(heights), func(i, j int) {
		heights[i], heights[j] = heights[j], heights[i]
	})

	reader := bufio.NewReader(os.Stdin)
	steps := 0

	// 4. 游戏循环
	for {
		currentScore := trap(heights)
		isWin := (currentScore == targetScore)

		// 调用渲染器
		RenderBlackBox(heights, currentScore, isWin, steps)

		if isWin {
			break
		}

		// 读取输入
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		parts := strings.Split(input, " ")

		if len(parts) > 0 && parts[0] == "q" {
			fmt.Printf("放弃了？这关的理论最高分其实是: %d\n", targetScore)
			break
		}

		if len(parts) == 2 {
			// 尝试把两个输入都转成数字
			idx1, err1 := strconv.Atoi(parts[0])
			idx2, err2 := strconv.Atoi(parts[1])

			// 如果转换成功，且索引在合法范围内
			if err1 == nil && err2 == nil && idx1 >= 0 && idx1 < n && idx2 >= 0 && idx2 < n {
				// 执行交换
				heights[idx1], heights[idx2] = heights[idx2], heights[idx1]
				steps++
				continue // 成功交换后，直接进入下一轮
			}
		}
	}
}
