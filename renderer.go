package main

import "fmt"

// RenderBlackBox 渲染盲盒界面
// isWin: true 显示真实柱子，false 显示问号
func RenderBlackBox(heights []int, currentWater int, isWin bool, steps int) {
	// 清屏 (兼容 Linux/Mac/Git Bash)
	fmt.Print("\033[H\033[2J")

	fmt.Println("\n🔒 ==========  黑 盒 破 解 (Black Box)  ========== 🔒")
	fmt.Printf("📊 当前步数: %d\n", steps)
	fmt.Println("-----------------------------------------------------")

	// 核心反馈：只显示当前水量
	fmt.Printf("\n      💧 当前水量:  \033[1;36m%d\033[0m  \n\n", currentWater)
	fmt.Println("-----------------------------------------------------")

	// 打印位置索引
	fmt.Print("位置:   ")
	for i := 0; i < len(heights); i++ {
		fmt.Printf(" [%d] ", i)
	}
	fmt.Println("\n")

	// 打印柱子状态
	fmt.Print("高度:   ")
	for i := 0; i < len(heights); i++ {
		if isWin {
			// 胜利时刻：显示橙色真实高度
			fmt.Printf(" \033[38;5;208m%d\033[0m   ", heights[i])
		} else {
			// 游戏中：显示深灰色问号
			// ✅ 正确写法 (删掉后面的 heights[i]):
			fmt.Printf(" \033[1;30m?\033[0m   ")
		}
	}
	fmt.Println("\n\n-----------------------------------------------------")

	if isWin {
		fmt.Println("🎉 破解成功！这就是完美的水库形态！")
		fmt.Println("恭喜你找到了理论最大值。")
	} else {
		fmt.Println("指令示例: swap 0 5 (交换第0和第5个位置) | q (退出)")
		fmt.Print("👉 你的操作: ")
	}
}
