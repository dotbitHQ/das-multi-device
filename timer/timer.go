package timer

import (
	"context"
	"das-multi-device/dao"
	"das-multi-device/tool"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/dascache"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

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
		for {
			select {
			case <-tickerRejected.C:
				tool.Log(nil).Info("checkRejected start ...")
				if err := t.checkRejected(); err != nil {
					tool.Log(nil).Error("checkRejected err: ", err.Error())
				}
				tool.Log(nil).Info("checkRejected end ...")

			case <-t.ctx.Done():
				tool.Log(nil).Info("timer done")
				t.wg.Done()
				return
			}
		}
	}()

	return nil
}
