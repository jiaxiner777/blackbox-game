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

	for {
		// 1. 显示主菜单
		if !showWelcome(reader) {
			break
		}

		// 2. 开始游戏循环
		playGame(reader)
	}

	fmt.Println("👋 感谢游玩，再见！")
}

// 显示欢迎界面
func showWelcome(reader *bufio.Reader) bool {
	fmt.Print("\033[H\033[2J") // 清屏
	fmt.Println("========================================")
	fmt.Println("🌊  S K Y L I N E   T Y C O O N  🌊")
	fmt.Println("       (天际线大亨：黑盒挑战)")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("📖  [ 游 戏 说 明 ]")
	fmt.Println("----------------------------------------")
	fmt.Println("1. 目标：移动柱子围成'碗状'接雨水。")
	fmt.Println("2. 柱子高度隐藏(显示为 ?)。")
	fmt.Println("3. 只能通过'水量变化'来推测高度。")
	fmt.Println()
	fmt.Println("🎮  [ 操 作 ]")
	fmt.Println("• 输入: 0 1 (交换位置)")
	fmt.Println("• 输入: q   (返回菜单)")
	fmt.Println()
	fmt.Println("👉 按 [回车键] 开始游戏...")
	fmt.Println("👉 输入 [exit] 彻底退出程序")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "exit" {
		return false
	}
	return true
}

// 游戏主要逻辑
func playGame(reader *bufio.Reader) {
	level := 1
	currentSteps := InitialSteps
	n := 5

	for {
		// --- 每一关的初始化 ---
		if level > 1 && level%2 == 1 && n < 12 {
			n++
		}

		heights := make([]int, n)
		for i := 0; i < n; i++ {
			heights[i] = rand.Intn(6) + 1
		}

		targetScore := solveMaxWater(heights)

		rand.Shuffle(len(heights), func(i, j int) {
			heights[i], heights[j] = heights[j], heights[i]
		})

		lastWater := trap(heights)
		feedback := ""

		// --- 关卡内循环 ---
		levelWin := false

		// 只要没输或者刚刚赢了，就继续
		for currentSteps > 0 {
			currentWater := trap(heights)

			// 1. 检查开局是否就赢了 (极小概率)
			if currentWater == targetScore {
				levelWin = true
				break
			}

			// 2. 渲染界面
			RenderGameUI(level, n, currentSteps, currentWater, heights, feedback, false)
			feedback = ""

			// 3. 获取输入
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			parts := strings.Split(input, " ")

			if input == "q" {
				fmt.Println("🔙 返回主菜单...")
				time.Sleep(500 * time.Millisecond)
				return
			}

			if len(parts) == 2 {
				idx1, err1 := strconv.Atoi(parts[0])
				idx2, err2 := strconv.Atoi(parts[1])

				if err1 == nil && err2 == nil && idx1 >= 0 && idx1 < n && idx2 >= 0 && idx2 < n {
					dist := idx1 - idx2
					if dist < 0 {
						dist = -dist
					}
					cost := dist

					// === 改动点 A: 移除步数不足的拦截 ===
					// 不再检查 currentSteps < cost，直接让玩家“透支”

					heights[idx1], heights[idx2] = heights[idx2], heights[idx1]
					currentSteps -= cost // 步数可能会变成负数，没关系！

					newWater := trap(heights)

					// === 改动点 B: 最后一搏的绝杀判断 ===
					// 如果这一步走完赢了，立刻锁定胜利，跳出循环
					// 否则下一次循环会因为 currentSteps <= 0 而判负
					if newWater == targetScore {
						levelWin = true
						break
					}

					if newWater > lastWater {
						feedback = "\033[1;32m🔥 涨了！\033[0m"
					} else if newWater < lastWater {
						feedback = "\033[1;31m❄️ 跌了！\033[0m"
					} else {
						feedback = "\033[1;30m⚪ 没变化\033[0m"
					}

					// 如果步数透支了，给个提示
					if currentSteps < 0 {
						// 这里不需要做额外处理，循环条件会自动处理结束
						// 但可以给个最后反馈（虽然玩家看不到了，因为马上会清屏结算）
					}

					lastWater = newWater

				} else {
					feedback = "⚠️ 坐标越界！"
				}
			}
		}

		// --- 结算阶段 ---
		// 无论输赢，先揭晓答案
		RenderGameUI(level, n, currentSteps, trap(heights), heights, "\033[1;33m✨ 答案揭晓 ✨\033[0m", true)

		if levelWin {
			fmt.Printf("\n🎉 \033[1;33m恭喜通过 Level %d！\033[0m\n", level)
			// 如果是透支过关，把步数回正到0再发奖励，或者直接给奖励
			if currentSteps < 0 {
				currentSteps = 0
			}
			fmt.Printf("🎁 获得奖励: +%d 步\n", StepsReward)
			currentSteps += StepsReward
			level++
			fmt.Println("\n👉 按 [回车键] 下一关... (或 q 返回)")
		} else {
			fmt.Println("\n💀 \033[1;31m步数耗尽！挑战失败！\033[0m")
			fmt.Printf("本关最佳答案是: %d\n", targetScore)
			// 这里实现了你想要的“再来一次”
			fmt.Println("\n👉 按 [回车键] 重新开始... (或 q 返回)")
			level = 1
			currentSteps = InitialSteps
			n = 5
		}

		nextInput, _ := reader.ReadString('\n')
		if strings.TrimSpace(nextInput) == "q" {
			return
		}
	}
}

// 界面渲染函数 (增加了 reveal 参数)
func RenderGameUI(level, n, steps, water int, heights []int, feedback string, reveal bool) {
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

	// === 核心修改：如果是揭晓模式，显示数字 ===
	for i := 0; i < n; i++ {
		if reveal {
			// 显示真实的数字，用黄色高亮
			fmt.Printf("\033[1;33m %d  \033[0m", heights[i])
		} else {
			// 隐藏模式，显示问号
			fmt.Printf(" ?  ")
		}
	}
	fmt.Println("\n\n👉 输入: 0 1 (交换) | q (菜单)")
}
