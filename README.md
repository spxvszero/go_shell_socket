# go_shell_socket

run shell scripts by websocket with go lauguage

一个在 web 端执行服务器指令的小工具。

## Usage
#### Build yourself

* pull project
* get [gin](https://github.com/gin-gonic/gin) and [websocket](https://github.com/gorilla/websocket) with `go get`
* config `config.json` file
* build it with `go build`
* open browser with config url and have fun



* 下载代码或者直接用 git 拉取过来
* 通过 go get 装好  [gin](https://github.com/gin-gonic/gin) 和  [websocket](https://github.com/gorilla/websocket)
* 配置 config.json 配置文件
* 然后就可以通过 go build 编译了，跨平台编译根据官网的说明来就好了，没有特殊配置
* 然后 go run 执行，就可以打开网页测试了


#### Use Release File

Sample For Linux :
```
$ curl -O -L https://github.com/spxvszero/go_shell_socket/releases/download/v1.0/go_shell_socket_linux
$ chmod +x go_shell_socket_linux

#generate config.json file
$ ./go_shell_socket_linux --generate config.json

#edit config.json
#run in background if config.json when in same path
$ ./go_shell_socket_linux &

```


## Example

![demo](readme_source/demo.gif)

## Command For Application

* `--help` to see what commands are supported.
* `--generate [file_path]` to generate a sample config.json file.
* `--signal reload` to reload command file which config in parameter `cmds_file` while server is running.

其实总共就两个指令

* help 指令是 golang 的 flag 包自带的。
* generate 指令会生成一个 config.json 文件，方便配置。
* signal 指令只接受 reload 参数，可以重新加载 cmds_file 配置好的文本文件。 

## Config

This is config.json which example used. Don't copy to your own config.json directly.

这个配置文件仅用做说明，不要直接拷贝使用。

```
{
//web page open port 
//网页打开的端口
  "port":18080,
  
  "Cmd_config": {
  
  //web socket page url 
  //自带的网页地址
    "web_url": "/cmdsocket",
    
  //socket url 
  //websocket的地址
    "socket_url": "/ws",
    
  //addition command files on server which can be reloaded without restart go program 
  //本地定义的一个文本文件，与 cmds 区别就是，这个文件可以通过此程序的 --signal reload 重载配置，而不需要重启程序
    "cmds_file": "cmdsFile",
    
  //inner cmds which loaded on first boot
  //这个字段的内容在启动的时候加载，不支持 reload 重载
    "cmds": [
      {
      
      //alias for actual command
      //指令的别名，web中仅能通过这个访问指令
        "exec_alias": "ll",
        
      //command
      //实际运行的指令
        "cmd": "ls",
        
      //param not supported long params or pipe, use script file instead
      //指令的参数，不支持管道和过长的指令，如果有需要，建议放在脚本中执行，这里只执行脚本
        "param": "-l"
      },
      {
        "exec_alias": "shtest",
        "cmd": "sh",
        "param": "test_go_exec.sh"
      },
      {
        "exec_alias": "pinggithub",
        "cmd": "ping",
        "param": "github.com"
      },
      {
        "exec_alias": "cshow",
        "cmd": "cat",
        "param": "config.json"
      }
    ],
    
    //exit current running cmd
    //用来中止当前运行的指令
    "exit_string": "exit"
  }
}
```

## Format of cmds_file

`[exec_alias]:[cmd] [param]`

It easy to add new command.

Look how to add a `ps` command.

```
#server is running in background, and there is a cmdsFile which i add path to config.json.
#add cmd
$ echo "ppps:ps aux" >> cmdsFile

#send signal to server
$ ./go_shell_socket_linux --signal reload
```

## Something more

* This program does not support any text editor such as `vim`.
* It works fin in my macOS Catalina and CentOS 8, and i am not sure about other system.
* `param` in config file will go through to `exec.Command`, so it does not support complicated command. If you want ,put command in shell script and run with `sh`.
* ssh command will stuck in goroutine which cannot be kill，but why use it.
* nginx do not support websocket default, if using nginx, see how to config it in this [site](https://nginx.org/en/docs/http/websocket.html).
* This program will generate a log file and a pid file while running. 



* 这个小工具不支持文本编辑了，不要把它当作一个终端来用，写这个的目的只是为了方便某些不适合登录终端的场景了，要做复杂的任务还是老老实实登入终端吧。
* 这个程序我在 mac 上和 centos8 上测试是正常的，windows 我也试了，交互正常，不过命令我用的少就没怎么测试了。
* 不支持多参数主要自己不太需要，就没有写了，如果有需要可以自己修改 `exec` 执行的形式。
* 还有 ssh 命令是会有问题的，开启的进程没办法正常 kill 掉，不晓得什么原因，不过通过一些方式修改之后，应该有办法实现的吧。
* 通过 nginx 反向代理需要一些特殊配置才能支持 websocket 了，可以看[官网](https://nginx.org/en/docs/http/websocket.html)的描述。
* 这个工具在运行的时候会生成一个 log 文件和一个 pid 文件。