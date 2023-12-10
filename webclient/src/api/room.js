var roomSocket;

let connectRoom = (cb, roomName, username) => {
    console.log(`connecting to room ${roomName}`);
    roomSocket = new WebSocket(`ws://12.12.12.10:8080/rooms/${roomName}?username=${username}`);

    roomSocket.onopen = () => {
        console.log("Successfully Connected");
    };

    roomSocket.onmessage = msg => {
        console.log(msg);
        cb(msg);
    };

    roomSocket.onclose = event => {
        console.log("Socket Closed Connection: ", event);
    };

    roomSocket.onerror = error => {
        console.log("Socket Error: ", error);
    };
};

let getRoomMessages = () => { return []; }

let sendMsg = msg => {
    console.log("sending msg: ", msg);
    roomSocket.send(msg);
};

export { connectRoom, getRoomMessages, sendMsg };