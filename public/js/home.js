
var ws;
function initWebsocket(wsObj) {
  ws = new WebSocket(wsObj);
  var output = document.getElementById("output");
  console.log(output);
  console.log(ws);
  ws.onmessage = function (evt) {
    printTweet(evt.data);
  };
  ws.onerror = function (evt) {
    console.log("ERROR: " + evt.data);
  };
  var printTweet = function (message) {
    const data = JSON.parse(message);

    var container = document.createElement("tweet-container");
    
    // tweet body
    const tweetText = document.createElement("p");
    tweetText.classList.add("tweet-body");
    tweetText.innerHTML = data.Tweet.Text;
    container.appendChild(tweetText);

    // ids
    const ids = document.createElement("tweet-ids");
    const cid = document.createElement("span");
    const dig = document.createElement("span");
    cid.innerHTML = data.Cid;
    dig.innerHTML = data.Digest;
    ids.appendChild(cid);
    ids.appendChild(dig);
    container.appendChild(ids);

    output.appendChild(container);
  }
};
