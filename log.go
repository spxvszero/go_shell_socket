package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Debug 	*log.Logger
)

type LogMode int
const (
	DebugMode LogMode = iota
	ReleaseMode
)

var Mode = DebugMode

func init()  {
	defaultLogger(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr,os.Stdout)
}

func defaultLogger(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
	debugHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Debug = log.New(debugHandle,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func SetLogMode(mode LogMode)  {
	Mode = mode
}

func SetWriter(w io.Writer) {
	Trace.SetOutput(w)
	Info.SetOutput(w)
	Warning.SetOutput(w)
	Error.SetOutput(w)
	if Mode == DebugMode {
		Debug.SetOutput(w)
	}else {
		Debug.SetOutput(ioutil.Discard)
	}
}