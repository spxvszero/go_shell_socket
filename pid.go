package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var pid_file_name = "cmdsocket.pid"

var pid_log_prefix = "PID : "

func setupPidModule()  {
	writePidFile()
	signalReceive()
}

func writePidFile()  {
	pid := os.Getpid()
	file, err := os.Create(pid_file_name)
	if err != nil {
		Error.Println("create pid file :", "error : ",err)
		return
	}
	defer file.Close()

	pidstr := fmt.Sprint(pid)
	_,err = io.WriteString(file,pidstr)
	if err != nil {
		Error.Println("write config file error : ",err)
		return
	}
	Info.Println("Finished! Pid ",pidstr," successfully generated.")
}

func readPidFromFile() int{
	buf,err := ioutil.ReadFile(pid_file_name)
	if err != nil {
		logPid("read pid file error",err)
		return -1
	}
	pid,err := strconv.Atoi(string(buf))
	if err !=nil {
		logPid("string convert err :",err)
		return -1
	}
	return pid
}

func signalReceive()  {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP)
	go func() {
		for s := range sigc  {
			// reload cmd from cmd file
			logPid("get signal ",s)
			serializeCmds(UserConfig.Cmd)
		}
	}()

}

func sendSignal()  {
	pid := readPidFromFile()
	if pid <= 0{
		logPid("pid num small ")
		return
	}
	proc,err := os.FindProcess(pid)
	if err != nil {
		logPid("process not found ",err)
		return
	}
	err = proc.Signal(syscall.SIGHUP)
	if err != nil {
		logPid("send signal err ",err)
	}else {
		logPid("send signal success")
	}
}

func logPid(v ...interface{})  {
	Debug.Println(pid_log_prefix,v)
}