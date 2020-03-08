window.addEventListener("load", function(evt) {
  var output = document.getElementById("output");
  var input = document.getElementById("input");
  var ws;
  ws = new WebSocket("{{ .WsHost }}");
  ws.onopen = function(evt) {
    console.log("OPEN");
  };
  ws.onclose = function(evt) {
    console.log("CLOSE");
    ws = null;
  };
  ws.onmessage = function(evt) {
    printTweet(evt.data);
  };
  ws.onerror = function(evt) {
    console.log("ERROR: " + evt.data);
  };
  var printTweet = function(message) {
    var container = document.createElement("div");
    var ids = document.createElement("div");
    const data = JSON.parse(message);
    const tweetText = document.createElement("p");
    tweetText.classList.add("red-class");
    tweetText.innerHTML = data.Tweet.Text;
    container.appendChild(tweetText);
    const cid = document.createElement("span");
    const dig = document.createElement("span");
    cid.innerHTML = data.Cid;
    dig.innerHTML = data.Digest;
    ids.appendChild(cid);
    ids.appendChild(dig);
    container.appendChild(ids);

    output.appendChild(container);
    
  };
  document.getElementById("close").onclick = function(evt) {
    if (!ws) {
      return false;
    }
    ws.close();
    return false;
  };
});