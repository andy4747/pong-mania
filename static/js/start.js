function startWsConn() {
  const ws = new WebSocket("/game/random/start");
  ws.onopen = (event) => {
    const data = JSON.parse(event.data);
    console.log("open data", data);
    //console.log(event);
  };

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log("data =>", data);
    if (data.type === "open") {
      window.location.replace(data.redirect_url);
    }
  };
}
