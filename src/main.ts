import svgSpriteContent from './assets/cards.svg?raw';
import { Card, PlayerInfo, PlayedHand, ServerMessage } from './types';
import * as domElements from './domElements';
import * as constants from './constants';
import { createApp } from 'vue';
import App from './App.vue';

// WebSocket and Game State Globals
let socket: WebSocket | null = null;
let gCurrentPlayerHand: Card[] = [];
let gYourPlayerId: string | null = null;
let gPlayersInfo: PlayerInfo[] = [];
let gLastPlayedHand: PlayedHand | null = null;
let gCurrentPlayerId: string | null = null;
let gPassCount: number = 0;
let gIsGameOver: boolean = false;
let gScores: { [playerId: string]: number } = {};
let gRoundScoresHistory: { [playerId: string]: number }[] = [];
let gRoundNumber: number = 1;
let gTargetScore: number = 100;
let gIsMatchOver: boolean = false;

// UI State Globals
let gAutoPassEnabled: boolean = false;
let gIsMyTurn: boolean = false;
let gPreviousCurrentPlayerId: string | null = null;
let gAudioUnlocked: boolean = false;
export let gCurrentSortPreference: 'rank' | 'suit' = 'rank';

// Setter functions for global state, to be called by App.vue
export function setPlayersInfo(newPlayersInfo: readonly PlayerInfo[]) {
    gPlayersInfo = newPlayersInfo ? [...newPlayersInfo] : [];
}
export function setYourPlayerId(id: string | null) { gYourPlayerId = id; }
export function setCurrentPlayerId(id: string | null) { gCurrentPlayerId = id; }
export function setIsMyTurn(isTurn: boolean) { gIsMyTurn = isTurn; }
export function setRoundNumber(roundNum: number) { gRoundNumber = roundNum; }
export function setTargetScore(score: number) { gTargetScore = score; }
export function setIsMatchOver(isOver: boolean) { gIsMatchOver = isOver; }
export function setScores(newScores: { readonly [playerId: string]: number } | null ) {
    gScores = newScores ? { ...newScores } : {}; // Create mutable copy
}
export function setRoundScoresHistory(history: readonly { readonly [playerId: string]: number }[] | null) {
    gRoundScoresHistory = history ? history.map(item => ({ ...item })) : [];
}
export function setLastPlayedHand(hand: PlayedHand | null) {
    // Ensure deep copy for nested cards array if present
    gLastPlayedHand = hand ? { ...hand, cards: hand.cards ? [...hand.cards] : [] } : null;
}
export function setPassCount(count: number) { gPassCount = count; }
export function setIsGameOver(isOver: boolean) { gIsGameOver = isOver; }
export function setAutoPassEnabled(isEnabled: boolean) { gAutoPassEnabled = isEnabled; }
export function setCurrentPlayerHand(hand: readonly Card[] | null) {
    gCurrentPlayerHand = hand ? [...hand] : [];
}
export function setPreviousCurrentPlayerId(id: string | null) { gPreviousCurrentPlayerId = id; }
// gCurrentSortPreference is updated by sortHandByRank/Suit, which are already exported
// gAudioUnlocked is updated by attemptAudioUnlock, not directly by App.vue state messages

