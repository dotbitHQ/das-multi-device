package task

import (
	"context"
	"das-multi-device/dao"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/scorpiotzh/mylog"
	"sync"
)

var log = mylog.NewLogger("parser", mylog.LevelDebug)

type Task struct {
	Ctx      context.Context
	Wg       *sync.WaitGroup
	DbDao    *dao.DbDao
	CkbCli   rpc.Client
	MaxRetry int
}

// smt_status,tx_status: (2,1)->(3,3)
func (t *Task) Run() {

}
