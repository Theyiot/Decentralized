let requestFile = function(fileName, hashRequest) {
    if(fileName.val() === "") {
        fileName.select();
        alert("You need to enter a name for the file you want to download");
        return false;
    }
    if(!checkValidHash(hashRequest.val())) {
        hashRequest.select();
        alert("You need to enter a valid SHA256 hash (in hex format), enter 64 hexadecimal characters string");
        return false;
    }
    $.ajax({
        type: "POST",
        url: "/fileRequesting",
        contentType: 'application/json; charset=utf-8',
        data: JSON.stringify({ "FileName": fileName.val(), "Request": hashRequest.val(),
            "Dest": document.getElementById("selectPrivate2").value }),
        dataType: 'json',
    }).done(function(indexedFiles) {
        let table = document.getElementById("tableFiles");
        table.innerHTML = `
                    <colgroup>
                        <col width="370px">
                        <col width="150px">
                    </colgroup>
                    <tr>
                        <th>Metahash</th>
                        <th>File name</th>
                    </tr>`;

        Object.keys(indexedFiles).forEach(function (metaHash) {
            let metaHashTable = document.createElement("td");
            let fileNameTable = document.createElement("td");
            metaHashTable.appendChild(document.createTextNode(metaHash));
            fileNameTable.appendChild(document.createTextNode(indexedFiles[metaHash]));
            let row = document.createElement("tr");
            row.appendChild(metaHashTable);
            row.appendChild(fileNameTable);
            table.appendChild(row);
        });
        alert("Your file was correctly downloaded !")
    });
};