function connectWebSocket() {
    socket = new WebSocket('ws://' + window.location.host + '/ws');

    socket.onopen = function(event: Event) {
        console.log("WebSocket connection established.");

        const storedAlias = localStorage.getItem('playerAlias');
        if (storedAlias) {
            console.log("Found stored alias:", storedAlias);
            sendAliasToServer(storedAlias);
            if (domElements.aliasModal) domElements.aliasModal.style.display = 'none';
            if (domElements.mainContainer) domElements.mainContainer.style.display = 'block';
        } else {
            console.log("No stored alias found. Displaying modal.");
            if (domElements.aliasModal) domElements.aliasModal.style.display = 'flex';
            if (domElements.mainContainer) domElements.mainContainer.style.display = 'none';
        }
    };

    socket.onmessage = function(event: MessageEvent) {
        console.log("Message from server (main.ts - will be IGNORED by switch): ", event.data);
        const parsedData = JSON.parse(event.data);

        if (parsedData && typeof parsedData === 'object' && parsedData.type && typeof parsedData.type === 'string') {
            // const message = parsedData as ServerMessage;

            // Clear previous game messages on any new message from server that isn't an error
            // COMMENTED OUT - App.vue will handle message effects
            // if (message.type !== 'error') {
            //     if (domElements.playerActionMessagesDiv) domElements.playerActionMessagesDiv.innerHTML = '';
            // }

            // COMMENTED OUT - App.vue's watcher now handles message processing logic
            /*
            switch (message.type) {
                case "gameState":
                    // ... (all original gameState processing logic)
                    break;
                case "chat":
                    // ... (all original chat processing logic)
                    break;
                case "error":
                    // ... (all original error processing logic)
                    break;
                default:
                    console.log("Unknown message type received from server (main.ts - IGNORED): ", message);
                    break;
            }
            */
        } else {
            console.error("Received malformed message from server (main.ts - IGNORED):", parsedData);
            // updateGameMessages("Received malformed message from server.", true, 'general'); // Also commented out
        }
    };

    socket.onclose = function(event: CloseEvent) {
        console.log("WebSocket is closed now.");
        const turnInfoP = document.getElementById('turn-info') as HTMLParagraphElement | null;
        if (turnInfoP) turnInfoP.textContent = "Disconnected. Refresh to rejoin.";
    };

    socket.onerror = function(error: Event) {
        console.error("WebSocket error: ", error);
        updateGameMessages("WebSocket error. Check console and refresh to try again.", true, 'general');
    };
}

function updatePlayerHand(cardsToDisplay: readonly Card[], yourId: string | null) {
    // Store current selections before clearing
    const previouslySelectedCards: { rank: string | undefined, suit: string | undefined }[] = [];
    if (domElements.playerHandDiv) {
        domElements.playerHandDiv.querySelectorAll('.card.selected').forEach(selectedCardElement => {
            const htmlElement = selectedCardElement as HTMLElement; // Assertion
            previouslySelectedCards.push({
                rank: htmlElement.dataset.rank,
                suit: htmlElement.dataset.suit
            });
        });
    }

    if (domElements.playerHandDiv) domElements.playerHandDiv.innerHTML = ''; // Clear current hand

    if (!cardsToDisplay || cardsToDisplay.length === 0) {
        if (domElements.playerHandDiv) domElements.playerHandDiv.innerHTML = '<p>No cards in hand.</p>';
        return;
    }

    cardsToDisplay.forEach(card => {
        const cardValue = constants.rankMap[card.rank];
        const suitValue = constants.suitMap[card.suit];

        const svgElement = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svgElement.setAttribute('class', 'card');
        svgElement.setAttribute('data-rank', String(card.rank));
        svgElement.setAttribute('data-suit', String(card.suit));
        svgElement.setAttribute('transform', 'scale(1)'); 

        const useElement = document.createElementNS('http://www.w3.org/2000/svg', 'use');
        useElement.setAttributeNS('http://www.w3.org/1999/xlink', 'xlink:href', `#${suitValue}_${cardValue}`);
        
        svgElement.appendChild(useElement);

        // Re-apply selection if this card was previously selected
        if (previouslySelectedCards.some(selCard => selCard.rank == String(card.rank) && selCard.suit == String(card.suit))) {
            svgElement.classList.add('selected');
        }

        svgElement.addEventListener('click', () => {
            if (!gIsGameOver && !gIsMatchOver) {
                svgElement.classList.toggle('selected');
            }
        });
        if (domElements.playerHandDiv) domElements.playerHandDiv.appendChild(svgElement);
    });
}

