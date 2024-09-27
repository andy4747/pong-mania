let socket;
let gameStarted = false;

function connectWebSocket() {
  socket = new WebSocket("ws://10.0.0.1:8080/ws");

  socket.onopen = function (e) {
    console.log("WebSocket connection established");
  };

  socket.onmessage = function (event) {
    if (event.data === "START_GAME") {
      gameStarted = true;
      window.location.href = "/play";
    } else {
      document.getElementById("status").textContent = event.data;
    }
  };

  socket.onclose = function (event) {
    if (!gameStarted) {
      console.log("WebSocket connection closed. Retrying in 5 seconds...");
      setTimeout(connectWebSocket, 5000);
    }
  };

  socket.onerror = function (error) {
    console.log("WebSocket error: " + error.message);
  };
}

function startGame() {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send("START");
    document.getElementById("status").textContent = "Waiting for opponent...";
  } else {
    console.log("WebSocket not connected");
  }
}

const startBtn = document.getElementById("startBtn");
startBtn.addEventListener("click", (e) => {
  connectWebSocket();
});
