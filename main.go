/**
 * @author liangbo
 * @email  liangbogopher87@gmail.com
 * @date   2017/9/24 19:20 
 */
package main

import (
    "third/go-logging"
    "flag"
    "runtime"
    "fmt"
    "pet/utils"
    "pet/models"
)

const (
    SERVERNAME = "pet"
)

var g_conf_file string
var g_config utils.Configure

var g_logger = logging.MustGetLogger(SERVERNAME)

var g_actor_type string = "boss"

func init() {
    const usage = "pet [-c config_file][-a actor_type]"
    flag.StringVar(&g_conf_file, "c", "", usage)
    flag.StringVar(&g_actor_type, "a", "", usage)
}

func main() {
    //set runtime variable
    runtime.GOMAXPROCS(runtime.NumCPU())
    //get flag
    flag.Parse()

    err := utils.InitConfigFileEtcd(SERVERNAME, g_conf_file, &g_config)
    if err != nil {
        fmt.Printf("init config failed, err: %v \n", err)
        return
    }

    // init log file
    g_logger, err = utils.InitLogger(g_config.LogFile)
    if err != nil {
        fmt.Printf("init log failed, err: %v \n", err)
        return
    }

    // init db
    err = models.InitAllDB()
    if err != nil {
        fmt.Printf("init db failed, err: %v", err)
        return
    }

    StartHttpServer()
}