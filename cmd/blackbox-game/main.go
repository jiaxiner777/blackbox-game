package main

import (
	"blackbox-game/internal/store"
	"blackbox-game/internal/web"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	lb := store.LoadLeaderboard()
	srv := web.NewServer(lb)

	localIP := getLocalIP()

	fmt.Println("==============================================")
	fmt.Println("🎮  SKYLINE TYCOON  黑盒挑战 - Web 服务器")
	fmt.Println("==============================================")
	fmt.Printf("📡 本机访问:   http://localhost:8080\n")
	if localIP != "" {
		fmt.Printf("🌐 局域网访问: http://%s:8080\n", localIP)
	}
	fmt.Println("==============================================")
	fmt.Println("  按 Ctrl+C 停止服务器")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":8080", srv.Handler()))
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
