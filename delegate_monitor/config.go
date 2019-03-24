package main

import (
	"github.com/larspensjo/config"
)

// Configure 配置参数结构体
type Configure struct {
	Main struct {
		Rpcuser         string
		Rpcpassword     string
		Rpcip           string
		Rpcport         int
		InitBlockHeight int
		AppCode         string
	}
	// Db struct {
	// 	Url    string
	// 	DbName string
	// }
}

// MustString 强制读取返回值
func MustString(str string, err error) string {
	if err != nil {
		//panic(err)
		return ""
	}
	return str
}

// MustInt 强制读取返回值
func MustInt(value int, err error) int {
	if err != nil {
		//panic(err)
		return 0
	}
	return value
}

// LoadConfigure 加载配置文件
func LoadConfigure(path string) (*Configure, error) {

	c, err := config.ReadDefault(path)
	if err != nil {
		return nil, err
	}
	configure := &Configure{}

	// main section

	configure.Main.Rpcuser = MustString(c.String("main", "rpcuser"))
	configure.Main.Rpcpassword = MustString(c.String("main", "rpcpassword"))
	configure.Main.Rpcip = MustString(c.String("main", "rpcip"))
	configure.Main.Rpcport = MustInt(c.Int("main", "rpcport"))
	configure.Main.InitBlockHeight = MustInt(c.Int("main", "initblockheight"))
	configure.Main.AppCode = MustString(c.String("main", "appcode"))
	// db section
	// configure.Db.Url = MustString(c.String("db", "url"))
	// configure.Db.DbName = MustString(c.String("db", "dbname"))

	return configure, nil
}