function updateLastPlayed(lastPlayedHand: PlayedHand | null, playersInfo: readonly PlayerInfo[]) {
    const targetDiv = domElements.lastPlayedDiv;
    if (!targetDiv) { 
        console.error("lastPlayedDiv is null and is required for updateLastPlayed.");
        return; 
    }
    targetDiv.innerHTML = ''; // Clear previous

    const lastPlayedHandInfoP = document.getElementById('last-played-hand-info');

    if (!lastPlayedHand || !lastPlayedHand.cards || lastPlayedHand.cards.length === 0) {
        const svgElement = document.createElementNS("http://www.w3.org/2000/svg", "svg");
        svgElement.setAttribute("class", "card");
        svgElement.setAttribute("viewBox", "0 0 169.075 244.64");
        const useElement = document.createElementNS("http://www.w3.org/2000/svg", "use");
        useElement.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", "#card_back"); 
        useElement.setAttribute('fill', 'gray');
        svgElement.appendChild(useElement);
        targetDiv.appendChild(svgElement); // Safe due to the check above
        
        if (lastPlayedHandInfoP) lastPlayedHandInfoP.textContent = 'Table is clear.';
        return;
    }

    lastPlayedHand.cards.forEach(card => {
        const cardValue = constants.rankMap[card.rank];
        const suitValue = constants.suitMap[card.suit];

        const svgElement = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svgElement.setAttribute("class", "card");
        svgElement.setAttribute("data-rank", String(card.rank));
        svgElement.setAttribute("data-suit", String(card.suit));
        svgElement.setAttribute("viewBox", "0 0 169.075 244.64");

        const useElement = document.createElementNS("http://www.w3.org/2000/svg", "use");
        useElement.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", `#${suitValue}_${cardValue}`);
        svgElement.appendChild(useElement);
        targetDiv.appendChild(svgElement); // Safe
    });

    const player = playersInfo.find(p => p.id === lastPlayedHand.playerId);
    const playerName = player ? player.name : lastPlayedHand.playerId;
    const handDescription = `${playerName} played: ${lastPlayedHand.handTypeString}`;
    if (lastPlayedHandInfoP) lastPlayedHandInfoP.textContent = handDescription;
}

function updateTurnInfo(currentPlayerName: string | undefined, passCount: number, currentPlayerIdFromGame: string | null) {
    const turnInfoP = document.getElementById('turn-info') as HTMLParagraphElement | null;
    const passCountInfoP = document.getElementById('pass-count-info') as HTMLParagraphElement | null;

    if (gIsGameOver || gIsMatchOver) {
        if (turnInfoP) turnInfoP.textContent = "Game Over";
        if (passCountInfoP) passCountInfoP.textContent = "";
        return;
    }

    // The "YOUR TURN!" text is now removed, using background highlight instead.
    if (turnInfoP) turnInfoP.textContent = currentPlayerName ? `Current Turn: ${currentPlayerName}` : "Waiting for player...";
    if (passCountInfoP) passCountInfoP.textContent = passCount > 0 ? `${passCount} player(s) passed.` : "";
}

function updatePlayersArea(playersInfo: readonly PlayerInfo[], yourPlayerId: string | null, currentPlayerIdFromGame: string | null) {
    const targetDiv = domElements.playersAreaDiv;
    if (!targetDiv) {
        console.error("playersAreaDiv is null and is required for updatePlayersArea.");
        return; 
    }
    targetDiv.innerHTML = ''; // Clear existing entries

    playersInfo.forEach(player => {
        const playerDiv = document.createElement('div');
        playerDiv.className = 'player-info-entry';
        if (player.id === currentPlayerIdFromGame) {
            playerDiv.classList.add('current-turn');
        }
        if (player.id === yourPlayerId) {
            playerDiv.classList.add('is-you');
        }

        let playerNameDisplay = player.name;
        if (player.id === yourPlayerId) {
            playerNameDisplay += " (You)";
        }

        const infoP = document.createElement('p');
        infoP.textContent = `${playerNameDisplay} - Cards: ${player.cardCount}`;
        playerDiv.appendChild(infoP);

        if (player.hasPassed) {
            const passIndicator = document.createElement('span');
            passIndicator.className = 'pass-indicator';
            passIndicator.textContent = 'PASS';
            playerDiv.appendChild(passIndicator);
        }

        const cardBacksDiv = document.createElement('div');
        cardBacksDiv.className = 'opponent-card-backs-display';
        for (let i = 0; i < player.cardCount; i++) {
            const svgElement = document.createElementNS("http://www.w3.org/2000/svg", "svg");
            svgElement.setAttribute("class", "opponent-card-back");
            svgElement.setAttribute("viewBox", "0 0 169.075 244.64");
            const useElement = document.createElementNS("http://www.w3.org/2000/svg", "use");
            useElement.setAttributeNS("http://www.w3.org/1999/xlink", "xlink:href", "#back");
            svgElement.appendChild(useElement);
            cardBacksDiv.appendChild(svgElement);

            if (player.id === yourPlayerId) {
                useElement.setAttribute('fill', 'blue'); 
            } else {
                useElement.setAttribute('fill', 'orange');
            }
        }
        playerDiv.appendChild(cardBacksDiv);
        targetDiv.appendChild(playerDiv);
    });
}

