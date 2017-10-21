/**
 * @author liangbo
 * @email  liangbogopher87@gmail.com
 * @date   2017/10/21 21:11 
 */
package main

import (
    "third/gin"
    "time"
    "net/http"
    "pet/protocol"
    "pet/utils"
    "pet/controllers"
)

// 用户电话注册
func UserPhoneRegist(c *gin.Context) {
    var http_code int = http.StatusOK
    handle_start_time := time.Now()

    var args protocol.UserPhoneRegistArgs
    var reply protocol.UserPhoneRegistReply

    r := c.Request
    err := utils.ParseHttpBodyToArgs(r, &args)
    if nil != err {
        goto NOTICE
    }
    err = controllers.UserPhoneRegist(&args, &reply)

NOTICE:
    g_logger.Notice("[cmd:suggest_user_plan][user_id:%s][Cost:%dus][Err:%v]",
        r.FormValue("user_id"), time.Now().Sub(handle_start_time).Nanoseconds()/1000, err)

    utils.SendResponse(c, http_code, &reply, err)
}