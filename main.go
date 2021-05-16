package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type ConfigFileStruct struct {
	Port 				int				`json:"port"`
	Cmd					CmdRoute		`json:"cmd_config"`
}
type CmdRoute struct {
	Web_url 	string 			`json:"web_url"`
	Socket_url 	string 			`json:"socket_url"`
	Cmds		[]CmdConfig 	`json:"cmds"`
	Exit		string 			`json:"exit_string"`
}
type CmdConfig struct {
	Alias	string `json:"exec_alias"`
	Cmd		string `json:"cmd"`
	Param	string `json:"param"`
}

type SocketSafeWriter struct {
	Conn 	*websocket.Conn
	mu 		sync.Mutex
}

func (sw *SocketSafeWriter) SendMsg(msg []byte)  {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.Conn.WriteMessage(websocket.TextMessage,msg)
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var UserConfig = ConfigFileStruct{}

func main() {
	readConfig()
	ginSetup()
}
func ginSetup()  {
	gin.SetMode(gin.ReleaseMode)

	route := gin.Default()
	//pprof.Register(route,"debug/pprof")

	port := 8888
	configRoute(route)

	if &UserConfig != nil && UserConfig.Port > 0 {
		port = UserConfig.Port
	}
	portStr := ":"+strconv.Itoa(port)
	log.Println("Server Open on Port ",port)
	err := route.Run(portStr)
	if err != nil {
		log.Println("Server Open Failed :",err)
	}
}

func configRoute(route *gin.Engine) {

	//play routes config
	if &UserConfig == nil {
		log.Println("Read Route Config Failed")
		return;
	}
	log.Println("Read Route Config Success")

	//config cmd websocket route
	configRoute_cmd_websocket(route)
}

func configRoute_cmd_websocket(route gin.IRouter){
	if len(UserConfig.Cmd.Web_url) > 0 {
		//download web page path
		route.GET(UserConfig.Cmd.Web_url, func(c *gin.Context) {

			cmdInfoPath := struct {
				Socket_Path string
			}{"ws://"+c.Request.Host+UserConfig.Cmd.Socket_url}

			tmpl, err := template.New("socket_page").Parse(CmdSocketHTMLPage)

			if err != nil {
				log.Println("Cmd Socket Web Page Tmpl Err ", err);
				c.String(http.StatusNotFound,"something is wrong in this page")
			}else {
				c.Status(200)
				tmpl.Execute(c.Writer,cmdInfoPath)
			}
		})

		log.Println("Build Cmd Socket")
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
}

func readConfigFile(filePath string) *ConfigFileStruct {

	res := ConfigFileStruct{}

	fileBlob,fileErr := ioutil.ReadFile(filePath)

	if fileErr != nil {
		log.Println("Config Read Error : ",fileErr)
		return nil
	}

	json.Unmarshal(fileBlob,&res)
	return &res
}

func cmdWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade: %+v", err)
		return
	}
	safeSocket := SocketSafeWriter{Conn: conn}

	remoteChan := make(chan string);
	cmdChan := make(chan bool)
	getCmd := false
	defer close(remoteChan)
	defer close(cmdChan)
	for {
		_,msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		select {
		case c:= <-cmdChan:
			getCmd = c
		default:
		}
		if len(string(msg)) > 0 {
			if getCmd {
				remoteChan <- string(msg)
				continue
			}
			fmt.Println("recieve ",string(msg))
			if cmd := seekCmdFromConfig(string(msg));cmd!=nil {
				excmd(*cmd,&safeSocket,remoteChan,&cmdChan)
				safeSocket.SendMsg([]byte("cmd open success,put any key go on"))
			} else {
				safeSocket.SendMsg( msg)
			}
		}
	}

	cmdChan = nil
	remoteChan = nil
}

func seekCmdFromConfig(alias string) *CmdConfig {
	for _,config := range UserConfig.Cmd.Cmds {
		if config.Alias == alias {
			return &config
		}
	}
	return nil
}

func writeInputToCmd(cmd *exec.Cmd,file *io.WriteCloser,stringChan chan string)  {

	for puts := range stringChan{
		fmt.Println(">>> ",puts)
		if puts == UserConfig.Cmd.Exit {
			err := cmd.Process.Kill()
			if err!=nil {
				fmt.Println("kill process error :",err)
			}
			break;
		}
		//string read from websocket does not contain '\n'
		wNum,err := io.WriteString(*file, puts + "\n")
		fmt.Println("write length ",wNum)
		if err!=nil {
			fmt.Println("write failed : ",err)
			break;
		}
	}
	//check cmd if still running
	fmt.Println("cmd process ",cmd.ProcessState)
	if cmd.ProcessState == nil {
		fmt.Println("cmd still running,close it")
		cmd.Process.Kill()
	}
	fmt.Println("write loop closed ")
}
func readOutputFromCmd(file *io.PipeReader, safeSocket *SocketSafeWriter)  {
	fmt.Println("output pipe file :",*file)

	//bufio.NewScanner is more easier than bufio.NewReader(),and won't cause loop leak
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		fmt.Println("read cmd script msg : ",str)
		if len(str) > 0 {
			safeSocket.SendMsg([]byte(str))
		}
	}
	fmt.Println("read loop closed ")
}
func startCmdScripts(config CmdConfig,safeSocket *SocketSafeWriter, inputChan chan string, errChan chan string, okChan *chan bool)  {
	*okChan <- true

	cmd := exec.Command(config.Cmd,config.Param)
	pf,perr := cmd.StdinPipe()
	if perr != nil {
		fmt.Println("pipe err ",perr)
	}
	defer pf.Close()

	go writeInputToCmd(cmd,&pf,inputChan)

	r,w := io.Pipe()
	cmd.Stdout = w
	defer r.Close()
	defer w.Close()
	go readOutputFromCmd(r,safeSocket)


	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err!=nil {
		log.Println("cmd failed ",err)
	}

	safeSocket.SendMsg([]byte("cmd socket finished"))
	fmt.Println("cmd exit")
	//without goroutine, okchan will stuck this routine until next okchan read
	go func() {
		*okChan <- false
	}()
}
func excmd(config CmdConfig, safeSocket *SocketSafeWriter, cmdChan chan string, okChan *chan bool) {
	go startCmdScripts(config,safeSocket,cmdChan,nil,okChan)
}