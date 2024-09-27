const gameContainer = document.getElementById("game-container");
const canvasContainer = document.getElementById("canvas-container");
const canvas = document.getElementById("gameCanvas");
const ctx = canvas.getContext("2d");
gameContainer.style = `cursor: url(https://cur.cursors-4u.net/cursors/cur-1/cur1.ani),
    url(https://cur.cursors-4u.net/cursors/cur-1/cur1.png), auto !important;`;

const GAME_WIDTH = 800;
const GAME_HEIGHT = 400;
let scale = 1;

let lastY = 0;
/* *
 * @type {string}
 */
const roomId = document.getElementById("game-container").attributes["data-room-id"].value;


const ws = new WebSocket(`/play/conn/${roomId}`);



function resizeCanvas() {
    const containerWidth = gameContainer.clientWidth;
    const containerHeight = gameContainer.clientHeight;
    const containerRatio = containerWidth / containerHeight;
    const gameRatio = GAME_WIDTH / GAME_HEIGHT;

    let canvasWidth, canvasHeight;

    if (containerRatio > gameRatio) {
        canvasHeight = containerHeight;
        canvasWidth = canvasHeight * gameRatio;
    } else {
        canvasWidth = containerWidth;
        canvasHeight = canvasWidth / gameRatio;
    }

    canvasContainer.style.width = `${canvasWidth}px`;
    canvasContainer.style.height = `${canvasHeight}px`;

    canvas.width = GAME_WIDTH;
    canvas.height = GAME_HEIGHT;
    canvas.style.width = `${canvasWidth}px`;
    canvas.style.height = `${canvasHeight}px`;

    scale = canvasWidth / GAME_WIDTH;
}

window.addEventListener("resize", resizeCanvas);
resizeCanvas();


ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === "gameState") {
        console.log(data.gameState)
        drawGame(data.gameState);
    } else if (data.type === "gameEnd") {
        displayFinalScore(data.score1, data.score2);
    }
};

function displayFinalScore(score1, score2) {
    ctx.clearRect(0, 0, GAME_WIDTH, GAME_HEIGHT);
    ctx.font = "36px Arial";
    ctx.fillStyle = "white";
    ctx.fillText(`Final Score: ${score1} - ${score2}`, 250, 200);

    setTimeout(() => {
        window.location.href = "/score";
    }, 3000);
}

function drawGame(state) {
    ctx.clearRect(0, 0, GAME_WIDTH, GAME_HEIGHT);

    // Draw paddles
    ctx.fillStyle = "white";
    ctx.fillRect(0, state.player1Y, 10, 100);
    ctx.fillRect(790, state.player2Y, 10, 100);

    // Draw ball
    ctx.beginPath();
    ctx.arc(state.ballX, state.ballY, 5, 0, Math.PI * 2);
    ctx.fill();

    // Draw scores
    ctx.font = "24px Arial";
    ctx.fillText(state.score1, 350, 30);
    ctx.fillText(state.score2, 450, 30);

    // Draw timer
    ctx.font = "18px Arial";
    ctx.fillText(`Time Left: ${formatTime(state.remainingTime)}`, 10, 30);
}

function formatTime(seconds) {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
}

canvas.addEventListener("mousemove", (event) => {
    const rect = canvas.getBoundingClientRect();
    const y = (event.clientY - rect.top) / scale;
    if (Math.abs(y - lastY) > 2) {
        lastY = y;
        ws.send(JSON.stringify({ type: "move", y: y }));
    }
});

canvas.addEventListener("mouseenter", () => {
    document.exitPointerLock();
});
