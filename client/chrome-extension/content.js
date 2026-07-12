console.log("content.js loaded")

function getPlayer() {
    return document.querySelector("ytmusic-player-bar");
}

function getNextButton() {
    return getPlayer()?.querySelector("yt-icon-button.next-button button")
}

function getPlayPauseButton() {
    return getPlayer()?.querySelector("yt-icon-button.play-pause-button button")
}

function getPreviousButton() {
    return getPlayer()?.querySelector("yt-icon-button.previous-button button")
}





function next(){
    getNextButton()?.click();
}

function playPause() {
    getPlayPauseButton()?.click();
}

function previous() {
    getPreviousButton()?.click();
}


chrome.runtime.onMessage.addListener((msg) => {

    console.log(msg)

    switch (msg.action) {
        case "playPause":
            playPause();
            break;
        case "next":
            next();
            break;
        case "previous":
            previous();
            break;
    }

});