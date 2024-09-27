let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
let waitCounter = 0;

function connectWebSocket(roomId) {
  let socket = new WebSocket(`/wait/conn?room_id=${roomId}`);
  if (socket && socket.readyState === WebSocket.OPEN) {
    console.log("WebSocket is already connected");
    return;
  }

  socket.onopen = function (event) {
    console.log("WebSocket connected");
    reconnectAttempts = 0;
  };

  socket.onmessage = function (event) {
    const message = JSON.parse(event.data);
    console.log(message);
    if (message.type === "redirect") {
      window.location.href = message.redirect_url;
    } else {
      console.log("Received message:", message);
      waitCounter += 1;
      document.getElementById("countdown").innerText = waitCounter;
    }
  };

  socket.onclose = function (event) {
    console.log("WebSocket disconnected");
    if (reconnectAttempts < maxReconnectAttempts) {
      setTimeout(() => {
        reconnectAttempts++;
        console.log(
          `Attempting to reconnect (${reconnectAttempts}/${maxReconnectAttempts})`
        );
        connectWebSocket(roomId);
      }, 3000); // Wait 3 seconds before attempting to reconnect
    } else {
      console.log("Max reconnection attempts reached");
    }
  };

  socket.onerror = function (error) {
    console.error("WebSocket error:", error);
  };
}

const roomId = document.getElementById("roomId").innerText;
connectWebSocket(roomId);
