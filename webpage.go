package main

const CmdSocketHTMLPage = `

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Cmd Socket</title>

    <style>
        .stick-head{
            display: flex;
            position: sticky;
            top: 0px;
            background-color: white;
        }
        .stick-foot{
            display: flex;
            position: sticky;
            bottom: 0px;
            background-color: white;
        }
    </style>
</head>
<body>

<h3 class="stick-head">Cmd Socket Go</h3>
<pre id="output"></pre>

<div class="stick-foot">
    <input id="cmd">
    <button id="sendcmd" onclick="sendmsg()">send</button>
    <button id="clearcmd" onclick="clearmsg()">clear</button>
</div>

<script>
    url = "ws://"+window.location.host+{{.Socket_Path}}
 	c = new WebSocket(url);

    function sendmsg(){
        console.log("click");
        div = document.getElementById("cmd");
        c.send(div.value);
    }
    function clearmsg(){
        let div = document.getElementById("output");
        if (div != null) div.innerHTML="";
    }

    c.onmessage = function(msg){
        document.getElementById("output").append((new Date())+ " <== "+msg.data+"\n");
        window.scrollTo(0,document.body.scrollHeight);
        console.log(msg);
    }

    c.onopen = function(evt){
        document.getElementById("output").append("WebSocket Open : \n");
        console.log("open : ",evt);
    }
    c.onclose = function(evt){
        document.getElementById("output").append("WebSocket Closed !! \n");
        console.log("close : ",evt);
    }
    c.onerror = function (evt) {
        document.getElementById("output").append("WebSocket Get Error : "+evt+"\n");
        console.log(evt);
    }

    document.getElementById("cmd").addEventListener("keyup", function (e) {
        if (e.key === 'Enter' || e.keyCode === 13) {
            sendmsg();
        }
    })


</script>


</body>
</html>

`