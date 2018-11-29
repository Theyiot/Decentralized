// GETTING ID AND DISPLAYING IT
$.ajax({
    type: "GET",
    url: "/id",
}).done(function (response) {
    let responseObject = JSON.parse(response);
    $("#textID").val("\nAddress : " + responseObject.Address + "\tName : " + responseObject.Name);
});