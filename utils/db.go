/**
 * @author liangbo
 * @email  liangbogopher87@gmail.com
 * @date   2017/10/8 21:11 
 */
package utils

import (
    "database/sql"
    "fmt"
    "third/gorm"
)

func InitDbPool(config *MysqlConfig) (*sql.DB, error) {

    dbPool, err := sql.Open("mysql", config.MysqlConn)
    if nil != err {
        Logger.Error("sql.Open error :%s,%v", config.MysqlConn, err)
        return nil, err
    }
    dbPool.SetMaxOpenConns(config.MysqlConnectPoolSize)
    dbPool.SetMaxIdleConns(config.MysqlConnectPoolSize)

    err = dbPool.Ping()
    if err != nil {
        return nil, err
    }

    fmt.Println("init db pool OK")
    return dbPool, nil
}

func InitGormDbPool(config *MysqlConfig) (*gorm.DB, error) {

    db, err := gorm.Open("mysql", config.MysqlConn)
    if err != nil {
        fmt.Println("init db err : ", config, err)
        return nil, err
    }

    db.DB().SetMaxOpenConns(config.MysqlConnectPoolSize)
    db.DB().SetMaxIdleConns(config.MysqlConnectPoolSize)
    db.LogMode(true)
    //	db.SetLogger(Logger)
    db.SetLogger(Logger)
    db.SingularTable(true)

    err = db.DB().Ping()
    if err != nil {
        Logger.Error("ping db error :%v", err)
        return nil, err
    }
    fmt.Println("init db pool OK")

    return &db, nil
}

