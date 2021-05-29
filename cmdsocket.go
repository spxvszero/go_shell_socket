package main

import (
	"bufio"
	"context"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type CmdRoute struct {
	Web_url 	string 			`json:"web_url"`
	Socket_url 	string 			`json:"socket_url"`
	Cmds		[]CmdConfig 	`json:"cmds"`
	CmdsFile	string			`json:"cmds_file"`
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
	if sw.Conn != nil {
		sw.Conn.WriteMessage(websocket.TextMessage,msg)
	}
}

var cmdsocket_log_prefix = "CMD Socket : "

var cmds_map map[string]CmdConfig

func serializeCmds(route CmdRoute)  {
	cmds_map = make(map[string]CmdConfig)
	//scan file
	if len(route.CmdsFile) > 0 {
		file,err := os.Open(route.CmdsFile)
		if err != nil {
			Error.Println(cmdsocket_log_prefix,"Open cmdsfile error : ",err)
		}else {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				str := scanner.Text()
				splitIdx := strings.Index(str,":")
				if splitIdx > 0 {
					alias := str[0:splitIdx]
					blankSplitIdx := strings.Index(str," ")
					if blankSplitIdx > 0{
						cmdStr := str[splitIdx+1:blankSplitIdx]
						paramStr := str[blankSplitIdx+1:]
						cmds_map[alias] = CmdConfig{Alias: alias,Cmd: cmdStr,Param: paramStr}
						logCmdsocket("Add cmd ",alias," ",cmdStr," ",paramStr)
					}else {
						cmdStr := str[splitIdx+1:]
						if len(cmdStr) > 0 {
							cmds_map[alias] = CmdConfig{Alias: alias,Cmd: cmdStr,Param: ""}
							logCmdsocket("Add cmd ",alias," ",cmdStr)
						}
					}
				}
			}
		}
	}
	//config file
	for _,cmd := range route.Cmds {
		cmds_map[cmd.Alias] = cmd
		logCmdsocket("Add cmd ",cmd.Alias," ",cmd)
	}
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
func cmdWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		Error.Println(cmdsocket_log_prefix,"Failed to set websocket upgrade: %+v", err)
		return
	}
	safeSocket := SocketSafeWriter{Conn: conn}

	remoteChan := make(chan string);
	okChan := make(chan bool)
	getCmd := false
	ctx,cancel := context.WithCancel(context.Background())
	defer close(remoteChan)
	defer close(okChan)
	for {
		_,msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		select {
		case c:= <-okChan:
			getCmd = c
		default:
		}
		if len(string(msg)) > 0 {
			//if reading cmd, socket string will gothrought cmd interactive without make a new cmd.
			if getCmd {
				remoteChan <- string(msg)
				continue
			}
			logCmdsocket("recieve ",string(msg))
			//get str
			if cmd := seekCmdFromConfig(string(msg));cmd!=nil {
				excmd(ctx,*cmd,&safeSocket,remoteChan,&okChan)
				safeSocket.SendMsg([]byte("CMD open success, put any key go on"))
			} else {
				safeSocket.SendMsg( msg)
			}
		}
	}

	//read last okchan which still waiting to receive before closed
	select {
	case c:= <-okChan:
		logCmdsocket("no use ",c)
	default:
	}
	okChan = nil
	remoteChan = nil
	safeSocket.Conn = nil
	cancel()
	logCmdsocket("Socket Handle Closed !")
}

func seekCmdFromConfig(alias string) *CmdConfig {
	cmd := cmds_map[alias]
	if len(cmd.Alias) > 0 {
		return &cmd
	}
	return nil
}

func writeInputToCmd(cmd *exec.Cmd,file *io.WriteCloser,stringChan chan string)  {

	for puts := range stringChan{
		logCmdsocket(">>> ",puts)
		if puts == UserConfig.Cmd.Exit {
			err := cmd.Process.Kill()
			if err!=nil {
				logCmdsocket("kill process error :",err)
			}
			break;
		}
		wNum,err := io.WriteString(*file, puts + "\n")
		logCmdsocket("write length ",wNum)
		if err!=nil {
			logCmdsocket("write failed : ",err)
			break;
		}
	}
	//check cmd if still running
	logCmdsocket("cmd process ",cmd.ProcessState)
	if cmd.ProcessState == nil && cmd.Process != nil {
		logCmdsocket("cmd still running, close it")
		cmd.Process.Kill()
	}
	logCmdsocket("goroutine write loop closed ")
}
func readOutputFromCmd(cmd *exec.Cmd,file *io.PipeReader, safeSocket *SocketSafeWriter)  {
	logCmdsocket("output pipe file :",*file)

	//bufio.NewScanner is more easier than bufio.NewReader(),and won't cause loop leak
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		logCmdsocket("read cmd script msg : ",str)
		if len(str) > 0 {
			if safeSocket.Conn == nil {
				break
			}
			safeSocket.SendMsg([]byte(str))
			logCmdsocket("socket : ",safeSocket.Conn)
		}
	}
	//check cmd if still running
	logCmdsocket("cmd process ",cmd.ProcessState)
	if cmd.ProcessState == nil && cmd.Process != nil {
		logCmdsocket("cmd still running, close it")
		cmd.Process.Kill()
	}
	logCmdsocket("goroutine read loop closed ")
}
func startCmdScripts(ctx context.Context, config CmdConfig,safeSocket *SocketSafeWriter, inputChan chan string, okChan *chan bool)  {
	logCmdsocket("ready for send cmd ",config.Cmd)
	*okChan <- true

	var cmd *exec.Cmd
	if len(config.Param) > 0{
		cmd = exec.Command(config.Cmd,config.Param)
	}else {
		cmd = exec.Command(config.Cmd)
	}

	pf,perr := cmd.StdinPipe()
	if perr != nil {
		logCmdsocket("pipe err ",perr)
	}
	defer pf.Close()

	go writeInputToCmd(cmd,&pf,inputChan)

	r,w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	defer r.Close()
	defer w.Close()
	go readOutputFromCmd(cmd,r,safeSocket)

	select {
	case <-ctx.Done():
		logCmdsocket("cmd context forcus finished, skip cmd ",config.Cmd)
	default:
		err := cmd.Run()
		if err!=nil {
			logCmdsocket("cmd failed ",err)
			safeSocket.SendMsg([]byte("CMD Error: "+err.Error()))
		}

		safeSocket.SendMsg([]byte("CMD Finished!"))
	}

	logCmdsocket("goroutine cmd exit")
	//without goroutine, okchan will stuck this routine until next okchan read
	go func() {
		if *okChan != nil{
			*okChan <- false
		}
		logCmdsocket("goroutine okchan finish!")
	}()
}
func excmd(ctx context.Context,config CmdConfig, safeSocket *SocketSafeWriter, cmdChan chan string, okChan *chan bool) {
	go startCmdScripts(ctx,config,safeSocket,cmdChan,okChan)
}

func logCmdsocket(v ...interface{})  {
	Debug.Println(cmdsocket_log_prefix,v)
}