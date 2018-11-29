// GETTING NAMES FROM BACKEND
let getNames = function() {
    $.ajax({
        type: "GET",
        url: "/name",
    }).done(function(answer) {
        let names = JSON.parse(answer);

        let table = document.getElementById("tableNames");
        table.innerHTML =
            `<col width="205">
            <tr>
                <th>Name</th>
            </tr>`;

        let select1 = document.getElementById("selectPrivate1");
        let value1 = select1.value;
        let select2 = document.getElementById("selectPrivate2");
        let value2 = select2.value;
        select1.innerHTML =
            `<select id="selectPrivate1">
                <option selected="selected" id="default1" disabled hidden>Choose a conversation</option>
            </select>`;
        select2.innerHTML =
            `<select id="selectPrivate2">
                <option selected="selected" id="default2" disabled hidden>Choose a peer to interact with</option>
            </select>`;

        for(let i = 0 ; i < names.length ; i++) {
            //NAMES TABLE
            let nameEntry = document.createElement("td");
            nameEntry.appendChild(document.createTextNode(names[i]));
            let row = document.createElement("tr");
            row.appendChild(nameEntry);
            table.appendChild(row);

            //NAMES SELECTS
            let opt1 = document.createElement("option");
            opt1.value = opt1.innerHTML = names[i];
            opt1.id = names[i] + "1";
            select1.appendChild(opt1);
            let opt2 = document.createElement("option");
            opt2.value = opt2.innerHTML = names[i];
            opt2.id = names[i] + "2";
            select2.appendChild(opt2);
        }
        select1.value = value1;
        select2.value = value2;
    });
};