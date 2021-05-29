package main

import (
	"encoding/json"
	"flag"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var generateConfig = flag.String("generate","",
	"generate sample config.json file, Example '--generate config.json' will generate config file in current directory ")
var signalFlag = flag.String("signal","","reload")

type ConfigFileStruct struct {
	Port 				int				`json:"port"`
	Cmd					CmdRoute		`json:"cmd_config"`
}

var UserConfig = ConfigFileStruct{}

type OUTPUT_MODE int
const (
	ModeDebug OUTPUT_MODE = iota
	ModeRelease
)

var GO_OUTPUT_MODE OUTPUT_MODE
var logFile = "go_log.txt"

func main() {
	GO_OUTPUT_MODE = ModeRelease

	checkFlag()

	//pid listen module
	setupPidModule()
	logConfig()

	readConfig()
	ginSetup()
}

func checkFlag()  {
	flag.Parse()

	if len(*generateConfig) > 0 {
		defer os.Exit(0)
		configPath := *generateConfig

		file, err := os.Create(configPath)
		if err != nil {
			Error.Println("create config file : ", configPath,"error : ",err)
			return
		}
		defer file.Close()

		_,err = io.WriteString(file,getSampleConfigJsonString())
		if err != nil {
			Error.Println("write config file error : ",err)
			return
		}
		Info.Println("Finished! ",configPath," file successfully generated.")
	}
	if len(*signalFlag) > 0 {
		if *signalFlag == "reload" {
			sendSignal()
			os.Exit(0)
		}
	}
}

func ginSetup()  {

	if GO_OUTPUT_MODE == ModeDebug {
		gin.SetMode(gin.DebugMode)
	}else {
		gin.SetMode(gin.ReleaseMode)
	}

	route := gin.Default()
	//pprof.Register(route,"debug/pprof")

	port := 8888
	configRoute(route)

	if &UserConfig != nil && UserConfig.Port > 0 {
		port = UserConfig.Port
	}
	portStr := ":"+strconv.Itoa(port)
	Info.Println("Server Open on Port ",port)
	err := route.Run(portStr)
	if err != nil {
		Error.Println("Server Open Failed :",err)
	}
}

func configRoute(route *gin.Engine) {

	//play routes config
	if &UserConfig == nil {
		Error.Println("Read Route Config Failed")
		return;
	}
	Info.Println("Read Route Config Success")

	//config cmd websocket route
	configRoute_cmd_websocket(route)
}

func configRoute_cmd_websocket(route gin.IRouter){
	if len(UserConfig.Cmd.Web_url) > 0 {
		//download web page path
		route.GET(UserConfig.Cmd.Web_url, func(c *gin.Context) {

			cmdInfoPath := struct {
				Socket_Path string
			}{UserConfig.Cmd.Socket_url}

			tmpl, err := template.New("socket_page").Parse(CmdSocketHTMLPage)

			if err != nil {
				Error.Println("Cmd Socket Web Page Tmpl Err ", err);
				c.String(http.StatusNotFound,"something is wrong in this page")
			}else {
				c.Status(200)
				tmpl.Execute(c.Writer,cmdInfoPath)
			}
		})

		Info.Println("Build Cmd Socket")
		route.GET(UserConfig.Cmd.Socket_url, func(c *gin.Context) {
			cmdWebSocketHandler(c.Writer,c.Request)
		})
	}
}

func readConfig() {
	filePossiblePath := "config.json"
	UserConfig = *readConfigFile(filePossiblePath)
	if &UserConfig == nil {
		filePossiblePath = "config/config.json"
		UserConfig = *readConfigFile(filePossiblePath)
	}

	//serailize cmd config
	serializeCmds(UserConfig.Cmd)
}

func readConfigFile(filePath string) *ConfigFileStruct {

	res := ConfigFileStruct{}

	fileBlob,fileErr := ioutil.ReadFile(filePath)

	if fileErr != nil {
		Error.Println("Config Read Error : ",fileErr)
		return nil
	}

	json.Unmarshal(fileBlob,&res)
	return &res
}

func logConfig()  {
	f,err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Error.Println("Open Log File Failed : ",err);
		return;
	}
	w := io.MultiWriter(os.Stdout,f)
	gin.DefaultWriter = w

	if GO_OUTPUT_MODE == ModeDebug {
		SetLogMode(DebugMode)
	}else {
		SetLogMode(ReleaseMode)
	}

	SetWriter(w)

	Info.Println("Log File Path : ",filepath.Dir(os.Args[0]) + "/" + logFile)
}

func getSampleConfigJsonString() string {
	return `
{
  "port":18080,
  "Cmd_config": {
    "web_url": "/cmdsocket",
    "socket_url": "/ws",
	"cmds_file": "cmdsFile",
    "cmds": [
      {
        "exec_alias": "pinggithub",
        "cmd": "ping",
        "param": "github.com"
      }
    ],
    "exit_string": "exit"
  }
}



`
}