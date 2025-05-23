const chatInput = document.getElementById('chat-input');
const playerHandDiv = document.getElementById('player-hand');
const lastPlayedDiv = document.getElementById('last-played-cards');
const chatLogDiv = document.getElementById('chat-log');
const playersAreaDiv = document.getElementById('players-area');
const gameMessagesDiv = document.getElementById('game-messages');
const playerActionMessagesDiv = document.getElementById('player-action-messages');
const passTurnButton = document.getElementById('pass-turn-button');
const aliasModal = document.getElementById('alias-modal');
const aliasInput = document.getElementById('alias-input');
const submitAliasButton = document.getElementById('submit-alias-button');
const aliasErrorP = document.getElementById('alias-error');
const audioDingTurn = document.getElementById('audioDingTurn');
const audioDingPlayed = document.getElementById('audioDingPlayed');

// WebSocket and Game State Globals
let socket = null;
let gCurrentPlayerHand = [];
let gYourPlayerId = null;
let gPlayersInfo = []; // Array of {id, name, cardCount, hasPassed}
let gLastPlayedHand = null;
let gCurrentPlayerId = null; // ID of the player whose turn it is
let gPassCount = 0;
let gIsGameOver = false; // Round over
let gWinnerId = null;    // Round winner
let gScores = {};        // PlayerID -> Total Score for the match
let gRoundScoresHistory = []; // Array of round score objects (each is PlayerID -> ScoreForThatRound)
let gRoundNumber = 1;
let gTargetScore = 100;
let gIsMatchOver = false; // Match over
let gOverallWinnerId = null; // Match winner

// UI State Globals
let gAutoPassEnabled = false;
let gIsMyTurn = false;
let gPreviousCurrentPlayerId = null; // For turn sound logic
let gAudioUnlocked = false;
let gCurrentSortPreference = 'rank'; // Added for persisting sort order

// Card Rank and Suit mapping (from card.go enums)
//Go Ranks: Three=3, Four=4, ..., Queen=12, King=13, Ace=14, Two=15
//SVG Ranks: 1 (Ace), 2-10, jack, queen, king
const rankMap = {
    3: '3', 4: '4', 5: '5', 6: '6', 7: '7', 8: '8', 9: '9', 10: '10',
    11: 'jack', 12: 'queen', 13: 'king',
    14: '1',  // Ace (game value 14) is '1' in SVG
    15: '2'   // Two (game value 15) is '2' in SVG
};

//Go Suits: Diamonds=0, Clubs=1, Hearts=2, Spades=3
//SVG Suits: diamond, club, heart, spade
const suitMap = {
    0: 'diamond',
    1: 'club',
    2: 'heart',
    3: 'spade'
};

