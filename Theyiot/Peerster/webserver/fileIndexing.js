let fileIndexing = function() {
    let fileInput = document.getElementById("inputFileUpload");
    let filename = fileInput.files[0].name;
    $.ajax({
        type: "POST",
        url: "/fileIndexing",
        contentType: 'application/json; charset=utf-8',
        data: JSON.stringify({ "Text": filename}),
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
    });
};

let getIndexedFiles = function() {
    $.ajax({
        type: "GET",
        url: "/fileIndexing",
    }).done(function(answer) {
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
        let indexedFiles = JSON.parse(answer);

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
    })
};