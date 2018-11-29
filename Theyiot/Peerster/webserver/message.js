// GETTING PUBLIC MESSAGES FROM BACKEND
let getPublicMessages = function() {
    $.ajax({
        type: "GET",
        url: "/message",
    }).done(function(answer) {
        let textRumors = $("#textReceivedPublicMessages");
        let rumors = JSON.parse(answer);
        let str = "";
        for(let i = 0 ; i < rumors.length ; i++) {
            let rumor = rumors[i].Rumor;
            str +=  rumor.Origin + " says :\n" + rumor.Text + "\n";
        }
        textRumors.val(str);
    });
};

// SENDING PUBLIC MESSAGE FROM WEB SERVER
let sendPublicMessage = function(textMsg) {
    $.ajax({
        type: "POST",
        url: "/message",
        contentType: 'application/json; charset=utf-8',
        data: JSON.stringify({ "Text": textMsg.val() }),
        dataType: 'json',
    });
    getPublicMessages();
};