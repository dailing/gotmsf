package db

import (
	"github.com/dailing/levlog"
	"github.com/go-xorm/xorm"
)

type DbInstance struct {
	Engine *xorm.Engine
}

func NewDB(host, dbName string) *DBDInterface {
	levlog.Info("Init DB with:", host, ":", dbName)
	var (
		e   *xorm.Engine
		err error
	)
	e, err = xorm.NewEngine("mysql",
		fmt.Sprintf("root:123456@tcp(%s)/%s", host, dbName))
	levlog.F(err)

	return e
}

func (d *DbInstance) InitTable(...tables) error {
	for(tab := range(tables)){
		levlog.info(tab)
	}
	levlog.F(err)
	return nil
}
