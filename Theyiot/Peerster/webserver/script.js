let timeoutDelay = 5000;

// LOAD EVERY DATA FROM THE BACKEND (MESSAGES, PEERS, ...)
let loadFromBackend = function() {
    getPublicMessages();
    getPrivateMessages();
    getAddresses();
    getNames();
    getIndexedFiles();
};

// SENDING MESSAGE (PUBLIC, PRIVATE AND FILE REQUEST) FROM WEB SERVER
$("#formSending").submit(function (e) {
    e.preventDefault();
    let textMsg = $("#textContent");
    let hashRequest = $("#inputHashRequest");
    let fileName = $("#inputFileName");
    if(textMsg.val() === "" && !document.getElementById("radioFileRequest").checked) {
        textMsg.select();
        alert("You cannot send an empty message, write something before sending");
        return;
    }
    let success = true;
    if(document.getElementById("radioPublic").checked) {
        sendPublicMessage(textMsg)
    } else if(document.getElementById("radioPrivate").checked) {
        success = sendPrivateMessage(textMsg)
    } else if(document.getElementById("radioFileRequest").checked) {
        success = requestFile(fileName, hashRequest);
    }
    if(success) {
        textMsg.val("");
        hashRequest.val("");
        fileName.val("");
        textMsg.select();
    }
});

// SELECT THE RIGHT PEER TO HAVE PRIVATE CONVERSATIONS
let choosePeer = function(index) {
    let selectedFrom = "selectPrivate" + index;
    let privatePeerIndex = document.getElementById(selectedFrom).selectedIndex;
    selectPrivatePeer(privatePeerIndex);
    getPrivateMessages();
};

let selectPrivatePeer = function(privatePeerIndex) {
    document.getElementById("selectPrivate1").selectedIndex = privatePeerIndex;
    document.getElementById("selectPrivate2").selectedIndex = privatePeerIndex;
};

loadFromBackend();