function connectWebSocket() {
    socket = new WebSocket('ws://' + window.location.host + '/ws');

    socket.onopen = function(event) {
        console.log("WebSocket connection established.");

        const storedAlias = localStorage.getItem('playerAlias');
        if (storedAlias) {
            console.log("Found stored alias:", storedAlias);
            sendAliasToServer(storedAlias);
            aliasModal.style.display = 'none';
            document.getElementById('main-container').style.display = 'block';
        } else {
            console.log("No stored alias found. Displaying modal.");
            aliasModal.style.display = 'flex';
            document.getElementById('main-container').style.display = 'none';
        }
    };

    socket.onmessage = function(event) {
        console.log("Message from server: ", event.data);
        const message = JSON.parse(event.data);

        // Clear previous game messages on any new message from server that isn't an error
        if (message.type !== 'error') {
            if (playerActionMessagesDiv) playerActionMessagesDiv.innerHTML = '';
        }

        switch (message.type) {
            case "gameState":
                console.log("Received gameState:", message);
                gPlayersInfo = message.playersInfo || []; 
                
                gYourPlayerId = message.yourPlayerId;
                gCurrentPlayerId = message.currentPlayerId; 
                gIsMyTurn = (message.currentPlayerId === gYourPlayerId && gYourPlayerId !== null);
                gRoundNumber = message.roundNumber || 1;
                gTargetScore = message.targetScore || 100; 
                gIsMatchOver = message.isMatchOver || false;
                gOverallWinnerId = message.overallWinnerId || null;
                gScores = message.scores || {};
                gRoundScoresHistory = message.roundScoresHistory || []; 
                gLastPlayedHand = message.lastPlayedHand;
                gPassCount = message.passCount;
                gIsGameOver = message.isGameOver;
                gWinnerId = message.winnerId;

                if (gAutoPassEnabled) {
                    const trickWinnerIsStartingNewTrick = gPassCount === 0 &&
                        gLastPlayedHand && gLastPlayedHand.cards && gLastPlayedHand.cards.length > 0 &&
                        gLastPlayedHand.playerId === gCurrentPlayerId;
                    
                    const veryFirstTurnOfRound = gPassCount === 0 && 
                                                 (!gLastPlayedHand || !gLastPlayedHand.cards || gLastPlayedHand.cards.length === 0);

                    if (trickWinnerIsStartingNewTrick || veryFirstTurnOfRound) {
                        console.log("Auto-pass is being reset because a new trick is starting or it's the beginning of a round.");
                        gAutoPassEnabled = false;
                    }
                }

                gCurrentPlayerHand = message.hand ? [...message.hand] : [];
                
                // Apply persistent sort order BEFORE updating the hand display
                if (gCurrentSortPreference === 'suit') {
                    sortHandBySuitInternal();
                } else {
                    sortHandByRankInternal(); // Default to rank
                }

                let baseTitle = "Big Two Game";
                if (gYourPlayerId) {
                    const playerInfo = gPlayersInfo.find(p => p.id === gYourPlayerId);
                    const playerName = playerInfo ? playerInfo.name : gYourPlayerId;
                    baseTitle = playerName ? `${playerName} - Big Two (R${gRoundNumber})` : `${gYourPlayerId} - Big Two (R${gRoundNumber})`;
                }
                document.title = (gIsMyTurn && !message.isGameOver && !gIsMatchOver) ? `YOUR TURN! - ${baseTitle}` : baseTitle;

                if (message.isMatchOver) {
                    displayGameOver(message.winnerId, message.scores, message.playersInfo, true, message.overallWinnerId, gRoundScoresHistory);
                    document.getElementById('play-selected-button').disabled = true;
                    if (passTurnButton) passTurnButton.disabled = true;
                    gAutoPassEnabled = false; 
                } else if (message.isGameOver) { 
                    displayGameOver(message.winnerId, message.scores, message.playersInfo, false, null, gRoundScoresHistory);
                    document.getElementById('play-selected-button').disabled = true;
                    if (passTurnButton) passTurnButton.disabled = true;
                    gAutoPassEnabled = false; 
                } else {
                    const gameOverScreen = document.getElementById('game-over-screen');
                    if (gameOverScreen) gameOverScreen.style.display = 'none';
                    
                    document.getElementById('play-selected-button').disabled = false;
                    if (passTurnButton) passTurnButton.disabled = false;

                    if (gIsMyTurn && gAutoPassEnabled) {
                        sendPassAction(); 
                    } 

                    updatePlayerHand(gCurrentPlayerHand, gYourPlayerId); 
                    updateLastPlayed(gLastPlayedHand, gPlayersInfo); 
                    updateTurnInfo(message.currentPlayerName, gPassCount, gCurrentPlayerId);
                    updatePlayersArea(gPlayersInfo, gYourPlayerId, gCurrentPlayerId); 

                    // Update player area background based on turn
                    const playerAreaDiv = document.getElementById('player-area');
                    if (playerAreaDiv) {
                        if (gIsMyTurn) {
                            playerAreaDiv.classList.add('active-player-turn');
                        } else {
                            playerAreaDiv.classList.remove('active-player-turn');
                        }
                    }
                }

                updateScoreTable(gPlayersInfo, gRoundScoresHistory, gRoundNumber, gScores, gTargetScore); 

                if (message.gameMessage) {
                    updateGameMessages(message.gameMessage, false, 'general');
                }
                updatePassButtonAppearance(); 

                const newCurrentPlayerId = message.currentPlayerId;

                if (gAudioUnlocked && newCurrentPlayerId !== gPreviousCurrentPlayerId && gPreviousCurrentPlayerId !== null) {
                    // Turn has changed
                    if (newCurrentPlayerId === gYourPlayerId) { // It's now YOUR turn
                        if (audioDingTurn) {
                            audioDingTurn.currentTime = 0; 
                            audioDingTurn.play().catch(error => {
                                console.warn("Audio play failed for your turn:", error);
                            });
                        }
                    } else { // It's now someone else's turn
                        if (audioDingPlayed) {
                            audioDingPlayed.currentTime = 0;
                            audioDingPlayed.play().catch(error => {
                                console.warn("Audio play failed for opponent's turn:", error);
                            });
                        }
                    }
                }
                gPreviousCurrentPlayerId = newCurrentPlayerId; 

                gIsMyTurn = (message.currentPlayerId === gYourPlayerId && gYourPlayerId !== null);
                break;
            case "chat":
                const chatEntry = document.createElement('p');
                chatEntry.textContent = `${message.sender}: ${message.content}`;
                chatLogDiv.appendChild(chatEntry);
                chatLogDiv.scrollTop = chatLogDiv.scrollHeight; // Scroll to bottom
                break;
            case "error":
                updateGameMessages(`Error: ${message.content}`, true, 'action');
                break;
            default:
                console.log("Unknown message type: ", message.type);
                updateGameMessages(`Received unknown message type: ${message.type}`, true, 'general');
        }
    };

    socket.onclose = function(event) {
        console.log("WebSocket is closed now.");
        document.getElementById('turn-info').textContent = "Disconnected. Refresh to rejoin.";
    };

    socket.onerror = function(error) {
        console.error("WebSocket error: ", error);
        updateGameMessages("WebSocket error. Check console and refresh to try again.", true, 'general');
    };
}

