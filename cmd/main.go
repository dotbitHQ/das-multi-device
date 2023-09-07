package main

import (
	"context"
	"das-multi-device/block_parser"
	"das-multi-device/cache"
	"das-multi-device/config"
	"das-multi-device/dao"
	"das-multi-device/http_server"
	"das-multi-device/tool"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/sirupsen/logrus"

	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/toolib"

	"github.com/urfave/cli/v2"
	"os"
	"sync"
	"time"
)

var (
	//log               = mylog.NewLogger("main", mylog.LevelDebug)
	exit              = make(chan struct{})
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}
)

type DefaultFieldHook struct {
}

func (hook *DefaultFieldHook) Fire(entry *logrus.Entry) error {
	entry.Data["service_name"] = "das-multi-device"
	entry.Data["env"] = "dev"
	return nil
}

func (hook *DefaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
func main() {
	var hook DefaultFieldHook
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.AddHook(&hook)
	logrus.SetFormatter(&logrus.JSONFormatter{
		//DisableColors:   true,
		TimestampFormat: "2006-01-02 15:03:04",
	})
	tool.Log(nil).Debugf("start：")
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		tool.Log(nil).Fatal(err)
	}
}

func runServer(ctx *cli.Context) error {
	// config file
	configFilePath := ctx.String("config")
	if err := config.InitCfg(configFilePath); err != nil {
		return err
	}

	// config file watcher
	watcher, err := config.AddCfgFileWatcher(configFilePath)
	if err != nil {
		return fmt.Errorf("AddCfgFileWatcher err: %s", err.Error())
	}
	// ============= service start =============

	// db
	dbDao, err := dao.NewGormDB(config.Cfg.DB.Mysql, config.Cfg.DB.ParserMysql, true)
	if err != nil {
		return fmt.Errorf("NewGormDB err: %s", err.Error())
	}
	tool.Log(nil).Infof("db ok")
	// redis
	red, err := toolib.NewRedisClient(config.Cfg.Cache.Redis.Addr, config.Cfg.Cache.Redis.Password, config.Cfg.Cache.Redis.DbNum)
	if err != nil {
		tool.Log(nil).Info("NewRedisClient err: %s", err.Error())
		//return fmt.Errorf("NewRedisClient err:%s", err.Error())
	} else {
		tool.Log(nil).Info("redis ok")
	}

	rc := cache.Initialize(red)

	// das core
	dasCore, dasCache, err := initDasCore()
	if err != nil {
		return fmt.Errorf("initDasCore err: %s", err.Error())
	}
	// tx builder
	txBuilderBase, serverScript, err := initTxBuilder(dasCore)
	if err != nil {
		return fmt.Errorf("initTxBuilder err: %s", err.Error())
	}

	// block parser
	bp := block_parser.BlockParser{
		DasCore:            dasCore,
		CurrentBlockNumber: config.Cfg.Chain.CurrentBlockNumber,
		DbDao:              dbDao,
		ConcurrencyNum:     config.Cfg.Chain.ConcurrencyNum,
		ConfirmNum:         config.Cfg.Chain.ConfirmNum,
		Ctx:                ctxServer,
		Cancel:             cancel,
		Wg:                 &wgServer,
	}
	if err := bp.Run(); err != nil {
		return fmt.Errorf("block parser err: %s", err.Error())
	}
	tool.Log(nil).Info("block parser ok")

	// http
	hs, err := http_server.Initialize(http_server.HttpServerParams{
		Address:         config.Cfg.Server.HttpServerAddr,
		InternalAddress: config.Cfg.Server.HttpServerInternalAddr,
		DbDao:           dbDao,
		Rc:              rc,
		Ctx:             ctxServer,
		DasCore:         dasCore,
		DasCache:        dasCache,
		TxBuilderBase:   txBuilderBase,
		ServerScript:    serverScript,
	})
	if err != nil {
		return fmt.Errorf("http server Initialize err:%s", err.Error())
	}
	hs.Run()
	tool.Log(nil).Info("httpserver ok")
	// ============= service end =============
	toolib.ExitMonitoring(func(sig os.Signal) {
		tool.Log(nil).Warn("ExitMonitoring:", sig.String())
		if watcher != nil {
			tool.Log(nil).Warn("close watcher ... ")
			_ = watcher.Close()
		}
		cancel()
		wgServer.Wait()
		tool.Log(nil).Warn("success exit server. bye bye!")
		time.Sleep(time.Second)
		exit <- struct{}{}
	})

	<-exit

	return nil
}

