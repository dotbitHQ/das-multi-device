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
	} `json:"server" yaml:"server"`

	AccountCheck struct {
		CroSpec             string `json:"spec" yaml:"spec"`
		Limit               int    `json:"limit" yaml:"limit"`
		CompareCount        int    `json:"compare_count" yaml:"compare_count"`
		THQCodeHash         string `json:"thq_code_hash" yaml:"thq_code_hash"`
		DasContractArgs     string `json:"das_contract_args" yaml:"das_contract_args"`
		DasContractCodeHash string `json:"das_contract_code_hash" yaml:"das_contract_code_hash"`
		AccountCellType     string `json:"account-cell-type" yaml:"account-cell-type"`
	} `json:"account_check" yaml:"account_check"`
	BusinessProcess struct {
		BusinessSuccessRate string `json:"business_success_rate" yaml:"business_success_rate"`
		TimeLimit           struct {
			DasRegister int `json:"das_register" yaml:"das_register"`
			SubAccount  int `json:"sub_account" yaml:"sub_account"`
		} `json:"time_limit" yaml:"time_limit"`
	} `json:"business_process" yaml:"business_process"`
	BalanceChange struct {
		StatisticBalance string `json:"statistic_balance" yaml:"statistic_balance"`
		AddressList      struct {
			PayServerAddress struct {
				Address      string `json:"address" yaml:"address"`
				TimeFrame    int    `json:"time_frame" yaml:"time_frame"`
				AmountChange uint64 `json:"amount_change" yaml:"amount_change"`
			} `json:"pay_server_address" yaml:"pay_server_address"`
			ContractAddress struct {
				Address      string `json:"address" yaml:"address"`
				TimeFrame    int    `json:"time_frame" yaml:"time_frame"`
				AmountChange uint64 `json:"amount_change" yaml:"amount_change"`
			} `json:"contract_address" yaml:"contract_address"`
		} `json:"address_list" yaml:"address_list"`
	} `json:"balance_change" yaml:"balance_change"`
	Origins []string `json:"origins" yaml:"origins"`
	Notify  struct {
		LarkMonitorKey string `json:"lark_monitor_key" yaml:"lark_monitor_key"`
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
}

type Server struct {
	Name   string `json:"name" yaml:"name"`
	Url    string `json:"url" yaml:"url"`
	Weight int    `json:"weight" yaml:"weight"`
}
