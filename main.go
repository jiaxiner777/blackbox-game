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

// 新增：新手引导界面
func showWelcome(reader *bufio.Reader) {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println("========================================")
	fmt.Println("🌊  S K Y L I N E   T Y C O O N  🌊")
	fmt.Println("       (天际线大亨：黑盒挑战)")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("📖  [ 游 戏 说 明 ]")
	fmt.Println("----------------------------------------")
	fmt.Println("1. 这是一个看不见的'接雨水'游戏。")
	fmt.Println("2. 你的目标是：移动柱子，让它们围成'碗状'，")
	fmt.Println("   从而接住最多的雨水。")
	fmt.Println("3. 所有的柱子高度都是隐藏的(显示为 ?)。")
	fmt.Println("4. 你只能通过'当前水量'的变化来推测柱子的高矮。")
	fmt.Println()
	fmt.Println("🎮  [ 操 作 方 法 ]")
	fmt.Println("----------------------------------------")
	fmt.Println("• 输入两个数字 (例如: 0 1) 来交换位置。")
	fmt.Println("• 交换需要消耗步数！步数耗尽则游戏结束。")
	fmt.Println("• 步数消耗 = 两个位置的距离 (移得越远越贵!)")
	fmt.Println()
	fmt.Println("🏆  [ 胜 利 秘 诀 ]")
	fmt.Println("----------------------------------------")
	fmt.Println("试着把最高的柱子移到最两边！")
	fmt.Println()
	fmt.Println("👉 按 [回车键] 开始挑战...")

	reader.ReadString('\n') // 等待用户按回车
}

func main() {
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	// 1. 显示新手引导
	showWelcome(reader)

	// 游戏状态
	level := 1
	currentSteps := InitialSteps
	n := 5 // 初始柱子数量

	for {
		// ====================
		// 每一关的初始化
		// ====================

		// 难度控制：每过两关，柱子加一根，最长12根
		if level > 1 && level%2 == 1 && n < 12 {
			n++
		}

		// 生成地图
		heights := make([]int, n)
		for i := 0; i < n; i++ {
			heights[i] = rand.Intn(6) + 1
		}

		// 计算本关目标
		targetScore := solveMaxWater(heights)

		// 打乱
		rand.Shuffle(len(heights), func(i, j int) {
			heights[i], heights[j] = heights[j], heights[i]
		})

		lastWater := trap(heights)
		feedback := ""

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

			// 渲染界面 (这里需要保证你已经在 renderer.go 或者 main.go 里有这个函数)
			// 如果你上一版把 RenderGameUI 放在了 main.go 里，就不用动
			RenderGameUI(level, n, currentSteps, currentWater, heights, feedback)

			feedback = ""

			// 获取输入
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			parts := strings.Split(input, " ")

			if input == "q" {
				fmt.Println("👋 溜了溜了...")
				return
			}

			// 解析输入: 0 4
			if len(parts) == 2 {
				idx1, err1 := strconv.Atoi(parts[0])
				idx2, err2 := strconv.Atoi(parts[1])

				if err1 == nil && err2 == nil && idx1 >= 0 && idx1 < n && idx2 >= 0 && idx2 < n {

					// 计算距离成本
					dist := idx1 - idx2
					if dist < 0 {
						dist = -dist
					}
					cost := dist

					if currentSteps < cost {
						feedback = fmt.Sprintf("\033[1;31m⚠️ 步数不足！需要 %d 步\033[0m", cost)
					} else {
						// 交换
						heights[idx1], heights[idx2] = heights[idx2], heights[idx1]
						currentSteps -= cost

						newWater := trap(heights)
						if newWater > lastWater {
							feedback = "\033[1;32m🔥 涨了！(Good)\033[0m"
						} else if newWater < lastWater {
							feedback = "\033[1;31m❄️ 跌了！(Bad)\033[0m"
						} else {
							feedback = "\033[1;30m⚪ 没变化\033[0m"
						}
						lastWater = newWater
					}
				} else {
					feedback = "⚠️ 坐标越界！"
				}
			}
		}

		// ====================
		// 结算阶段
		// ====================
		// 显示最终赢了的状态
		RenderGameUI(level, n, currentSteps, targetScore, heights, "")

		if levelWin {
			fmt.Printf("\n🎉 \033[1;33m恭喜通过 Level %d！\033[0m\n", level)
			fmt.Printf("🎁 获得奖励: +%d 步\n", StepsReward)
			currentSteps += StepsReward
			level++
			fmt.Println("\n👉 按 [回车键] 进入下一关...")
			reader.ReadString('\n')
		} else {
			fmt.Println("\n💀 \033[1;31m步数耗尽！挑战失败！\033[0m")
			fmt.Printf("本关目标是: %d\n", targetScore)
			fmt.Println("👉 按 [回车键] 重新开始...")
			reader.ReadString('\n')
			// 重置游戏
			level = 1
			currentSteps = InitialSteps
			n = 5
		}
	}
}

// 补充：为了防止报错，把渲染函数也保留在 main.go 里 (如果你之前分文件了，这段可以删掉)
func RenderGameUI(level, n, steps, water int, heights []int, feedback string) {
	fmt.Print("\033[H\033[2J")
	fmt.Println("========================================")
	fmt.Printf("🏰 \033[1;33mEndless Tower - Level %d\033[0m\n", level)
	fmt.Println("========================================")
	fmt.Print("⚡ 剩余步数: ")
	if steps <= 3 {
		fmt.Printf("\033[1;31m%d (警告!)\033[0m", steps)
	} else {
		fmt.Printf("\033[1;32m%d\033[0m", steps)
	}
	fmt.Println()
	fmt.Printf("💧 当前水量: \033[1;36m%d\033[0m", water)
	if feedback != "" {
		fmt.Printf("  %s", feedback)
	}
	fmt.Println("\n----------------------------------------")
	fmt.Print("索引: ")
	for i := 0; i < n; i++ {
		fmt.Printf("[%d] ", i)
	}
	fmt.Println()
	fmt.Print("柱子: ")
	for i := 0; i < n; i++ {
		fmt.Printf(" ?  ")
	}
	fmt.Println("\n\n👉 输入: 0 1 (交换索引0和1) | q (退出)")
}
