let checkValidHash = function(hash) {
    return new RegExp('\\b[A-Fa-f0-9]{64}\\b').test(hash)
};

let checkValidIP = function(ip) {
    if(typeof ip !== "string") {
        alert("Conception error, checkValidIP should get a string, but it was " + typeof ip);
        return false;
    }
    let blocks = ip.split(".");
    if(blocks.length !== 4) {
        return false;
    }
    let intValue;
    for(let i = 0 ; i < blocks.length ; i++) {
        intValue = parseInt(blocks[i]);
        if(isNaN(intValue) || 0 > intValue > 255) {
            return false;
        }
    }
    return true;
};

let checkValidPort = function(port) {
    if(typeof port !== "string") {
        alert("Conception error, checkValidPort should get a string, but it was " + typeof port);
        return false;
    }
    let intValue;
    try {
        intValue = parseInt(port);
    } catch (e) {
        return false;
    }
    return 1024 < intValue < 65536;
}