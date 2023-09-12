package timer

import (
	"context"
	"das-multi-device/dao"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/dotbitHQ/das-lib/http_api/logger"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

var log = logger.NewLogger("timer", logger.LevelDebug)

type TxTimer struct {
	ctx           context.Context
	wg            *sync.WaitGroup
	dbDao         *dao.DbDao
	dasCore       *core.DasCore
	dasCache      *dascache.DasCache
	txBuilderBase *txbuilder.DasTxBuilderBase
	cron          *cron.Cron
}

type TxTimerParam struct {
	DbDao         *dao.DbDao
	Ctx           context.Context
	Wg            *sync.WaitGroup
	DasCore       *core.DasCore
	DasCache      *dascache.DasCache
	TxBuilderBase *txbuilder.DasTxBuilderBase
}

func NewTxTimer(p TxTimerParam) *TxTimer {
	var t TxTimer
	t.ctx = p.Ctx
	t.wg = p.Wg
	t.dbDao = p.DbDao
	t.dasCore = p.DasCore
	t.dasCache = p.DasCache
	t.txBuilderBase = p.TxBuilderBase
	return &t
}

func (t *TxTimer) Run() error {

	tickerRejected := time.NewTicker(time.Minute * 3)
	t.wg.Add(5)
	go func() {
		defer http_api.RecoverPanic()
		for {
			select {
			case <-tickerRejected.C:
				log.Info("checkRejected start ...")
				if err := t.checkRejected(); err != nil {
					log.Error("checkRejected err: ", err.Error())
				}
				log.Info("checkRejected end ...")

			case <-t.ctx.Done():
				log.Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	return nil
}
