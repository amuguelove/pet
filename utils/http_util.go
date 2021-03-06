/**
 * @author liangbo
 * @email  liangbogopher87@gmail.com
 * @date   2017/9/24 21:14
 */
package utils

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "reflect"
    "regexp"
    "strconv"
    "strings"
    "time"

    "third/gin"
    "third/go-local"
    "third/http_client_cluster"
    "runtime"
    "third/mapstructure"
)

const DEFAULT_API_TIMEOUT = 1 * time.Second

type ApiResponse struct {
    Status string      `json:"status"`
    Data   interface{} `json:"data"`
    Desc   string      `json:"desc"`
}

func (this ApiResponse) MarshalJSON() ([]byte, error) {
    return json.Marshal(map[string]interface{}{
        "status": this.Status,
        "data":   this.Data,
        "desc":   this.Desc,
    })
}

func CodoonGetHeader(c *gin.Context) {
    // 获取token
    r := c.Request
    Logger.Info("+++++++++++request header: %+v", r.Header)
}

var sliceOfInts = reflect.TypeOf([]int(nil))
var sliceOfStrings = reflect.TypeOf([]string(nil))

// parse form values to struct via tag.
func ParseForm(form url.Values, obj interface{}) error {
    objT := reflect.TypeOf(obj)
    objV := reflect.ValueOf(obj)
    if !IsStructPtr(objT) {
        return fmt.Errorf("%v must be  a struct pointer", obj)
    }
    objT = objT.Elem()
    objV = objV.Elem()

    for i := 0; i < objT.NumField(); i++ {
        fieldV := objV.Field(i)
        if !fieldV.CanSet() {
            continue
        }

        fieldT := objT.Field(i)
        tags := strings.Split(fieldT.Tag.Get("form"), ",")
        var tag string
        if len(tags) == 0 || len(tags[0]) == 0 {
            tag = fieldT.Name
        } else if tags[0] == "-" {
            continue
        } else {
            tag = tags[0]
        }

        value := form.Get(tag)
        if len(value) == 0 {
            continue
        }

        switch fieldT.Type.Kind() {
        case reflect.Bool:
            if strings.ToLower(value) == "on" || strings.ToLower(value) == "1" || strings.ToLower(value) == "yes" {
                fieldV.SetBool(true)
                continue
            }
            if strings.ToLower(value) == "off" || strings.ToLower(value) == "0" || strings.ToLower(value) == "no" {
                fieldV.SetBool(false)
                continue
            }
            b, err := strconv.ParseBool(value)
            if err != nil {
                return err
            }
            fieldV.SetBool(b)
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
            x, err := strconv.ParseInt(value, 10, 64)
            if err != nil {
                return err
            }
            fieldV.SetInt(x)
        case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
            x, err := strconv.ParseUint(value, 10, 64)
            if err != nil {
                return err
            }
            fieldV.SetUint(x)
        case reflect.Float32, reflect.Float64:
            x, err := strconv.ParseFloat(value, 64)
            if err != nil {
                return err
            }
            fieldV.SetFloat(x)
        case reflect.Interface:
            fieldV.Set(reflect.ValueOf(value))
        case reflect.String:
            fieldV.SetString(value)
        case reflect.Struct:
            switch fieldT.Type.String() {
            case "time.Time":
                format := time.RFC3339
                if len(tags) > 1 {
                    format = tags[1]
                }
                t, err := time.Parse(format, value)
                if err != nil {
                    return err
                }
                fieldV.Set(reflect.ValueOf(t))
            }
        case reflect.Slice:
            if fieldT.Type == sliceOfInts {
                formVals := form[tag]
                fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(int(1))), len(formVals), len(formVals)))
                for i := 0; i < len(formVals); i++ {
                    val, err := strconv.Atoi(formVals[i])
                    if err != nil {
                        return err
                    }
                    fieldV.Index(i).SetInt(int64(val))
                }
            } else if fieldT.Type == sliceOfStrings {
                formVals := form[tag]
                fieldV.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf("")), len(formVals), len(formVals)))
                for i := 0; i < len(formVals); i++ {
                    fieldV.Index(i).SetString(formVals[i])
                }
            }
        }
    }
    return nil
}

