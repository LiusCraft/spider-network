package main

import (
	"github.com/liuscraft/spider-network/client"
	"github.com/liuscraft/spider-network/pkg/xlog"
)

func main() {
	xl := xlog.New()
	xl.Info("Starting spider client...")
	uid := "test-client-1"
	xl = xlog.WithLogId(xl, uid)
	cli := client.NewClient(uid, "Test Client 1")
	if err := cli.Connect("127.0.0.1:19730"); err != nil {
		xl.Errorf("Failed to connect to server: %v", err)
		return
	}
	xl.Info("Connected to server successfully")

	// 启动命令行界面
	cli.StartCLI()
}
