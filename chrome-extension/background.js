console.log("Location:", self.location.href);
console.log("Origin:", self.origin);

console.log("background.js loaded")

const socket = new WebSocket("ws://localhost:8080/ws");



socket.onopen = () => {
    console.log("Conectado al servidor");

    //register player
    const data = {type: "register", data: "player"}
    socket.send(JSON.stringify(data))
};

socket.onmessage = async (event) => {

    console.log("Mensaje recibido")

    const tabs = await chrome.tabs.query({ url: "https://music.youtube.com/*" });

    const message = JSON.parse(event.data)

    for (const tab of tabs) {
        chrome.tabs.sendMessage(tab.id, message.data);
    }
};

socket.onerror = (e) => {
    console.error("WebSocket error:", e);
};

socket.onclose = (e) => {
    console.log("WebSocket cerrado", e.code, e.reason);
};