function createCardDiv(card) {
    const cardDiv = document.createElement('div');
    cardDiv.classList.add('card');
    cardDiv.textContent = `${rankMap[card.Rank]}${suitMap[card.Suit]}`;
    cardDiv.dataset.rank = card.Rank;
    cardDiv.dataset.suit = card.Suit;
    if (card.Suit === 0 || card.Suit === 2) { // Diamonds or Hearts
        cardDiv.classList.add('red-card');
    } else {
        cardDiv.classList.add('black-card');
    }
    return cardDiv;
}

function updatePlayerHand(cardsToDisplay, yourId) {
    // Store current selections before clearing
    const previouslySelectedCards = [];
    playerHandDiv.querySelectorAll('.card.selected').forEach(selectedCardElement => {
        previouslySelectedCards.push({
            rank: selectedCardElement.dataset.rank,
            suit: selectedCardElement.dataset.suit
        });
    });

    playerHandDiv.innerHTML = ''; // Clear previous cards
    if (!cardsToDisplay) {
        cardsToDisplay = []; 
    }

    cardsToDisplay.forEach(card => {
        const svgRank = rankMap[card.rank];
        const svgSuit = suitMap[card.suit];
        const cardId = `${svgSuit}_${svgRank}`;

        const cardSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        cardSvg.classList.add('card');
        cardSvg.setAttribute('data-rank', String(card.rank)); // Ensure dataset attributes are strings
        cardSvg.setAttribute('data-suit', String(card.suit)); // Ensure dataset attributes are strings
        cardSvg.setAttribute('xmlns', "http://www.w3.org/2000/svg");
        cardSvg.setAttribute('xmlns:xlink', 'http://www.w3.org/1999/xlink');
        cardSvg.setAttribute('transform', 'scale(0.7)');

        const useElement = document.createElementNS('http://www.w3.org/2000/svg', 'use');
        useElement.setAttribute('href', `cards.svg#${cardId}`);
        
        cardSvg.appendChild(useElement);

        // Reapply selection if this card was previously selected
        if (previouslySelectedCards.some(selCard => selCard.rank === String(card.rank) && selCard.suit === String(card.suit))) {
            cardSvg.classList.add('selected');
        }

        cardSvg.onclick = function() {
            this.classList.toggle('selected');
        };
        playerHandDiv.appendChild(cardSvg);
    });
}