function updateGameMessages(messageText: string, isError: boolean = false, targetArea: 'general' | 'action' = 'general') {
    let targetDiv: HTMLDivElement | null;
    if (targetArea === 'action') {
        targetDiv = domElements.playerActionMessagesDiv;
    } else { 
        targetDiv = domElements.gameMessagesDiv; 
    }

    if (!targetDiv) {
        console.error("Target message div not found for area:", targetArea);
        return;
    }

    // Clear previous messages only in the targeted div
    // targetDiv.innerHTML = ''; 
    // Decided against auto-clearing general game messages on each new message, let them accumulate or be cleared specifically.
    // Action messages are cleared on each new non-error server message (see socket.onmessage).

    const messageP = document.createElement('p');
    messageP.textContent = messageText;
    if (isError) {
        messageP.style.color = 'red';
    } else {
        messageP.style.color = 'blue'; 
    }
    targetDiv.appendChild(messageP);
}

function getSelectedCards(): Card[] {
    const selectedCards: Card[] = [];
    if (domElements.playerHandDiv) {
        domElements.playerHandDiv.querySelectorAll('.card.selected').forEach(cardElement => {
            const htmlElement = cardElement as HTMLElement; // Assertion
            if (htmlElement.dataset.rank && htmlElement.dataset.suit) {
                selectedCards.push({
                    rank: parseInt(htmlElement.dataset.rank),
                    suit: parseInt(htmlElement.dataset.suit)
                });
            }
        });
    }
    return selectedCards;
}

function sendPlayAction() {
    if (gIsGameOver || gIsMatchOver) {
        updateGameMessages("The game/match is over.", true, 'action');
        return;
    }
    if (!gIsMyTurn) {
        updateGameMessages("Not your turn.", true, 'action');
        return;
    }
    const selectedCards = getSelectedCards();
    if (selectedCards.length === 0) {
        updateGameMessages("No cards selected.", true, 'action');
        return;
    }
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: "playCards", cards: selectedCards }));
        // clearSelectedCards(); // Clear selection after sending, or wait for server to confirm?
                                // Let server state dictate hand, so selection will clear naturally on update.
    } else {
        updateGameMessages("Not connected to server.", true, 'action');
    }
}

function sendPassAction() {
    if (gIsGameOver || gIsMatchOver) {
        updateGameMessages("The game/match is over.", true, 'action');
        return;
    }
    if (!gIsMyTurn) {
        updateGameMessages("Not your turn.", true, 'action');
        return;
    }
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: "passTurn" }));
    } else {
        updateGameMessages("Not connected to server.", true, 'action');
    }
}

function handlePassButtonClick() {
    if (gIsMyTurn && !gIsGameOver && !gIsMatchOver) {
        if (gAutoPassEnabled) {
            gAutoPassEnabled = false; // Toggle off
            updateGameMessages("Auto-pass disabled.", false, 'action');
        }
        sendPassAction(); // Standard pass if not toggling auto-pass
    } else {
        gAutoPassEnabled = !gAutoPassEnabled;
    }
    updatePassButtonAppearance();
}

function updatePassButtonAppearance() {
    if (!domElements.passTurnButton) return;
    if (gAutoPassEnabled) {
        domElements.passTurnButton.classList.add('auto-pass-active');
        domElements.passTurnButton.textContent = "Disable Auto-Pass";
    } else {
        domElements.passTurnButton.classList.remove('auto-pass-active');
        domElements.passTurnButton.textContent = "Pass / Auto-Pass";
    }
}

function sendChatMessage() {
    if (domElements.chatInput && domElements.chatInput.value.trim() !== "") {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: "chat", content: domElements.chatInput.value }));
            domElements.chatInput.value = ''; // Clear input after sending
        } else {
            updateGameMessages("Not connected to server to send chat.", true, 'action');
        }
    }
}

async function loadSvgSprite(): Promise<void> {
    try {
        const spriteContainer = document.getElementById('svg-sprite-container');
        if (spriteContainer) {
            spriteContainer.innerHTML = svgSpriteContent;
        } else {
            console.error('SVG sprite container not found');
        }
    } catch (error) {
        console.error('Failed to load SVG sprite:', error);
    }
}

