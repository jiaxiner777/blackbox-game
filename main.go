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

// ==========================================
// 游戏配置
// ==========================================
const (
	InitialSteps = 15 // 初始步数
	StepsReward  = 5  // 过关奖励步数
)

func main() {
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	// 游戏状态
	level := 1
	currentSteps := InitialSteps
	n := 5 // 初始柱子数量

	for {
		// ====================
		// 每一关的初始化
		// ====================

		// 1. 难度控制：每过两关，柱子加一根，最长12根
		if level > 1 && level%2 == 1 && n < 12 {
			n++
		}

		// 2. 生成地图
		heights := make([]int, n)
		for i := 0; i < n; i++ {
			heights[i] = rand.Intn(6) + 1 // 1-6 高度 (避免全是0)
		}

		// 3. 计算本关目标 (上帝视角)
		fmt.Printf("\n🚀 正在生成 Level %d (柱子数: %d)...\n", level, n)
		targetScore := solveMaxWater(heights)

		// 4. 打乱
		rand.Shuffle(len(heights), func(i, j int) {
			heights[i], heights[j] = heights[j], heights[i]
		})

		// 记录上一轮的水量，用于给反馈
		lastWater := trap(heights)
		feedback := "" // 用于存 "好球!" 或 "糟糕!"

		// ====================
		// 关卡内循环
		// ====================
		levelWin := false
		for currentSteps > 0 {
			currentWater := trap(heights)

			// 检查胜利
			if currentWater == targetScore {
				levelWin = true
				break
			}

			// 渲染界面
			RenderGameUI(level, n, currentSteps, currentWater, heights, feedback)
			feedback = "" // 消费掉反馈，下次不显示

			// 获取输入
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			parts := strings.Split(input, " ")

			// 处理指令
			if input == "q" {
				fmt.Println("👋 溜了溜了...")
				return
			}

			// 解析输入: 0 4 (只输数字)
			if len(parts) == 2 {
				idx1, err1 := strconv.Atoi(parts[0])
				idx2, err2 := strconv.Atoi(parts[1])

				if err1 == nil && err2 == nil && idx1 >= 0 && idx1 < n && idx2 >= 0 && idx2 < n {
					// === 核心玩法：交换并扣费 ===
					heights[idx1], heights[idx2] = heights[idx2], heights[idx1]
					currentSteps--

					// 计算新水量，给出反馈
					newWater := trap(heights)
					if newWater > lastWater {
						feedback = "\033[1;32m🔥 涨了！方向正确！\033[0m"
					} else if newWater < lastWater {
						feedback = "\033[1;31m❄️ 跌了！刚才那步是臭棋！\033[0m"
					} else {
						feedback = "\033[1;30m⚪ 没变化...\033[0m"
					}
					lastWater = newWater
				} else {
					feedback = "⚠️ 坐标越界！"
				}
			}
		}

		// ====================
		// 结算阶段
		// ====================
		RenderGameUI(level, n, currentSteps, targetScore, heights, "") // 显示最终状态(当作赢了显示)

		if levelWin {
			fmt.Printf("\n🎉 \033[1;33m恭喜通过 Level %d！\033[0m\n", level)
			fmt.Printf("💧 完美水量: %d\n", targetScore)
			fmt.Printf("🎁 获得奖励: +%d 步\n", StepsReward)

			currentSteps += StepsReward
			level++
			fmt.Println("\n按回车键进入下一关...")
			reader.ReadString('\n')
		} else {
			fmt.Println("\n💀 \033[1;31m步数耗尽！挑战失败！\033[0m")
			fmt.Printf("本关目标是: %d，你只接了: %d\n", targetScore, lastWater)
			fmt.Printf("最终定格在 Level %d\n", level)
			break // 游戏彻底结束
		}
	}
}

// RenderGameUI 整合了渲染逻辑，直接放在这里方便你复制
func RenderGameUI(level, n, steps, water int, heights []int, feedback string) {
	// 清屏
	fmt.Print("\033[H\033[2J")

	fmt.Println("========================================")
	fmt.Printf("🏰 \033[1;33mEndless Tower - Level %d\033[0m\n", level)
	fmt.Println("========================================")

	// 步数血条 (关键压力源)
	fmt.Print("⚡ 剩余步数: ")
	if steps <= 3 {
		fmt.Printf("\033[1;31m%d (警告!)\033[0m", steps)
	} else {
		fmt.Printf("\033[1;32m%d\033[0m", steps)
	}
	fmt.Println()

	// 水量显示
	fmt.Printf("💧 当前水量: \033[1;36m%d\033[0m", water)
	if feedback != "" {
		fmt.Printf("  %s", feedback) // 显示冷热反馈
	}
	fmt.Println("\n----------------------------------------")

	// 打印位置
	fmt.Print("索引: ")
	for i := 0; i < n; i++ {
		fmt.Printf("[%d] ", i)
	}
	fmt.Println()

	// 打印盲盒状态 (这里为了增加策略，我们显示一个简易的图形而不是问号)
	// 让玩家能稍微记忆一下哪里是高哪里是低，但还是用 ? 遮挡高度数值
	fmt.Print("柱子: ")
	for i := 0; i < n; i++ {
		// 这里只显示 ?，你可以改成显示 'H' 'M' 'L' (High/Mid/Low) 只要你想降低难度
		fmt.Printf(" ?  ")
	}
	fmt.Println("\n\n👉 输入: 0 1 (交换索引0和1) | q (退出)")
}