function updateLastPlayed(lastPlayedHand, playersInfo) {
    const lastPlayedHandInfoDiv = document.getElementById('last-played-hand-info');
    lastPlayedDiv.innerHTML = ''; // Clear previous cards
    lastPlayedHandInfoDiv.innerHTML = ''; // Clear previous hand info

    if (lastPlayedHand && lastPlayedHand.cards && lastPlayedHand.cards.length > 0) {
        const playerID = lastPlayedHand.playerId;
        // Ensure gPlayersInfo is an array before using find
        const playerInfo = Array.isArray(gPlayersInfo) ? gPlayersInfo.find(p => p.id === playerID) : null;
        const playerName = playerInfo ? playerInfo.name : playerID; // Use name if available, fallback to ID
        const handTypeString = lastPlayedHand.handTypeString ? lastPlayedHand.handTypeString.replace(/([A-Z])/g, ' $1').trim() : "Hand";
        
        lastPlayedHandInfoDiv.innerHTML = `Played by: ${playerName} (${handTypeString})`;

        lastPlayedHand.cards.sort((a, b) => { 
            if (a.rank !== b.rank) {
                return a.rank - b.rank;
            }
            return a.suit - b.suit;
        });
        console.log("lastPlayedHand", lastPlayedHand);
        lastPlayedHand.cards.forEach(card => {
            const svgRank = rankMap[card.rank];
            const svgSuit = suitMap[card.suit];
            const cardId = `${svgSuit}_${svgRank}`;

            const cardSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
            cardSvg.classList.add('card');
            cardSvg.setAttribute('data-rank', String(card.rank));
            cardSvg.setAttribute('data-suit', String(card.suit));
            cardSvg.setAttribute('xmlns', "http://www.w3.org/2000/svg");
            cardSvg.setAttribute('xmlns:xlink', 'http://www.w3.org/1999/xlink');
            cardSvg.setAttribute('transform', 'scale(0.7)');

            const useElement = document.createElementNS('http://www.w3.org/2000/svg', 'use');
            useElement.setAttribute('href', `cards.svg#${cardId}`);

            cardSvg.appendChild(useElement);
            lastPlayedDiv.appendChild(cardSvg);
        });
    } else {
        lastPlayedHandInfoDiv.innerHTML = "Table is clear.";
        // Display a card back if the table is clear
        const cardSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        cardSvg.classList.add('card'); // Use general card styling
        // cardSvg.classList.add('table-clear-card-back'); // Or a more specific class if needed
        cardSvg.setAttribute('xmlns', "http://www.w3.org/2000/svg");
        cardSvg.setAttribute('xmlns:xlink', 'http://www.w3.org/1999/xlink');
        cardSvg.setAttribute('transform', 'scale(0.7)'); 
        cardSvg.setAttribute('viewBox', "0 0 169.075 244.64"); // Ensure viewBox is set for proper scaling of the back

        const useElement = document.createElementNS('http://www.w3.org/2000/svg', 'use');
        useElement.setAttribute('href', `#back`); // Assumes #back is defined in your cards.svg for the card back
        useElement.setAttribute('fill', 'gray'); // Example fill, can be styled via CSS too
        
        cardSvg.appendChild(useElement);
        lastPlayedDiv.appendChild(cardSvg);
    }
}

function updateTurnInfo(currentPlayerName, passCount, currentPlayerIdFromGame) {
    const turnInfoDiv = document.getElementById('turn-info');
    if (turnInfoDiv) {
        turnInfoDiv.innerHTML = `Current Turn: <strong>${currentPlayerName || "N/A"}</strong>`;
    } else {
        console.error("Element with ID 'turn-info' not found.");
    }

    const passCountInfoDiv = document.getElementById('pass-count-info');
    if (passCountInfoDiv) {
        passCountInfoDiv.textContent = `Pass Count: ${passCount !== undefined ? passCount.toString() : "0"}`;
    } else {
        console.error("Element with ID 'pass-count-info' not found.");
    }
    // Highlighting of current player is handled in updatePlayersArea
}