// result must be pointer
func ParseHttpBodyToArgs(c *gin.Context, result interface{}) error {
    var err error
    var args_map map[string]interface{} = make(map[string]interface{})

    r := c.Request
    if nil != c.Request.Body {
        decoder := json.NewDecoder(r.Body)
        decoder.UseNumber()
        decoder.Decode(&args_map)

    }
    c.Request.Body.Close()
    for _, param := range c.Params {
        if _, ok := args_map[param.Key]; !ok {
            value_int, err := strconv.ParseInt(param.Value, 10, 0)
            if err != nil || 1 == string_key[param.Key] {
                args_map[param.Key] = param.Value
            } else {
                args_map[param.Key] = value_int
            }
        }
    }

    r.ParseForm()
    for key, value := range r.Form {
        if _, ok := args_map[key]; !ok {
            value_int, err := strconv.ParseInt(value[0], 10, 0)
            if err != nil || 1 == string_key[key] {
                args_map[key] = value[0]
            } else {
                args_map[key] = value_int
            }
        }
    }

    Logger.Info("parse body map args: %+v", args_map)
    err = mapstructure.Decode(args_map, result)
    return err
}

func OptionHandler(c *gin.Context) {}
func GinCrossDomain() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Cache-Control", "no-cache")
        CheckCrossdomain(c)
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusOK)
        }
    }
}

func GinFilter() gin.HandlerFunc {
    return func(c *gin.Context) {

        if c.Request.Method == "HEAD" {
            c.AbortWithStatus(http.StatusOK)
        }
    }
}

func CheckCrossdomain(c *gin.Context) {
    c.Writer.Header().Add("Access-Control-Allow-Headers", "content-type, authorization")
    c.Writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
    c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")

    origin_list := []string{"lhd.codoon.com", ".codoon.com$", ".runtopia.net$", ".blastapp.net$", "http://localhost", "192.168.\\d+"}
    for _, str := range origin_list {
        reg := regexp.MustCompile(str)
        if reg.MatchString(c.Request.Header.Get("Origin")) {
            fmt.Println("match origin: ", c.Request.Header.Get("Origin"))
            c.Writer.Header().Add("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
            break
        }
    }
}

var htmlEscape bool = true // default is true in json.Marshal

func SetHtmlEscape(b bool) {
    htmlEscape = b
}

func HTMLUnEscape(src []byte) []byte {
    src = bytes.Replace(src, []byte("\\u003c"), []byte("<"), -1)
    src = bytes.Replace(src, []byte("\\u003e"), []byte(">"), -1)
    src = bytes.Replace(src, []byte("\\u0026"), []byte("&"), -1)
    return src
}

func SendResponse(c *gin.Context, http_code int, data interface{}, err error) error {
    var resp ApiResponse = ApiResponse{"OK", nil, "success"}
    if err != nil {
        is_user_err, code, info := IsUserErr(err)
        if is_user_err {
            resp.Status = "Error"
            resp.Data = code
            resp.Desc = info
        } else {
            http_code = 500
            c.String(http_code, http.StatusText(http_code))
            return nil
        }
    } else {
        resp.Data = data
    }
    c.Writer.Header().Set("Content-Type", "application/json")
    c.Writer.Header().Set("ServerTime", strconv.FormatInt(time.Now().Unix(), 10))

    b, err := json.Marshal(&resp)
    if err != nil {
        Logger.Error("Marshal json to bytes error :%v", err)
    }

    c.Writer.Header().Del("Content-length")
    if 0 != len(b) {
    }

    if !htmlEscape {
        b = HTMLUnEscape(b)
    }

    // 输出结果，当大于3000个字符不在输出
    if len(b) > 3000 {
        //Logger.Info(string(b[:3000]))
    } else {
        Logger.Info("response: %s", string(b))
    }

    // 定死的变量不需要输出
    //Logger.Info("+++++++++++response header: %v ", c.Writer.Header())
    c.Writer.Write(b)

    return err
}

// modified by linagbo on 2017-08-10, 定义不被转化成int的key
var string_key map[string]int = map[string]int{

}

//func ForwardHttpToRpc(c *gin.Context, client *RpcClient, method string, args map[string]interface{}, reply interface{}, http_code *int) error {
//
//    r := c.Request
//    if nil != c.Request.Body {
//        decoder := json.NewDecoder(r.Body)
//        decoder.UseNumber()
//        decoder.Decode(&args)
//
//    }
//    c.Request.Body.Close()
//    for _, param := range c.Params {
//        if _, ok := args[param.Key]; !ok {
//            value_int, err := strconv.ParseInt(param.Value, 10, 0)
//            if err != nil || 1 == string_key[param.Key] {
//                args[param.Key] = param.Value
//            } else {
//                args[param.Key] = value_int
//            }
//        }
//    }
//
//    r.ParseForm()
//    for key, value := range r.Form {
//        if _, ok := args[key]; !ok {
//            value_int, err := strconv.ParseInt(value[0], 10, 0)
//            if err != nil || 1 == string_key[key] {
//                args[key] = value[0]
//            } else {
//                args[key] = value_int
//            }
//        }
//    }
//    args["user_agent"] = r.Header.Get("User-Agent")
//
//    // 日志输出统一都放到client.Call方法里面，重复
//    err := client.Call(method, &args, reply)
//    if nil != err {
//        is_user_err, _, _ := IsUserErr(err)
//        if !is_user_err {
//            *http_code = http.StatusInternalServerError
//        }
//    }
//    return err
//}

func GinRecovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            err := recover()
            if err != nil {
                switch err.(type) {
                case error:
                    //CheckError(err.(error))
                default:
                    err = errors.New(fmt.Sprint(err))
                    //CheckError(err)
                }

                stack := stack(3)
                Logger.Error("PANIC: %s\n%s", err, stack)

                c.Writer.WriteHeader(http.StatusInternalServerError)
            }
        }()

        c.Next()
    }
}

