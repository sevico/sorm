package sorm

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"sorm/dialect"
	"sorm/log"
	"sorm/session"
)

type Engine struct{
	db *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver,source string) (e *Engine,err error){
	db,err:=sql.Open(driver,source)
	if err!=nil{
		log.Error(err)
		return
	}
	if err=db.Ping();err!=nil{
		log.Error(err)
		return
	}
	dial,ok:=dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return
	}

	e = &Engine{db:db,dialect: dial}
	log.Info("connect database success")
	return
}
func (e *Engine) Close() {
	if err:=e.db.Close();err!=nil{
		log.Error("Failed to close database")
		return
	}
	log.Info("Close database success")
}
func (e *Engine) NewSession() *session.Session {
	return session.New(e.db,e.dialect)
}

type TxFunc func(*session.Session) (interface{},error)

func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s:=engine.NewSession()
	if err = s.Begin();err!=nil{
		return nil,err
	}
	defer func() {
		if p:=recover();p!=nil{
			_=s.Rollback()
			panic(p)
		}else if err!=nil{
			_=s.Rollback()
		}else{
			err=s.Commit()
		}
	}()
	return f(s)
}