function updatePlayersArea(playersInfo, yourPlayerId, currentPlayerIdFromGame) {
    if (!playersAreaDiv) return;
    playersAreaDiv.innerHTML = '<h2>Players</h2>'; 

    if (playersInfo && Array.isArray(playersInfo)) {
        playersInfo.forEach(player => {
            const playerDiv = document.createElement('div');
            playerDiv.classList.add('player-info-entry'); 
            
            if (player.id === yourPlayerId) {
                playerDiv.classList.add('is-you');
            }
            if (player.id === currentPlayerIdFromGame) {
                playerDiv.classList.add('current-turn');
            }

            const textInfoDiv = document.createElement('div');
            textInfoDiv.style.display = 'inline-block'; // To allow pass indicator to float next to it
            textInfoDiv.innerHTML = `
                <p><strong>${player.name} (${player.id})</strong></p>
            `;
            playerDiv.appendChild(textInfoDiv);

            // Add PASS indicator if player.hasPassed is true
            if (player.hasPassed) {
                const passIndicator = document.createElement('span');
                passIndicator.classList.add('pass-indicator');
                passIndicator.textContent = 'PASS';
                playerDiv.appendChild(passIndicator);
            }

            const cardBacksDiv = document.createElement('div');
            cardBacksDiv.classList.add('opponent-card-backs-display');
            // Clear float if pass indicator was added
            if (player.hasPassed) {
                cardBacksDiv.style.clear = 'right';
            }

            for (let i = 0; i < player.cardCount; i++) {
                const cardSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
                cardSvg.classList.add('opponent-card-back'); 
                cardSvg.setAttribute('viewBox', "0 0 169.075 244.64"); 

                const useElement = document.createElementNS('http://www.w3.org/2000/svg', 'use');
                useElement.setAttribute('href', `#back`);
                
                if (player.id === yourPlayerId) {
                    useElement.setAttribute('fill', 'blue'); 
                } else {
                    useElement.setAttribute('fill', 'orange');
                }
                
                cardSvg.appendChild(useElement);
                cardBacksDiv.appendChild(cardSvg);
            }
            playerDiv.appendChild(cardBacksDiv);
            playersAreaDiv.appendChild(playerDiv);
        });
    }
}

function updateGameMessages(messageText, isError = false, targetArea = 'general') {
    let targetDiv;
    if (targetArea === 'action') {
        targetDiv = playerActionMessagesDiv;
    } else {
        targetDiv = gameMessagesDiv;
    }

    if (!targetDiv) return;

    // Clear the target div before adding a new message to prevent accumulation if it's an action message
    // or if we want general messages to also be singular.
    // For now, let's only clear action messages this way, general messages can accumulate up to a limit.
    if (targetArea === 'action') {
        targetDiv.innerHTML = '';
    }

    const messageEl = document.createElement('p');
    messageEl.textContent = messageText;
    if (isError) {
        messageEl.style.color = 'red'; // Already styled by inline style in HTML for player-action-messages, but good for general errors too
    }

    if (targetArea === 'general') {
        // Add to the top of game messages and limit
        if (targetDiv.firstChild) {
            targetDiv.insertBefore(messageEl, targetDiv.firstChild);
        } else {
            targetDiv.appendChild(messageEl);
        }
        const maxMessages = 10;
        while (targetDiv.children.length > maxMessages) {
            targetDiv.removeChild(targetDiv.lastChild);
        }
    } else {
        // For action messages, just set the content
        targetDiv.appendChild(messageEl);
    }
}

function getSelectedCards() {
    const selectedCards = [];
    const cardElements = playerHandDiv.querySelectorAll('.card.selected');
    cardElements.forEach(cardElement => {
        selectedCards.push({
            Rank: parseInt(cardElement.dataset.rank, 10),
            Suit: parseInt(cardElement.dataset.suit, 10)
        });
    });
    return selectedCards;
}

function clearSelectedCards() {
    const cardElements = playerHandDiv.querySelectorAll('.card.selected');
    cardElements.forEach(cardElement => {
        cardElement.classList.remove('selected');
    });
}

function sendPlayAction() {
    const selectedCards = getSelectedCards();
    if (selectedCards.length === 0) {
        updateGameMessages('No cards selected to play.', true, 'action');
        return;
    }
    // If auto-pass was enabled, a play action should cancel it.
    if (gAutoPassEnabled) {
        gAutoPassEnabled = false;
        updatePassButtonAppearance();
    }
    const message = {
        type: "playCards",
        cards: selectedCards
    };
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(message));
        clearSelectedCards();
    } else {
        updateGameMessages('Not connected to server. Cannot play cards.', true, 'action');
    }
}

function sendPassAction() {
    const message = { type: "passTurn" };
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(message));
        // If this pass was triggered by auto-pass (gAutoPassEnabled would be true),
        // it should be reset. This happens here or in the gameState logic that called it.
        // For safety, ensure it's reset if this function is ever called while it's true.
        if (gAutoPassEnabled) {
            gAutoPassEnabled = false;
            // updatePassButtonAppearance(); // Will be handled by next gameState or explicitly where needed
        }
    } else {
        updateGameMessages('Not connected to server. Cannot pass.', true, 'action');
    }
}