func MyRecovery() {

    err := recover()
    if err != nil {
        switch err.(type) {
        case error:
            //CheckError(err.(error))
        default:
            err = errors.New(fmt.Sprint(err))
            //CheckError(err)
        }

        stack := stack(3)
        Logger.Error("PANIC: %s\n%s", err, stack)
    }

}

func GinLogger() gin.HandlerFunc {

    return func(c *gin.Context) {
        // Start timer
        start := time.Now()

        // Process request
        c.Next()

        // Stop timer
        end := time.Now()
        latency := end.Sub(start)

        clientIP := c.ClientIP()
        method := c.Request.Method
        statusCode := c.Writer.Status()
        Logger.Notice("[GIN] %v | %3d | %12v | %s | %-7s %s %s\n%s\n%s",
            end.Format("2006/01/02 - 15:04:05"),
            statusCode,
            latency,
            clientIP,
            method,
            c.Request.URL.String(),
            c.Request.URL.Opaque,
            c.Errors.String(),
            c.Keys)

        if statusCode == 500 {

        }

        if latency > DEFAULT_API_TIMEOUT {
            Logger.Error("[TIMEOUT] %v | %3d | %12v | %s | %-7s %s %s | %s",
                end.Format("2006/01/02 - 15:04:05"),
                statusCode,
                latency,
                clientIP,
                method,
                c.Request.URL.String(),
                c.Request.URL.Opaque,
                c.Errors.String())

        }
    }
}

