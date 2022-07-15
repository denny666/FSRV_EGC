package main

import (
	"FSRV_Edge/global"
	_ "FSRV_Edge/init/initdb"
	initlog "FSRV_Edge/init/initlog"
	"FSRV_Edge/nodeattr"
	"FSRV_Edge/router/accountrouter"
	"FSRV_Edge/router/connectrecordrouter"
	"FSRV_Edge/router/connectrouter"
	"FSRV_Edge/router/devicerouter"
	"FSRV_Edge/router/historyrouter"
	"FSRV_Edge/router/kanbanrouter"
	"FSRV_Edge/router/loginrouter"
	"FSRV_Edge/router/templaterouter"
	"FSRV_Edge/service/edgeservice"
	"FSRV_Edge/service/opcuaservice"
	"FSRV_Edge/service/routerservice"
	"flag"
	"net"
	"os"
	"syscall"
	"time"

	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os/signal"
	"strings"

	_ "FSRV_Edge/docs"

	"github.com/aWildProgrammer/fconf"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gopcua/opcua/debug"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	settingPath = "./setting.ini"
)

var (
	globalLogger = initlog.GetLogger()
)

func init() {
	c, err := fconf.NewFileConf(settingPath)
	if err != nil {
		fmt.Println(err)
	}
	global.WriteDBSetting, err = c.Int("Setting.write_db")
	if err != nil {
		fmt.Println(err)
	}
	global.PingWiseSetting, err = c.Int("Setting.ping_wise")
	if err != nil {
		fmt.Println(err)
	}
	global.BuildNodeSetting, err = c.Int("Setting.build_node")
	if err != nil {
		fmt.Println(err)
	}
	global.MainIP = c.String("Setting.main_ip")
	global.WorkShopNumber = c.String("Setting.work_shop_number")
	global.DcTr = &http.Transport{
		DisableKeepAlives:   true,
		MaxIdleConnsPerHost: -1}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			globalLogger.Criticalf("[main error]:%v", r)
		}
	}()
	setupCloseHandler()
	flag.BoolVar(&debug.Enable, "debug", false, "enable debug logging")
	flag.Parse()
	log.SetFlags(0)
	var err error
	opcuaservice.BuildServer(nodeattr.DefaultOPCPort)
	go edgeservice.UpdateConStatus()
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(recoverHandle())
	router.Use(corsMiddleware())
	kanbanrouter.AddKanbanRouter(router)
	templaterouter.AddTemplateRouter(router)
	devicerouter.AddDeviceRouter(router)
	historyrouter.AddHistoryRouter(router)
	connectrouter.AddConnectRouter(router)
	connectrecordrouter.AddConnectRecordRouter(router)
	accountrouter.AddAccountRouter(router)
	loginrouter.AddLoginRouter(router)
	url := ginSwagger.URL("/swagger/doc.json")
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	router.Use(static.Serve("/static", static.LocalFile("./static", true)))
	err = router.Run(nodeattr.DefaultEGPortStr)
	if err != nil {
		globalLogger.Criticalf("Run router error: %s", err)
	}

}
func recoverHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var record nodeattr.ConnectRecord
		defer func() {
			if r := recover(); r != nil {
				globalLogger.Criticalf("[Recover error]: %v", r)
				ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
				record.Type = 0
				record.IP = ip
				record.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
				record.URL = c.Request.RequestURI
				routerservice.InsertConRecordInfo(record)
			}
		}()
		c.Next()
	}
}

// corsMiddleware 允許cors請求
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		// 核心处理方式
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
		c.Set("content-type", "application/json")
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		c.Next()
	}
}

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

func getDcIP(macStr string) ([]string, error) {
	var ipArr []string
	args := []string{"arp-scan", "-l"}
	cmd := exec.Command("sudo", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ipArr, err
	}
	a := strings.Split(string(out), "\t")
	for k, v := range a {
		if strings.Contains(v, macStr) && k > 0 {
			c := strings.Split(a[k-1], "\n")
			if len(c) > 1 {
				ipArr = append(ipArr, c[1])
			}
		}
	}
	return ipArr, nil
}