function sendNewGameRequest() {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: "new_game" }));
        // Re-enable buttons and clear game over screen, though server state should handle most of this
        const gameOverScreen = document.getElementById('game-over-screen') as HTMLDivElement | null;
        if (gameOverScreen) gameOverScreen.style.display = 'none';
        const playButton = document.getElementById('play-selected-button') as HTMLButtonElement | null;
        if (playButton) playButton.disabled = false;
        if (domElements.passTurnButton) domElements.passTurnButton.disabled = false;
        gAutoPassEnabled = false;
        updatePassButtonAppearance();
        if (domElements.gameMessagesDiv) domElements.gameMessagesDiv.innerHTML = '';
        if (domElements.playerActionMessagesDiv) domElements.playerActionMessagesDiv.innerHTML = '';
    } else {
        updateGameMessages("Not connected to server to start a new game.", true, 'general');
    }
}

function sendAliasToServer(alias: string) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: "setAlias", alias: alias }));
    } else {
        // This case should ideally not happen if called from socket.onopen
        // or if alias modal is shown only when disconnected.
        console.error("Socket not open when trying to send alias.");
        if (domElements.aliasErrorP) domElements.aliasErrorP.textContent = "Connection error. Please refresh.";
    }
}

// Attempt to unlock audio context on first user interaction
function attemptAudioUnlock() {
    if (gAudioUnlocked) return;

    const tryUnlock = (audioElement: HTMLAudioElement | null, soundName: string): Promise<boolean> => {
        if (!audioElement) return Promise.resolve(false); // Resolve if element doesn't exist
        return audioElement.play().then(() => {
            audioElement.pause();
            audioElement.currentTime = 0;
            console.log(`${soundName} audio unlocked.`);
            return true;
        }).catch((error: any) => {
            // Not critical if one fails, could be user settings, etc.
            // console.warn(`${soundName} audio unlock failed:`, error.name, error.message);
            if (error.name === 'NotAllowedError') {
                // This is common before user interaction. We don't want to spam logs here.
            } else {
                console.warn(`${soundName} audio unlock test play failed:`, error);
            }
            return false;
        });
    };

    Promise.all([
        tryUnlock(domElements.audioDingTurn, "Turn notification"),
        tryUnlock(domElements.audioDingPlayed, "Card played notification")
    ]).then(results => {
        // If any sound successfully played (even if then paused), we consider audio unlocked.
        if (results.some(unlocked => unlocked)) {
            gAudioUnlocked = true;
            console.log("Audio context unlocked by user interaction.");
            // Remove the global event listener once unlocked
            document.removeEventListener('click', attemptAudioUnlock);
            document.removeEventListener('keydown', attemptAudioUnlock);
        } else {
            // console.log("Audio context still locked after attempts.");
        }
    });
}

function displayGameOver(
    roundWinnerId: string | null, 
    currentScores: { readonly [playerId: string]: number }, 
    playersInfoFromServer: readonly PlayerInfo[], 
    isMatchOver: boolean, 
    overallMatchWinnerId: string | null, 
    roundScoresHistory: { [playerId: string]: number }[]
) {
    const gameOverScreen = document.getElementById('game-over-screen') as HTMLDivElement | null;
    const gameOverContent = document.getElementById('game-over-content') as HTMLDivElement | null;
    const winnerInfoP = document.getElementById('winner-info') as HTMLParagraphElement | null;
    const scoresList = document.getElementById('player-scores') as HTMLUListElement | null; // This is for ROUND scores if match not over
    const newGameButton = document.getElementById('new-game-button') as HTMLButtonElement | null;
    const overallWinnerInfoP = document.getElementById('overall-winner-info') as HTMLParagraphElement | null; // New element for match winner
    const roundScoreTableContainer = document.getElementById('round-score-table-container') as HTMLDivElement | null; // For final scores

    if (!gameOverScreen || !winnerInfoP || !scoresList || !newGameButton || !gameOverContent || !overallWinnerInfoP || !roundScoreTableContainer) {
        console.error('One or more game over screen elements are missing from the DOM.');
        return;
    }

    // Ensure player names are up-to-date for the display
    const getPlayerName = (id: string | null): string => {
        if (!id) return "Unknown";
        const player = playersInfoFromServer.find(p => p.id === id);
        return player ? player.name : id;
    };

    const winnerName = getPlayerName(roundWinnerId);
    const overallWinnerName = getPlayerName(overallMatchWinnerId);

    gameOverContent.style.textAlign = 'center'; // Ensure content is centered
    overallWinnerInfoP.innerHTML = ''; // Clear previous match winner info
    roundScoreTableContainer.innerHTML = ''; // Clear previous final scores table
    scoresList.innerHTML = ''; // Clear previous round scores list

    if (isMatchOver) {
        winnerInfoP.innerHTML = `<h2>Match Over!</h2>`;
        if (overallMatchWinnerId) {
            overallWinnerInfoP.innerHTML = `<h3>Overall Winner: ${overallWinnerName}!</h3>`;
        }
        if (newGameButton) newGameButton.textContent = "Start New Match";
        // Display final scores using the score table component
        const finalScoresTable = createScoreTableDOM(playersInfoFromServer, roundScoresHistory, gRoundNumber, currentScores, gTargetScore, true);
        finalScoresTable.classList.add('game-over-score-table'); // For potential specific styling
        roundScoreTableContainer.appendChild(finalScoresTable);

    } else { // Round over, but match continues
        winnerInfoP.innerHTML = `<h2>Round ${gRoundNumber} Over!</h2>`;
        if (roundWinnerId) {
            overallWinnerInfoP.innerHTML = `<h3>Round Winner: ${winnerName}</h3>`; 
        }
        if (newGameButton) newGameButton.textContent = "Start Next Round";
        
        // Display current scores (total for the match so far)
        const scoresTitle = document.createElement('h4');
        scoresTitle.textContent = "Current Match Scores:";
        scoresList.appendChild(scoresTitle);

        Object.entries(currentScores).forEach(([playerId, score]) => {
            const listItem = document.createElement('li');
            listItem.textContent = `${getPlayerName(playerId)}: ${score}`;
            scoresList.appendChild(listItem);
        });
    }

    // Always show the game over screen with display: flex for centering
    gameOverScreen.style.display = 'flex'; 
}