function handlePassButtonClick() {
    if (passTurnButton && passTurnButton.disabled) return; // Do nothing if button is disabled

    if (gIsMyTurn) {
        // If it's my turn and I click pass, disable any pending auto-pass
        if (gAutoPassEnabled) {
            gAutoPassEnabled = false;
            // updatePassButtonAppearance(); // sendPassAction will be called, then gameState update will fix appearance
        }
        sendPassAction();
    } else {
        // If not my turn, toggle auto-pass
        gAutoPassEnabled = !gAutoPassEnabled;
    }
    updatePassButtonAppearance(); // Update immediately after toggle or before manual pass
}

function updatePassButtonAppearance() {
    if (!passTurnButton) return;
    if (gAutoPassEnabled) {
        passTurnButton.textContent = 'Cancel Auto-Pass';
        passTurnButton.classList.add('auto-pass-active');
    } else {
        passTurnButton.textContent = 'Pass Turn';
        passTurnButton.classList.remove('auto-pass-active');
    }
    // The disabled state is handled by the gameState logic based on game over or connection state.
}

function sendChatMessage() {
    if (chatInput.value.trim() === '') return;
    if (socket && socket.readyState === WebSocket.OPEN) {
        const message = { type: 'chat', content: chatInput.value };
        socket.send(JSON.stringify(message));
        chatInput.value = '';
    } else {
        updateGameMessages('Not connected to server. Cannot send chat.', true, 'general');
    }
}

chatInput.addEventListener('keypress', function(event) {
    if (event.key === 'Enter') {
        sendChatMessage();
    }
});

async function loadSvgSprite() {
    try {
        const response = await fetch('cards.svg');
        if (!response.ok) {
            throw new Error(`Failed to fetch cards.svg: ${response.status} ${response.statusText}`);
        }
        const svgText = await response.text();
        const spriteContainer = document.createElement('div');
        spriteContainer.style.display = 'none'; // Hide the container
        spriteContainer.innerHTML = svgText;
        document.body.appendChild(spriteContainer);
        console.log('SVG sprite loaded and injected.');
    } catch (error) {
        console.error('Error loading SVG sprite:', error);
        updateGameMessages('Error loading card images. Game may not display correctly.', true, 'general');
    }
}

function sendNewGameRequest() {
    if (socket && socket.readyState === WebSocket.OPEN) {
        const message = { type: "newGame" };
        socket.send(JSON.stringify(message));
        console.log("Sent newGame request to server.");
    } else {
        updateGameMessages('Not connected to server. Cannot start a new game.', true, 'general');
    }
}

function sendAliasToServer(alias) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        const message = { type: "setAlias", alias: alias };
        socket.send(JSON.stringify(message));
        console.log("Sent setAlias message with alias:", alias);
        attemptAudioUnlock(); // Attempt to unlock audio when alias is sent
    } else {
        console.error("Cannot send alias, WebSocket is not open.");
        // Handle this case, maybe retry or show error
        updateGameMessages('Error: Could not send alias to server.', true, 'general');
    }
}

function attemptAudioUnlock() {
    if (gAudioUnlocked) return; // Already unlocked

    let unlockedCount = 0;
    let  expectedUnlockCount = 0;

    const tryUnlock = (audioElement, soundName) => {
        if (!audioElement) return;
        expectedUnlockCount++;
        const promise = audioElement.play();
        if (promise !== undefined) {
            promise.then(_ => {
                audioElement.pause();
                audioElement.currentTime = 0;
                unlockedCount++;
                console.log(`Audio for ${soundName} unlocked.`);
                if (unlockedCount === expectedUnlockCount) {
                    gAudioUnlocked = true;
                    console.log("All audio elements unlocked by user interaction.");
                }
            }).catch(error => {
                console.warn(`Audio unlock attempt failed for ${soundName}:`, error);
            });
        } else {
            // For browsers that don't return a promise or if play() doesn't work synchronously here
            // We might assume it worked or log a specific warning.
            // For simplicity, we are relying on promise-based play.
        }
    };

    tryUnlock(audioDingTurn, "ding-turn");
    tryUnlock(audioDingPlayed, "ding-played");

    // If no audio elements were found to unlock, and expectedUnlockCount is 0, 
    // we might set gAudioUnlocked to true if there are no sounds to manage.
    // However, for this game, we expect sounds.
}

