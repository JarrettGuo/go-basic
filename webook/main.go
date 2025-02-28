package main

import "github.com/spf13/viper"

func main() {

	initViper()
	server := InitWebServer()

	server.Run(":8080")
}

func initViper() {
	// 配置文件名字，不包含文件拓展名
	viper.SetConfigName("dev")
	// 配置用的文件格式
	viper.SetConfigType("yaml")
	// 当前工作目录下的 config 子目录
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
