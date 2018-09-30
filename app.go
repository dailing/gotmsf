package gotmsf

import (
	"encoding/json"
	"github.com/dailing/gotmsf/util"
	"github.com/dailing/levlog"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"reflect"
	"time"
)

type App struct {
	mux        *http.ServeMux
	listenAddr string
}

func (app *App) Run() error {
	server := &http.Server{Handler: app.mux}
	l, err := net.Listen("tcp4", app.listenAddr)
	if err != nil {
		levlog.Fatal(err)
	}
	levlog.Info("starting app at", app.listenAddr)
	return server.Serve(l)
}

const (
	RequestKeyBody      = "__req_body__"
	RequestKeyObj       = "__req_obj__"
	RequestKeyJson      = "__req_json__"
	RequestKeyToken     = "__req_token__"
	RequestKeyHeader    = "__req_header__"
	RequestKeyRawReqObj = "__req_raw_req__"
	ResponseKeyHeader   = "__resp_header__"
	ResponseKeyRawDara  = "__payload__"
)

const (
	RESPONSE_TYPE_JSON = iota
	RESPONSE_TYPE_ERROR
	RESPONSE_TYPE_RAW
	RESPONSE_TYPE_MIDWARE
)

type JsonRespReqFunc func(*util.JsonType) (*util.JsonType, int)

func HandleJsonReq(funcList ...JsonRespReqFunc) http.HandlerFunc {
	var err error
	return func(w http.ResponseWriter, r *http.Request) {
		j := util.NewJson()
		levlog.Trace("vvvvvvvvvvvvvvvvvvvvvREQvvvvvvvvvvvvvvvvvvvvvv")
		levlog.Trace("HOST:", r.Host)
		levlog.Trace("URL:", r.URL.Path)
		levlog.Trace("METHOD:", r.Method)
		levlog.Trace("----------------------------------------------")
		j.Set(RequestKeyToken, r.Header.Get("token"))
		j.Set(RequestKeyHeader, r.Header)
		j.Set(RequestKeyRawReqObj, r)
		j.Set(ResponseKeyHeader, &w)
		j.Set("__req_url__", r.URL.Path)
		var lastRet *util.JsonType
		var succ int
		for _, f := range funcList {
			lastRet, succ = f(j)
			if succ == RESPONSE_TYPE_MIDWARE {
				continue
			} else if succ == RESPONSE_TYPE_ERROR {
				w.WriteHeader(lastRet.GetInt("code"))
				_, err := w.Write(lastRet.GetBytes("info"))
				levlog.E(err)
				return
			}
		}
		var payload []byte
		w.Header()
		if succ == RESPONSE_TYPE_RAW {
			payload = lastRet.GetBytes(ResponseKeyRawDara)
		} else if succ == RESPONSE_TYPE_JSON {
			payload, err = json.Marshal(lastRet)
			levlog.E(err)
			if err != nil {
				w.WriteHeader(500)
				_, err := w.Write([]byte(err.Error()))
				levlog.E(err)
			}
		}
		w.WriteHeader(200)
		_, err := w.Write(payload)
		levlog.E(err)
		levlog.Trace("^^^^^^^^^^^^^^^^^^^^^END^^^^^^^^^^^^^^^^^^^^^^")
	}
}

func (app *App) Handle(url string, funcList ...JsonRespReqFunc) {
	app.mux.HandleFunc(url, HandleJsonReq(funcList...))
}

func (app *App) HandleStatics(prefix, path string) {
	app.mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(path))))
}

func NewWebApp(addr string) *App {
	return &App{
		mux:        http.NewServeMux(),
		listenAddr: addr,
	}
}

func ResponseError(code int, info string) (*util.JsonType, int) {
	j := util.NewJson()
	j.Set("code", code)
	j.Set("info", info)
	return j, RESPONSE_TYPE_ERROR
}

func ResponseJson(j *util.JsonType) (*util.JsonType, int) {
	return j, RESPONSE_TYPE_JSON
}

func ResponseAnyToJson(any interface{}) (*util.JsonType, int) {
	payload, err := json.Marshal(any)
	levlog.E(err)
	return ResponseRaw(string(payload))
}

func ResponseRaw(s string) (*util.JsonType, int) {
	j := util.NewJson()
	j.Set(ResponseKeyRawDara, s)
	return j, RESPONSE_TYPE_RAW
}
func ResponseSucc() (*util.JsonType, int) {
	return ResponseRaw("succ")
}
func NoResponse() (*util.JsonType, int) {
	return nil, RESPONSE_TYPE_MIDWARE
}

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func randStringGen(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// request related functions
func GetReqRaw(j *util.JsonType) *http.Request {
	return j.GetObj(RequestKeyRawReqObj).(*http.Request)
}

func ReadBody(jsonType *util.JsonType) (*util.JsonType, int) {
	r := GetReqRaw(jsonType)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		levlog.E(err)
		ResponseError(500, err.Error())
	}
	if len(body) < 2048 {
		levlog.Trace("REQ BODY:", string(body))
	}
	jsonType.Set(RequestKeyBody, string(body))
	return NoResponse()
}

func BodyToJson(jsonType *util.JsonType) (*util.JsonType, int) {
	j := util.NewJson()
	err := json.Unmarshal(jsonType.GetBytes(RequestKeyBody), j)
	if err != nil {
		levlog.E(err)
		return ResponseError(403, "Request Data Error")
	}
	jsonType.Set(RequestKeyJson, j)
	return NoResponse()
}

func BodyToObj(obj interface{}) func(jsonType *util.JsonType) (*util.JsonType, int) {
	return func(jsonType *util.JsonType) (*util.JsonType, int) {
		j := reflect.New(reflect.TypeOf(obj).Elem()).Interface()
		err := json.Unmarshal(jsonType.GetBytes(RequestKeyBody), j)
		if err != nil {
			levlog.E(err)
			return ResponseError(403, "Request Data Error")
		}
		jsonType.Set(RequestKeyObj, j)
		return NoResponse()
	}
}

func GetReqObj(j *util.JsonType) interface{} {
	return j.GetObj(RequestKeyObj)
}