if (submitAliasButton && aliasInput && aliasModal && aliasErrorP) {
    submitAliasButton.onclick = function() {
        const alias = aliasInput.value.trim();
        if (alias === "") {
            aliasErrorP.textContent = "Please enter a name.";
            return;
        }
        if (alias.length > 20) { // Example validation
            aliasErrorP.textContent = "Name too long (max 20 chars).";
            return;
        }
        aliasErrorP.textContent = ""; // Clear error

        localStorage.setItem('playerAlias', alias);
        sendAliasToServer(alias);
        aliasModal.style.display = 'none';
        document.getElementById('main-container').style.display = 'block'; // Show game now
    };
}

// Load SVG sprite first, then connect WebSocket
loadSvgSprite().then(() => {
    connectWebSocket();
    // Add event listeners for sort buttons after DOM is ready and WebSocket might connect
    document.getElementById('sort-by-rank-button').addEventListener('click', sortHandByRank);
    document.getElementById('sort-by-suit-button').addEventListener('click', sortHandBySuit);
    if (passTurnButton) {
        passTurnButton.addEventListener('click', handlePassButtonClick);
    }
});

function displayGameOver(roundWinnerId, currentScores, playersInfoFromServer, isMatchOver, overallMatchWinnerId, roundScoresHistory) {
    const gameOverScreen = document.getElementById('game-over-screen');
    const winnerAnnouncement = document.getElementById('winner-announcement');
    const scoresListUl = document.getElementById('player-scores'); // Existing UL
    const newGameButton = document.getElementById('new-game-button');
    const gameOverTitle = gameOverScreen.querySelector('h2');

    if (!gameOverScreen || !winnerAnnouncement || !scoresListUl || !newGameButton || !gameOverTitle) {
        console.error("Game Over screen elements not found!");
        return;
    }

    let winnerName = "N/A";
    if (isMatchOver) {
        gameOverTitle.textContent = "Match Over!";
        const overallWinnerInfo = playersInfoFromServer.find(p => p.id === overallMatchWinnerId);
        winnerName = overallWinnerInfo ? overallWinnerInfo.name : (overallMatchWinnerId || "N/A");
        winnerAnnouncement.textContent = `Overall Winner: ${winnerName}! Target score was ${gTargetScore}.`;
        newGameButton.textContent = "New Match";

        // Replace UL with a score table for match over
        let scoreTableHTML = '<table class="score-table game-over-score-table"><caption>Final Scores</caption><thead><tr><th>Player</th>';
        const numRoundsPlayed = roundScoresHistory ? roundScoresHistory.length : 0;
        for (let i = 1; i <= Math.max(numRoundsPlayed, 1); i++) {
            scoreTableHTML += `<th>R${i}</th>`;
        }
        scoreTableHTML += '<th>Total</th></tr></thead><tbody>';

        if (playersInfoFromServer && playersInfoFromServer.length > 0) {
            playersInfoFromServer.forEach(player => {
                if (!player || !player.id) return;
                scoreTableHTML += `<tr><td>${player.name || player.id}</td>`;
                for (let i = 0; i < Math.max(numRoundsPlayed, 1); i++) {
                    const roundScore = (roundScoresHistory && roundScoresHistory[i] && roundScoresHistory[i][player.id] !== undefined) ? roundScoresHistory[i][player.id] : '-';
                    scoreTableHTML += `<td>${roundScore}</td>`;
                }
                scoreTableHTML += `<td>${currentScores[player.id] !== undefined ? currentScores[player.id] : 0}</td>`;
                scoreTableHTML += '</tr>';
            });
        }
        scoreTableHTML += '</tbody></table>';
        scoresListUl.innerHTML = scoreTableHTML; // Replace UL content with table
        scoresListUl.className = ''; // Remove any UL specific classes if any

    } else { // Round over, not match over
        gameOverTitle.textContent = "Round Over!";
        const roundWinnerInfo = playersInfoFromServer.find(p => p.id === roundWinnerId);
        winnerName = roundWinnerInfo ? roundWinnerInfo.name : (roundWinnerId || "N/A");
        winnerAnnouncement.textContent = `Winner of Round ${gRoundNumber}: ${winnerName}!`;
        newGameButton.textContent = "Next Round";

        // Keep using UL for simple round scores, or update to simple table if preferred later
        scoresListUl.innerHTML = ''; // Clear previous scores
        if (currentScores && Object.keys(currentScores).length > 0) {
            playersInfoFromServer.forEach(player => {
                const playerName = player.name || player.id;
                const score = currentScores[player.id] !== undefined ? currentScores[player.id] : "N/A";
                const li = document.createElement('li');
                li.textContent = `${playerName}: ${score}`;
                scoresListUl.appendChild(li);
            });
        }
    }
    gameOverScreen.style.display = 'flex';
}