function sortHandByRank() {
    gCurrentSortPreference = 'rank';
    // This now relies on App.vue to handle the actual sorting logic and re-rendering
    // We can keep this if other non-Vue parts need to trigger a preference change,
    // but right now it seems redundant if PlayerControls is the only trigger.
}

function sortHandBySuit() {
    gCurrentSortPreference = 'suit';
     // Similar to sortHandByRank, this is likely redundant.
}

// Helper function to create and populate the score table DOM element (reusable)
function createScoreTableDOM(
    players: readonly PlayerInfo[], 
    roundScoresHistory: { [playerId: string]: number }[], 
    currentRoundNum: number, 
    totalScores: { readonly [playerId: string]: number }, 
    targetScore: number, 
    isFinal: boolean = false
): HTMLTableElement {
    const table = document.createElement('table');
    table.classList.add('score-table');

    const caption = table.createCaption();
    caption.textContent = isFinal ? `Final Scores (Target: ${targetScore})` : `Scores After Round ${currentRoundNum -1} (Target: ${targetScore})`;

    const thead = table.createTHead();
    const headerRow = thead.insertRow();
    const playerHeaderCell = document.createElement('th');
    playerHeaderCell.textContent = 'Player';
    headerRow.appendChild(playerHeaderCell);

    // Add round number headers
    for (let i = 1; i < currentRoundNum; i++) {
        const roundCell = document.createElement('th');
        roundCell.textContent = `R${i}`;
        headerRow.appendChild(roundCell);
    }
    const totalHeaderCell = document.createElement('th');
    totalHeaderCell.textContent = 'Total';
    headerRow.appendChild(totalHeaderCell);

    const tbody = table.createTBody();
    players.forEach(player => {
        const row = tbody.insertRow();
        const nameCell = row.insertCell();
        nameCell.textContent = player.name;

        for (let i = 0; i < currentRoundNum - 1; i++) {
            const scoreCell = row.insertCell();
            const roundScore = roundScoresHistory[i] ? (roundScoresHistory[i][player.id] || 0) : 0;
            scoreCell.textContent = String(roundScore);
        }

        const totalCell = row.insertCell();
        totalCell.textContent = String(totalScores[player.id] || 0);
    });

    return table;
}

// Main function to update the score table in the UI
function updateScoreTable(
    players: readonly PlayerInfo[], 
    roundScoresHistory: { [playerId: string]: number }[], 
    currentRoundNum: number, 
    totalScores: { readonly [playerId: string]: number }, 
    targetScore: number
) {
    const scoreTableContainer = document.getElementById('score-table-container') as HTMLDivElement | null;
    if (!scoreTableContainer) return;

    scoreTableContainer.innerHTML = ''; // Clear previous table

    if (!players || players.length === 0) {
        scoreTableContainer.textContent = 'No player data for score table.';
        return;
    }
    // Only show table if there's at least one completed round or it's the very start (round 1, no history)
    if (roundScoresHistory.length > 0 || currentRoundNum === 1) {
         const table = createScoreTableDOM(players, roundScoresHistory, currentRoundNum, totalScores, targetScore, false);
         scoreTableContainer.appendChild(table);
    } else {
        scoreTableContainer.textContent = 'Waiting for first round to complete to show scores.';
    }
}

