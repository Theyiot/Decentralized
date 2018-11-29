// GETTING ADDRESSES FROM BACKEND
let getAddresses = function() {
    $.ajax({
        type: "GET",
        url: "/node",
    }).done(function(answer) {
        let peers = JSON.parse(answer);

        let tabBody=document.getElementById("tableAddresses");
        tabBody.innerHTML =
            `<colgroup>
                <col width="150">
                <col width="55">
            </colgroup>

            <tr>
                <th>IP address</th>
                <th>Port</th>
            </tr>`;

        for(let i = 0 ; i < peers.length ; i++) {
            let addrTable = document.createElement("td");
            let portTable = document.createElement("td");
            let addrPort = peers[i].split(":");
            addrTable.appendChild(document.createTextNode(addrPort[0]));
            portTable.appendChild(document.createTextNode(addrPort[1]));
            let row = document.createElement("tr");
            row.appendChild(addrTable);
            row.appendChild(portTable);
            tabBody.appendChild(row);
        }
    }).always(function() {
        setTimeout(loadFromBackend, timeoutDelay)
    });
};

// ADDING NEW PEERS FROM FORM
$("#formAddPeer").submit(function (e) {
    e.preventDefault();
    let ipAddress = $("#inputAddress"), port = $("#inputPort");
    let ipVal = ipAddress.val(), portVal = port.val();
    if(ipVal === "" || portVal === "") {
        //Redundant check, but used to provide clearer error message
        alert("The IP address and the port fields cannot be empty")
        return;
    } else if(!checkValidIP(ipVal)) {
        alert("The IP address should have the form X.X.X.X, where each X is a number between 0 and 255 included, but was " + ipVal);
        return;
    } else if(!checkValidPort(portVal)) {
        alert("The port should be between 1025 and 65535 included, but was " + portVal);
        return;
    }
    $.ajax({
        type: "POST",
        url: "/node",
        contentType: 'application/json; charset=utf-8',
        data: JSON.stringify({ "Text": ipVal + ':' + portVal }),
        dataType: 'json',
    });
    ipAddress.val("");
    port.val("");
    getAddresses()
});