// The actual modification for socket.onmessage would be more complex as it's a large function.
// For now, this updateLastPlayed function assumes gPlayersInfo might exist.
// If not, it gracefully falls back to PlayerID.

// Make sure to initialize gPlayersInfo in your main gameState processing logic in socket.onmessage:
// In socket.onmessage, when handling "gameState":
// if (message.playersInfo) {
//   gPlayersInfo = message.playersInfo; // Store for lookup
//   updatePlayersArea(message.playersInfo, message.yourPlayerId, message.currentPlayerId);
// } 

// --- New Sorting Functions (Internal and Public) ---
function sortHandByRankInternal() {
    if (!gCurrentPlayerHand) return;
    gCurrentPlayerHand.sort((a, b) => { 
        if (a.rank !== b.rank) {
            return a.rank - b.rank;
        }
        return a.suit - b.suit; // Secondary sort by suit
    });
}

function sortHandBySuitInternal() {
    if (!gCurrentPlayerHand) return;
    gCurrentPlayerHand.sort((a, b) => {
        if (a.suit !== b.suit) {
            return a.suit - b.suit;
        }
        return a.rank - b.rank; // Secondary sort by rank
    });
}

function sortHandByRank() {
    gCurrentSortPreference = 'rank';
    sortHandByRankInternal();
    updatePlayerHand(gCurrentPlayerHand, gYourPlayerId); // Re-render the sorted hand
}

function sortHandBySuit() {
    gCurrentSortPreference = 'suit';
    sortHandBySuitInternal();
    updatePlayerHand(gCurrentPlayerHand, gYourPlayerId); // Re-render the sorted hand
}
// --- End of New Sorting Functions ---

// Rules Modal
const rulesModal = document.getElementById('rules-modal');
const rulesBtn = document.getElementById('rules-button');
const rulesCloseBtn = document.querySelector('#rules-modal .modal-close-button'); // Assuming close button has this class within the modal

if (rulesBtn && rulesModal && rulesCloseBtn) {
    rulesBtn.onclick = function() {
        rulesModal.style.display = "flex"; // Or "block", depending on your CSS for modal display
    }

    rulesCloseBtn.onclick = function() {
        rulesModal.style.display = "none";
    }

    window.onclick = function(event) {
        if (event.target == rulesModal) {
            rulesModal.style.display = "none";
        }
    }
}

function updateScoreTable(players, roundScoresHistory, currentRoundNum, totalScores, targetScore) {
    const scoreTableArea = document.getElementById('score-table-area');
    if (!scoreTableArea) return;

    let tableHTML = '<table class="score-table">';
    tableHTML += '<caption>Match Scores (Target: ' + targetScore + ')</caption>';
    // Header Row
    tableHTML += '<thead><tr><th>Player</th>';
    const numRoundsPlayed = roundScoresHistory.length;
    for (let i = 1; i <= Math.max(numRoundsPlayed, 1); i++) { // Always show at least Round 1, even if no history yet
        tableHTML += `<th>Round ${i}</th>`;
    }
    tableHTML += '<th>Total</th></tr></thead>';

    // Body Rows
    tableHTML += '<tbody>';
    if (players && players.length > 0) {
        players.forEach(player => {
            if (!player || !player.id) return; // Skip if player info is incomplete
            tableHTML += `<tr><td>${player.name || player.id}</td>`;
            for (let i = 0; i < Math.max(numRoundsPlayed,1) ; i++) {
                const roundScore = (roundScoresHistory[i] && roundScoresHistory[i][player.id] !== undefined) ? roundScoresHistory[i][player.id] : '-';
                tableHTML += `<td>${roundScore}</td>`;
            }
            tableHTML += `<td>${totalScores[player.id] !== undefined ? totalScores[player.id] : 0}</td>`;
            tableHTML += '</tr>';
        });
    }
    tableHTML += '</tbody></table>';

    scoreTableArea.innerHTML = tableHTML;
} 