var secretTable string = "_WY+Ytpa=A^Fm(Jl-rx@EVLl-Yx$v4+YgOhxB4s$Lqcen+BflOj_lgS3xuh5bSN-Jnhj69OSa(CmV5*91MRh8XIY423aPH_k$-u@XwaMgmPFCL1Ne-dx!kV$Q_US7f7fMV!H2CgjXmk)8aY3ftssyOrL-(c(UcW*QRd^8Fhcfs)A@qmR$8A8TFm8#)CvNE_CZ2lkvgVCC-vZaeDv^jb1QOv@W2+Ph!eQM=CtbtZPz(wX%gY)J$gdC8Rbc1L*(x6%tVO7RUutHAZF#6@sl(LzBP1DAzU7ttpHfqvKN$e5C@c!pg=@c$zL55$kg!8KJ$1SCbMbL^BYKaK9&_yxUU#XZF&GqY_tS!MN$zsWsLX*4uvCVG_EJ3-96qejb3z9m7e)BrmQMlTS9fVkA%5J5OL12BY8pzTJIeWC1z#jQaTwjnEl$cZj(sqY*LkMJG+(7l*ZNuY1rU5Tvcf6NH%5%7P8r&&yIsj=z2z4c=8VL5gelN-ZGOas$xpX8hf-qOK+MO8s"

func HttpRequest(method, url string, data []byte) (status int, body []byte, err error) {
    var data_reader io.Reader = nil
    if len(data) > 0 {
        data_reader = bytes.NewReader(data)
    }

    req, err := http.NewRequest(method, url, data_reader)
    if err != nil {
        return
    }
    local.FillTraceHttp(req)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

    resp, err := HttpClientClusterDo(req)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    body, err = ioutil.ReadAll(resp.Body)
    return resp.StatusCode, body, err
}

func HttpClientClusterDo(request *http.Request) (*http.Response, error) {

    resp, err := http_client_cluster.HttpClientClusterDo(request)
    return resp, err
}

func GetTokenFromHeader(r *http.Request) string {
    var token string
    auth := r.Header.Get("Authorization")
    if "" == auth {
        cookie, err := r.Cookie("sessionid")
        if nil != err || nil == cookie {
            return token
        }
        auth = cookie.Value
        if "" == auth {
            return token
        }
    }
    auths := strings.Split(auth, " ")
    if 2 != len(auths) {
        return token
    }
    if "Bearer" == auths[0] {
        token = auths[1]
    }
    if "" == token {
        return token
    }
    return token
}

func GetGinRawPath(c *gin.Context) string {
    path := c.Request.URL.Path
    for i := range c.Params {
        path = strings.Replace(path, c.Params[i].Value, ":"+c.Params[i].Key, -1)
    }
    fmt.Println("path :%s", path)
    return path
}

// stack returns a nicely formated stack frame, skipping skip frames
func stack(skip int) []byte {
    buf := new(bytes.Buffer) // the returned data
    // As we loop, we open files and read them. These variables record the currently
    // loaded file.
    var lines [][]byte
    var lastFile string
    for i := skip; ; i++ { // Skip the expected number of frames
        pc, file, line, ok := runtime.Caller(i)
        if !ok {
            break
        }
        // Print this much at least.  If we can't find the source, it won't show.
        fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
        if file != lastFile {
            data, err := ioutil.ReadFile(file)
            if err != nil {
                continue
            }
            lines = bytes.Split(data, []byte{'\n'})
            lastFile = file
        }
        fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
    }
    return buf.Bytes()
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
    fn := runtime.FuncForPC(pc)
    if fn == nil {
        return dunno
    }
    name := []byte(fn.Name())
    // The name includes the path name to the package, which is unnecessary
    // since the file name is already included.  Plus, it has center dots.
    // That is, we see
    //	runtime/debug.*T·ptrmethod
    // and want
    //	*T.ptrmethod
    // Also the package path might contains dot (e.g. code.google.com/...),
    // so first eliminate the path prefix
    if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
        name = name[lastslash+1:]
    }
    if period := bytes.Index(name, dot); period >= 0 {
        name = name[period+1:]
    }
    name = bytes.Replace(name, centerDot, dot, -1)
    return name
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
    n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
    if n < 0 || n >= len(lines) {
        return dunno
    }
    return bytes.TrimSpace(lines[n])
}
