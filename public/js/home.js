const initWebsocket = (wsObj) => {
  const ws = new WebSocket(wsObj);
  ws.onmessage = (evt) => {
    printTweet(evt.data);
  };
  ws.onerror = (evt) => {
    console.log("ERROR: " + evt.data);
  };
  const printTweet = (message) => {
    const data = JSON.parse(message);

    const container = document.createElement("tweet-container");
    
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

    const output = document.getElementById("output");
    output.appendChild(container);
  }
};
