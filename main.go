package main

import (
	"android-remoter/utils"
	"bytes"
	"flag"
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

	go runFrpc()
	go runTerm()
	runAtxAgent()

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
			// log(termlog.String())
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
			// log(frpclog.String())
			log(err.Error())
		}

		log("frpc 10秒后重启")
		time.Sleep(10 * time.Second)
	}

}

func runAtxAgent() {

	log("atx runing")

	cmd := utils.Command{
		Args:       []string{ATXPATHv7 + "/atx-agent", "server", "-d", "--stop"},
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

		c.JSON(200, gin.H{
			"message": frpclog.String(),
		})
	})
	r.GET("/atxlog", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"message": atxlog.String(),
		})
	})
	r.GET("/termlog", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"message": termlog.String(),
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

		output, err := utils.RunShell("remount")
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

	r.GET("/getfrpc", func(c *gin.Context) {
		f, err := getFrpcIni()

		if err != nil {
			c.JSON(200, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"message": f,
		})
	})

	r.GET("/restart-atx", func(c *gin.Context) {

		runAtxAgent()

		c.JSON(200, gin.H{
			"message": atxlog.String(),
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

	if len(logs) > 99 {
		logs = append(logs[1:], msg)
	} else {
		logs = append(logs, msg)
	}

}

func setFrpcIni(serverAddr, serverPort, deviceID, user, pwd string) (err error) {
	cfg := ini.Empty()

	common := cfg.Section("common")
	atx := cfg.Section("atx-" + deviceID)
	term := cfg.Section("term-" + deviceID)
	frpc := cfg.Section("frpc-" + deviceID)
	ctrl := cfg.Section("ctrl-" + deviceID)

	common.Key("server_addr").SetValue(serverAddr)
	common.Key("server_port").SetValue(serverPort)
	common.Key("admin_addr").SetValue("127.0.0.1")
	common.Key("admin_port").SetValue("8200")
	common.Key("admin_user").SetValue(user)
	common.Key("admin_pwd").SetValue(pwd)
	common.Key("deviceID").SetValue(deviceID)

	atx.Key("type").SetValue("http")
	atx.Key("local_port").SetValue("7912")
	atx.Key("http_user").SetValue(user)
	atx.Key("http_pwd").SetValue(pwd)
	atx.Key("subdomain").SetValue("atx-" + deviceID)

	ctrl.Key("type").SetValue("http")
	ctrl.Key("local_port").SetValue("8000")
	ctrl.Key("http_user").SetValue(user)
	ctrl.Key("http_pwd").SetValue(pwd)
	ctrl.Key("subdomain").SetValue("ctrl-" + deviceID)

	term.Key("type").SetValue("http")
	term.Key("local_port").SetValue("8100")
	term.Key("http_user").SetValue(user)
	term.Key("http_pwd").SetValue(pwd)
	term.Key("subdomain").SetValue("term-" + deviceID)

	frpc.Key("type").SetValue("http")
	frpc.Key("local_port").SetValue("8200")
	frpc.Key("subdomain").SetValue("frpc-" + deviceID)

	err = cfg.SaveTo(FRPCPATH + "/frpc.ini")

	return err
}

func getFrpcIni() (frpcini map[string]string, err error) {

	cfg, err := ini.Load(FRPCPATH + "/frpc.ini")
	if err != nil {
		log("Fail to read file: %v", err)
		return
	}

	common := cfg.Section("common")
	frpcini = map[string]string{
		"serverAddr": common.Key("server_addr").Value(),
		"serverPort": common.Key("server_port").Value(),
		"user":       common.Key("admin_user").Value(),
		"pwd":        common.Key("admin_pwd").Value(),
		"deviceID":   common.Key("deviceID").Value(),
	}

	return
}
