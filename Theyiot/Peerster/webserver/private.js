// GETTING PRIVATE MESSAGES FROM BACKEND
let getPrivateMessages = function() {
    $.ajax({
        type: "GET",
        url: "/private",
    }).done(function(answer) {
        let textPrivate = $("#textReceivedPrivateMessages");
        let privates = JSON.parse(answer);
        let str = "";
        let privatePeer = document.getElementById("selectPrivate1").value;
        if(privates[privatePeer] === undefined) {
            return;
        }
        let msgList = privates[privatePeer];
        for(let i = 0 ; i < msgList.length ; i++) {
            let msg = msgList[i].Private;
            str +=  (msg.Origin === privatePeer ? msg.Origin : "Me") + " :\n" + msg.Text + "\n";
        }
        textPrivate.val(str);
    });
};

// SENDING PRIVATE MESSAGE FROM WEB SERVER
let sendPrivateMessage = function(textMsg) {
    if(document.getElementById("default2").selected) {
        alert("You have to select a peer to send a private message");
        return false;
    }
    $.ajax({
        type: "POST",
        url: "/private",
        contentType: 'application/json; charset=utf-8',
        data: JSON.stringify({ "Text": textMsg.val(), "Peer": document.getElementById("selectPrivate2").value }),
        dataType: 'json',
    });
    getPrivateMessages();
    return true;
};