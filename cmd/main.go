package main

import (
	"context"
	"fmt"
	"github.com/2pgcn/tcp_debug/conf"
	"github.com/2pgcn/tcp_debug/internal/client"
	"github.com/2pgcn/tcp_debug/internal/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

var confPath string
var dail string
var startNum int

var mapping = [...]string{"", "", "abc", "def", "ghi", "jkl", "mno", "pqrs", "tuv", "wxyz"}

func main() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:9999", nil))
	}()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := NewServerArgs()
	cmd.SetContext(ctx)
	cmd.PersistentFlags().StringVar(&confPath, "conf", "", "cur conf.yaml path")
	cmd.PersistentFlags().StringVar(&dail, "dail", "", "cur conf.yaml path")
	cmd.PersistentFlags().IntVar(&startNum, "startNum", 1, "cur conf.yaml path")
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Kill, os.Interrupt)
	go func() {
		for {
			select {
			case sig := <-exitChan:
				fmt.Println("exitChan:", sig)
				cancel()
				return
			}
		}
	}()
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}

}

func NewServerArgs() *cobra.Command {
	return &cobra.Command{
		Short: "srv",
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "srv":
				c := conf.InitCometServerConfig(confPath)
				server.ServerTcp(cmd.Context(), c)
				break
			case "cli":
				c := conf.InitCometClientConfig(confPath)
				if c.DailUrl != "" {
					c.DailUrl = dail
				}
				c.StartNum = int32(startNum)
				client.ClientTcp(cmd.Context(), c)
				break
			}
		},
	}
}
