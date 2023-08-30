package config

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/fsnotify/fsnotify"
	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
)

var (
	Cfg CfgServer
	log = mylog.NewLogger("config", mylog.LevelDebug)
)

func InitCfg(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "../config/config.yaml"
	}
	log.Info("config file：", configFilePath)
	if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
		return fmt.Errorf("UnmarshalYamlFile err:%s", err.Error())
	}
	log.Info("config file：ok")
	return nil
}

func AddCfgFileWatcher(configFilePath string) (*fsnotify.Watcher, error) {
	if configFilePath == "" {
		configFilePath = "../config/config.yaml"
	}
	return toolib.AddFileWatcher(configFilePath, func() {
		log.Info("update config file：", configFilePath)
		if err := toolib.UnmarshalYamlFile(configFilePath, &Cfg); err != nil {
			log.Error("UnmarshalYamlFile err:", err.Error())
		}
		log.Info("update config file：ok")
	})
}

type DbMysql struct {
	Addr        string `json:"addr" yaml:"addr"`
	User        string `json:"user" yaml:"user"`
	Password    string `json:"password" yaml:"password"`
	DbName      string `json:"db_name" yaml:"db_name"`
	MaxOpenConn int    `json:"max_open_conn" yaml:"max_open_conn"`
	MaxIdleConn int    `json:"max_idle_conn" yaml:"max_idle_conn"`
}

type CfgServer struct {
	Slb struct {
		SvrName string   `json:"svr_name" yaml:"svr_name"`
		Servers []Server `json:"servers" yaml:"servers"`
	} `json:"slb" yaml:"slb"`
	Server struct {
		IsUpdate               bool              `json:"is_update" yaml:"is_update"`
		Net                    common.DasNetType `json:"net" yaml:"net"`
		HttpServerAddr         string            `json:"http_server_addr" yaml:"http_server_addr"`
		HttpServerInternalAddr string            `json:"http_server_internal_addr" yaml:"http_server_internal_addr"`
		ParserUrl              string            `json:"parser_url" yaml:"parser_url"`
		PayServerAddress       string            `json:"pay_server_address" yaml:"pay_server_address"`
		PayPrivate             string            `json:"pay_private" yaml:"pay_private"`
		ServerAddress          string            `json:"server_address" yaml:"server_address"`
		ServerPrivateKey       string            `json:"server_private_key" yaml:"server_private_key"`
		SplitCkb               uint64            `json:"split_ckb" yaml:"split_ckb"`
		RemoteSignApiUrl       string            `json:"remote_sign_api_url" yaml:"remote_sign_api_url"`
		PushLogUrl             string            `json:"push_log_url" yaml:"push_log_url"`
		PushLogIndex           string            `json:"push_log_index" yaml:"push_log_index"`
		NotExit                bool              `json:"not_exit" yaml:"not_exit"`
	} `json:"server" yaml:"server"`

	Origins []string `json:"origins" yaml:"origins"`
	Notify  struct {
		LarkErrorKey   string `json:"lark_error_key" yaml:"lark_error_key"`
		DiscordWebhook string `json:"discord_webhook" yaml:"discord_webhook"`
	} `json:"notify" yaml:"notify"`
	Chain struct {
		CkbUrl             string `json:"ckb_url" yaml:"ckb_url"`
		IndexUrl           string `json:"index_url" yaml:"index_url"`
		CurrentBlockNumber uint64 `json:"current_block_number" yaml:"current_block_number"`
		ConfirmNum         uint64 `json:"confirm_num" yaml:"confirm_num"`
		ConcurrencyNum     uint64 `json:"concurrency_num" yaml:"concurrency_num"`
	} `json:"chain" yaml:"chain"`
	DB struct {
		Mysql       DbMysql `json:"mysql" yaml:"mysql"`
		ParserMysql DbMysql `json:"parser_mysql" yaml:"parser_mysql"`
	} `json:"db" yaml:"db"`
	Cache struct {
		Redis struct {
			Addr     string `json:"addr" yaml:"addr"`
			Password string `json:"password" yaml:"password"`
			DbNum    int    `json:"db_num" yaml:"db_num"`
		} `json:"redis" yaml:"redis"`
	} `json:"cache" yaml:"cache"`
	SuspendMap map[string]string `json:"suspend_map" yaml:"suspend_map"`
	OldCid     map[string]string `json:"old_cid" yaml:"old_cid"`
}

type Server struct {
	Name   string `json:"name" yaml:"name"`
	Url    string `json:"url" yaml:"url"`
	Weight int    `json:"weight" yaml:"weight"`
}