// Export necessary functions for Vue app to call
export {
    updatePlayerHand,
    updateLastPlayed,
    updateTurnInfo,
    updatePlayersArea,
    updateScoreTable,
    displayGameOver,
    updateGameMessages,
    updatePassButtonAppearance,
    sortHandByRank,
    sortHandBySuit
};

document.addEventListener('DOMContentLoaded', () => {
    loadSvgSprite();
    connectWebSocket();

    // Attempt to unlock audio on first interaction
    document.addEventListener('click', attemptAudioUnlock, { once: false });
    document.addEventListener('keydown', attemptAudioUnlock, { once: false });

    // Event Listeners for buttons
    const playSelectedButton = document.getElementById('play-selected-button') as HTMLButtonElement | null;
    if (playSelectedButton) playSelectedButton.addEventListener('click', sendPlayAction);
    
    if (domElements.passTurnButton) {
        domElements.passTurnButton.addEventListener('click', handlePassButtonClick);
    }
    const sendChatButton = document.getElementById('send-chat-button') as HTMLButtonElement | null;
    if (sendChatButton) sendChatButton.addEventListener('click', sendChatMessage);
    
    if (domElements.chatInput) {
        domElements.chatInput.addEventListener('keypress', function(event: KeyboardEvent) {
            if (event.key === 'Enter') {
                sendChatMessage();
            }
        });
    }

    if (domElements.submitAliasButton && domElements.aliasInput && domElements.aliasModal && domElements.aliasErrorP) {
        domElements.submitAliasButton.addEventListener('click', () => {
            const alias = domElements.aliasInput!.value.trim(); 
            if (alias) {
                localStorage.setItem('playerAlias', alias);
                sendAliasToServer(alias);
                domElements.aliasModal!.style.display = 'none'; 
                if (domElements.mainContainer) domElements.mainContainer.style.display = 'block';
                domElements.aliasErrorP!.textContent = ''; 
            } else {
                domElements.aliasErrorP!.textContent = "Alias cannot be empty."; 
            }
        });
        domElements.aliasInput.addEventListener('keypress', function(event: KeyboardEvent) {
            if (event.key === 'Enter') {
                domElements.submitAliasButton!.click(); 
            }
        });
    }

    const newGameBtn = document.getElementById('new-game-button') as HTMLButtonElement | null;
    if (newGameBtn) {
        newGameBtn.addEventListener('click', sendNewGameRequest);
    }

    const rulesModalElement = document.getElementById('rules-modal') as HTMLDivElement | null;
    const openRulesButton = document.getElementById('open-rules-button') as HTMLButtonElement | null;
    const closeRulesButton = document.getElementById('close-rules-modal-button') as HTMLSpanElement | null;

    if (openRulesButton && rulesModalElement) {
        openRulesButton.addEventListener('click', () => { rulesModalElement.style.display = 'flex'; });
    }
    if (closeRulesButton && rulesModalElement) {
        closeRulesButton.addEventListener('click', () => { rulesModalElement.style.display = 'none'; });
    }

    // Sort button listeners
    const sortByRankButton = document.getElementById('sort-by-rank-button') as HTMLButtonElement | null;
    if (sortByRankButton) sortByRankButton.addEventListener('click', sortHandByRank);
    const sortBySuitButton = document.getElementById('sort-by-suit-button') as HTMLButtonElement | null;
    if (sortBySuitButton) sortBySuitButton.addEventListener('click', sortHandBySuit);

    // Close modal if backdrop is clicked
    window.addEventListener('click', (event: MouseEvent) => {
        if (event.target === domElements.aliasModal) {
            // No automatic close for alias modal, must submit or explicitly close if a button existed
        }
        if (event.target === rulesModalElement) {
            if (rulesModalElement) rulesModalElement.style.display = 'none';
        }
        // Note: game-over-screen doesn't have a separate close, only New Game button.
    });

    createApp(App).mount('#app');
}); 