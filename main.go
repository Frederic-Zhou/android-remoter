package main

import (
	"android-remoter/utils"
	"bytes"
	"flag"
	"io/ioutil"
	"strings"
	"time"

	"fmt"

	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

var (
	MYNAME     = os.Args[0]
	PATH       = utils.GetCurrentDirectory()
	ASSETSPATH = fmt.Sprintf("%s/%s", PATH, "assets")
	ATXPATH    = fmt.Sprintf("%s/%s", ASSETSPATH, "atx-agent")
	ATXPATH64  = fmt.Sprintf("%s/%s", ATXPATH, "arm64")
	ATXPATHv6  = fmt.Sprintf("%s/%s", ATXPATH, "armv6")
	ATXPATHv7  = fmt.Sprintf("%s/%s", ATXPATH, "armv7")

	FRPCPATH = fmt.Sprintf("%s/%s", ASSETSPATH, "frpc")
	TERMPATH = fmt.Sprintf("%s/%s", ASSETSPATH, "term")

	LOGFILE = fmt.Sprintf("%s.log", MYNAME)
)

var logs []string
var atxlog, termlog, frpclog bytes.Buffer

func main() {
	log("ver:0.0.1")

	setFrpcIni("192.168.3.100", "7000", "test02", "admin", "123")

	go runTerm()
	go runFrpc()
	go runAtxAgent()

	port := flag.String("p", "8000", "address")
	flag.Parse()

	runService(":" + *port)
}

func runTerm() {
	for {
		log("term runing")
		cmd := utils.Command{
			Args:       []string{TERMPATH + "/term", "-p 8100", "bash"},
			Shell:      true,
			ShellQuote: false,
			Timeout:    10 * time.Minute,
		}

		cmd.Stderr = &termlog
		cmd.Stdout = &termlog

		err := cmd.Run()
		if err != nil {
			log(termlog.String())
			log(err.Error())
		}

		log("term 10秒后重启")
		time.Sleep(10 * time.Second)
	}

}

func runFrpc() {

	for {
		log("frpc runing")
		cmd := utils.Command{
			Args:       []string{FRPCPATH + "/frpc", "-c ", FRPCPATH + "/frpc.ini"},
			Shell:      true,
			ShellQuote: false,
			Timeout:    10 * time.Minute,
		}

		cmd.Stderr = &frpclog
		cmd.Stdout = &frpclog

		err := cmd.Run()
		if err != nil {
			log(frpclog.String())
			log(err.Error())
		}

		log("frpc 10秒后重启")
		time.Sleep(10 * time.Second)
	}

}

func runAtxAgent() {
	for {
		log("atx runing")

		cmd := utils.Command{
			Args:       []string{ATXPATHv7 + "/atx-agent", "server"},
			Shell:      true,
			ShellQuote: false,
			Timeout:    10 * time.Minute,
		}

		cmd.Stderr = &atxlog
		cmd.Stdout = &atxlog

		err := cmd.Run()
		if err != nil {
			log(atxlog.String())
			log(err.Error())
		}
		log("atx 10秒后重启")
		time.Sleep(10 * time.Second)
	}

}

func runService(port string) {
	log("server runing")
	r := gin.Default()
	r.GET("/log", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": logs,
		})
	})
	r.GET("/frpclog", func(c *gin.Context) {
		data, _ := ioutil.ReadAll(&frpclog)

		c.JSON(200, gin.H{
			"message": string(data),
		})
	})
	r.GET("/atxlog", func(c *gin.Context) {

		data, _ := ioutil.ReadAll(&atxlog)

		c.JSON(200, gin.H{
			"message": string(data),
		})
	})
	r.GET("/termlog", func(c *gin.Context) {

		data, _ := ioutil.ReadAll(&termlog)

		c.JSON(200, gin.H{
			"message": string(data),
		})
	})

	r.POST("/setfrpc", func(c *gin.Context) {

		input := struct {
			ServerAddr string `form:"serverAddr"  binding:"required"`
			ServerPort string `form:"serverPort"  binding:"required"`
			DevicesID  string `form:"devicesID"  binding:"required"`
			User       string `form:"user"  binding:"required"`
			Pwd        string `form:"pwd"  binding:"required"`
		}{}

		err := c.ShouldBind(&input)
		if err != nil {
			c.JSON(200, gin.H{
				"message": err.Error(),
			})
			return
		}

		output, err := utils.RunShell("mount -o rw,remount /")
		if err != nil {
			c.JSON(200, gin.H{
				"message": string(output) + " " + err.Error(),
			})
			return
		}

		err = setFrpcIni(input.ServerAddr, input.ServerPort, input.DevicesID, input.User, input.Pwd)
		if err != nil {
			c.JSON(200, gin.H{
				"message": err.Error(),
			})
			return
		}

		output, err = utils.RunShell("mount -o ro,remount /")
		if err != nil {
			c.JSON(200, gin.H{
				"message": string(output) + " " + err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "success",
		})
	})
	r.Run(port)
}

func log(args ...interface{}) {
	l := []string{}
	for _, a := range args {
		l = append(l, fmt.Sprintf("%v", a))
	}
	msg := fmt.Sprintf("[%s]:%s", time.Now().Format("2006-01-02 15:04:05"), strings.Join(l, " "))
	logs = append(logs, msg)
}

func setFrpcIni(serverAddr, serverPort, devicesID, user, pwd string) (err error) {
	cfg := ini.Empty()

	common := cfg.Section("common")
	atx := cfg.Section("atx-" + devicesID)
	term := cfg.Section("term-" + devicesID)
	frpc := cfg.Section("frpc-" + devicesID)
	ctrl := cfg.Section("ctrl-" + devicesID)

	common.Key("server_addr").SetValue(serverAddr)
	common.Key("server_port").SetValue(serverPort)
	common.Key("admin_addr").SetValue("127.0.0.1")
	common.Key("admin_port").SetValue("8200")
	common.Key("admin_user").SetValue(user)
	common.Key("admin_pwd").SetValue(pwd)

	atx.Key("type").SetValue("http")
	atx.Key("local_port").SetValue("7912")
	atx.Key("http_user").SetValue(user)
	atx.Key("http_pwd").SetValue(pwd)
	atx.Key("subdomain").SetValue("atx-" + devicesID)

	ctrl.Key("type").SetValue("http")
	ctrl.Key("local_port").SetValue("8000")
	ctrl.Key("http_user").SetValue(user)
	ctrl.Key("http_pwd").SetValue(pwd)
	ctrl.Key("subdomain").SetValue("ctrl-" + devicesID)

	term.Key("type").SetValue("http")
	term.Key("local_port").SetValue("8100")
	term.Key("http_user").SetValue(user)
	term.Key("http_pwd").SetValue(pwd)
	term.Key("subdomain").SetValue("term-" + devicesID)

	frpc.Key("type").SetValue("http")
	frpc.Key("local_port").SetValue("8200")
	frpc.Key("subdomain").SetValue("frpc-" + devicesID)

	err = cfg.SaveTo(FRPCPATH + "/frpc.ini")

	return err
}

//adb root && adb shell mount -o rw,remount / && GOOS=linux GOARCH=arm GOARM=7 go build && adb push ./android-remoter /system/xbin/AR && adb shell reboot
