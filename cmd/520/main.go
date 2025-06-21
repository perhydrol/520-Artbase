package main

import (
	demo520 "demo520/internal/520"
	"demo520/internal/pkg/log"
	"os"
)

// Go 程序的默认入口函数(主函数).
func main() {
	command := demo520.NewCommand()
	// 初始化日志
	log.Init(demo520.LogOptions())
	defer log.Sync() // Sync 将缓存中的日志刷新到磁盘文件中
	if err := command.Execute(); err != nil {
		os.Exit(2)
	}
}
