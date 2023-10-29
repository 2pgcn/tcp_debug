package conf

import (
	"fmt"
	"github.com/spf13/viper"
)

func InitCometServerConfig(CfgFile string) (c *Server) {
	if len(CfgFile) == 0 {
		panic(fmt.Errorf("config file %s not found", CfgFile))
	}
	viper.AddConfigPath(CfgFile)
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.UnmarshalKey("Server", &c); err != nil {
		panic(err)
	}
	fmt.Println(c)
	return
}

func InitCometClientConfig(CfgFile string) (c *Client) {
	if len(CfgFile) == 0 {
		panic(fmt.Errorf("config file %s not found", CfgFile))
	}
	viper.AddConfigPath(CfgFile)
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.UnmarshalKey("Client", &c); err != nil {
		panic(err)
	}
	fmt.Println(c)
	return
}
