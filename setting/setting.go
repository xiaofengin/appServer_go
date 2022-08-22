package setting

import "gopkg.in/ini.v1"

var Conf = new(AppConfig)

// AppConfig 应用程序配置
type AppConfig struct {
	Release      bool `ini:"release"`
	Port         int  `ini:"port"`
	*MySQLConfig `ini:"mysql"`
	*AgoraConfig `ini:"agora"`
}

// MySQLConfig 数据库配置
type MySQLConfig struct {
	User     string `ini:"user"`
	Password string `ini:"password"`
	DB       string `ini:"db"`
	Host     string `ini:"host"`
	Port     int    `ini:"port"`
}

// AgoraConfig 配置
type AgoraConfig struct {
	AgoraAppId  string `ini:"agoraAppId"`
	AgoraCert   string `ini:"agoraCert"`
	AgoraExpire uint32 `ini:"agoraExpire"`
	OrgName     string `ini:"orgName"`
	AppName     string `ini:"appName"`
	BaseUri     string `ini:"baseUri"`
}

func Init() error {
	c, err := ini.Load("./conf/config.ini")
	c.MapTo(&Conf)
	return err
}