func initDasCore() (*core.DasCore, *dascache.DasCache, error) {
	// ckb node
	ckbClient, err := rpc.DialWithIndexer(config.Cfg.Chain.CkbUrl, config.Cfg.Chain.IndexUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("rpc.DialWithIndexer err: %s", err.Error())
	}
	tool.Log(nil).Info("ckb node ok")

	// das init
	env := core.InitEnvOpt(config.Cfg.Server.Net, common.DasContractNameConfigCellType, common.DasContractNameAccountCellType,
		common.DasContractNameBalanceCellType, common.DasContractNameDispatchCellType, common.DasContractNameAlwaysSuccess,
		common.DASContractNameEip712LibCellType, common.DasKeyListCellType)
	ops := []core.DasCoreOption{
		core.WithClient(ckbClient),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(config.Cfg.Server.Net),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dasCore := core.NewDasCore(ctxServer, &wgServer, ops...)
	dasCore.InitDasContract(env.MapContract)
	if err := dasCore.InitDasConfigCell(); err != nil {
		return nil, nil, fmt.Errorf("InitDasConfigCell err: %s", err.Error())
	}
	if err := dasCore.InitDasSoScript(); err != nil {
		return nil, nil, fmt.Errorf("InitDasSoScript err: %s", err.Error())
	}
	dasCore.RunAsyncDasContract(time.Minute * 3)   // contract outpoint
	dasCore.RunAsyncDasConfigCell(time.Minute * 5) // config cell outpoint
	dasCore.RunAsyncDasSoScript(time.Minute * 7)   // so

	tool.Log(nil).Info("das contract ok")

	// das cache
	dasCache := dascache.NewDasCache(ctxServer, &wgServer)
	dasCache.RunClearExpiredOutPoint(time.Minute * 15)
	tool.Log(nil).Info("das cache ok")

	return dasCore, dasCache, nil
}

func initTxBuilder(dasCore *core.DasCore) (*txbuilder.DasTxBuilderBase, *types.Script, error) {
	payServerAddressArgs := ""
	var serverScript *types.Script
	if config.Cfg.Server.PayServerAddress != "" {
		parseAddress, err := address.Parse(config.Cfg.Server.PayServerAddress)
		if err != nil {
			tool.Log(nil).Error("pay server address.Parse err: ", err.Error())
		} else {
			payServerAddressArgs = common.Bytes2Hex(parseAddress.Script.Args)
			serverScript = parseAddress.Script
		}
	}
	var handleSign sign.HandleSignCkbMessage
	if config.Cfg.Server.RemoteSignApiUrl != "" && payServerAddressArgs != "" {
		remoteSignClient, err := sign.NewClient(ctxServer, config.Cfg.Server.RemoteSignApiUrl)
		if err != nil {
			return nil, nil, fmt.Errorf("sign.NewClient err: %s", err.Error())
		}
		handleSign = sign.RemoteSign(remoteSignClient, config.Cfg.Server.Net, payServerAddressArgs)
	} else if config.Cfg.Server.PayPrivate != "" {
		handleSign = sign.LocalSign(config.Cfg.Server.PayPrivate)
	}
	txBuilderBase := txbuilder.NewDasTxBuilderBase(ctxServer, dasCore, handleSign, payServerAddressArgs)
	tool.Log(nil).Info("tx builder ok")

	return txBuilderBase, serverScript, nil
}
