package main

const CmdSocketHTMLPage = `

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Cmd Socket</title>
    <script src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
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
</div>


<script>
    url = {{.Socket_Path}}
    c = new WebSocket(url);

    function sendmsg(){
        console.log("click")
        div = document.getElementById("cmd");
        c.send(div.value)
    }


    send = function(data){
        $("#output").append((new Date())+ " ==> "+data+"\n")
        c.send(data)
    }

    c.onmessage = function(msg){
        $("#output").append((new Date())+ " <== "+msg.data+"\n")
        window.scrollTo(0,document.body.scrollHeight);
        console.log(msg)
    }

    c.onopen = function(err){
        console.log("open : ",err)
    }
    c.onerror = function (err) {
        console.log(err)
    }

    $("#cmd").on('keyup', function (e) {
        if (e.key === 'Enter' || e.keyCode === 13) {
            sendmsg()
        }
    });
</script>


</body>
</html>

`