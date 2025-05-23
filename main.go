package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool { // Allow all connections for simplicity
		return true
	},
}

// client represents a single WebSocket connection and its associated player.
// We use a pointer to Player to share the Player state from GameState.
type client struct {
	conn   *websocket.Conn
	player *Player // Reference to the Player struct in the GameState
}

var (
	clients           = make(map[*websocket.Conn]*client)
	clientsMu         sync.Mutex
	gameInstance      *GameState // Global pointer to our single game instance
	gameInstanceMutex sync.Mutex // Mutex to protect gameInstance
)

// Helper function to parse card data received from the client
func parseCardsFromClientData(cardsData interface{}) (Deck, error) {
	cardsArr, ok := cardsData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cards data is not an array")
	}

	var deck Deck
	for _, cardInterface := range cardsArr {
		cardMap, ok := cardInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("card data item is not a map")
		}

		rankVal, rankOk := cardMap["Rank"].(float64) // JSON numbers are float64 by default
		suitVal, suitOk := cardMap["Suit"].(float64)

		if !rankOk || !suitOk {
			return nil, fmt.Errorf("card rank or suit is missing or not a number: RankOk=%v, SuitOk=%v, cardMap=%v", rankOk, suitOk, cardMap)
		}
		deck = append(deck, Card{Rank: Rank(rankVal), Suit: Suit(suitVal)})
	}
	deck.Sort() // Sort the parsed deck to match canonical form for rule engine
	return deck, nil
}

// broadcastGameState sends the current game state to all connected clients.
func broadcastGameState(game *GameState) {
	log.Printf("DEBUG: broadcastGameState called. Game state players: %d, PassCount: %d, GameOver: %v", len(game.Players), game.PassCount, game.IsGameOver) // More concise log

	if game == nil {
		log.Println("ERROR: broadcastGameState called with nil game state")
		return
	}

	var currentPlayerID string
	var currentPlayerName string
	if game.CurrentTurnPlayerIndex >= 0 && game.CurrentTurnPlayerIndex < len(game.Players) {
		currentPlayerID = game.Players[game.CurrentTurnPlayerIndex].ID
		currentPlayerName = game.Players[game.CurrentTurnPlayerIndex].Name
	} else {
		log.Printf("Warning: CurrentTurnPlayerIndex (%d) is out of bounds for players list (len %d)", game.CurrentTurnPlayerIndex, len(game.Players))
		// Assign default/empty values if index is out of bounds
		currentPlayerID = ""
		currentPlayerName = "N/A"
	}

	// Create a temporary list of clients to iterate over to avoid issues if clients map changes during iteration
	// This also allows releasing locks sooner if applicable.
	clientsSnapshot := make([]*client, 0, len(clients))
	clientsMu.Lock() // Use Lock for sync.Mutex
	for _, client := range clients {
		clientsSnapshot = append(clientsSnapshot, client)
	}
	clientsMu.Unlock() // Use Unlock for sync.Mutex

	for _, c := range clientsSnapshot { // Iterate over the snapshot
		clientHand := Deck{}
		playerIDForClient := "Observer"
		if c.player != nil {
			playerIDForClient = c.player.ID
			// Ensure client's player instance is up-to-date from game.Players for hand info
			foundPlayerInGame := false
			for _, gamePlayer := range game.Players {
				if gamePlayer.ID == c.player.ID {
					clientHand = gamePlayer.Hand // Get the most current hand from the game state
					foundPlayerInGame = true
					break
				}
			}
			if !foundPlayerInGame {
				log.Printf("Warning: Client %s player ID %s not found in current game.Players. Sending empty hand.", c.conn.RemoteAddr(), c.player.ID)
			}
		} else {
			// Observer or unassigned client, receives empty hand
		}

		// playerInfoList is constructed for each client payload if it needs to be specific
		// or can be constructed once outside the loop if it's identical for all.
		// For simplicity here, assuming it's general player info.
		playerInfoListPayload := make([]map[string]interface{}, len(game.Players))
		for i, p := range game.Players {
			playerInfoListPayload[i] = map[string]interface{}{
				"id":        p.ID,   // Already camelCase from Player struct tag
				"name":      p.Name, // Already camelCase from Player struct tag
				"cardCount": len(p.Hand),
				"hasPassed": p.HasPassed, // Already camelCase from Player struct tag
			}
		}

		payload := struct {
			Type              string                   `json:"type"`
			Hand              Deck                     `json:"hand"`
			LastPlayedHand    *PlayedHand              `json:"lastPlayedHand"`
			CurrentPlayerID   string                   `json:"currentPlayerId"`
			CurrentPlayerName string                   `json:"currentPlayerName"`
			YourPlayerID      string                   `json:"yourPlayerId"`
			PassCount         int                      `json:"passCount"`
			PlayersInfo       []map[string]interface{} `json:"playersInfo"`
			GameMessage       string                   `json:"gameMessage,omitempty"`
			IsGameOver        bool                     `json:"isGameOver"`
			WinnerID          string                   `json:"winnerId,omitempty"`
			Scores            map[string]int           `json:"scores,omitempty"`

			// New fields for multi-round/match payload
			RoundNumber        int              `json:"roundNumber"`
			TargetScore        int              `json:"targetScore"`
			IsMatchOver        bool             `json:"isMatchOver"`
			OverallWinnerID    string           `json:"overallWinnerId,omitempty"`
			RoundScoresHistory []map[string]int `json:"roundScoresHistory,omitempty"`
		}{
			Type:              "gameState",
			Hand:              clientHand,
			LastPlayedHand:    game.LastPlayedHand,
			CurrentPlayerID:   currentPlayerID,
			CurrentPlayerName: currentPlayerName,
			YourPlayerID:      playerIDForClient,
			PassCount:         game.PassCount,
			PlayersInfo:       playerInfoListPayload,
			GameMessage:       "",
			IsGameOver:        game.IsGameOver,
			WinnerID:          game.WinnerID,
			Scores:            game.Scores,

			// New fields for multi-round/match payload
			RoundNumber:        game.RoundNumber,
			TargetScore:        game.TargetScore,
			IsMatchOver:        game.IsMatchOver,
			OverallWinnerID:    game.OverallWinnerID,
			RoundScoresHistory: game.RoundScoresHistory,
		}

		log.Printf("DEBUG: Preparing payload for client. PlayerIDForClient: %s. Hand size: %d. LastPlayedHand: %v. RoundScoresHistory items: %d", playerIDForClient, len(clientHand), game.LastPlayedHand != nil, len(game.RoundScoresHistory))
		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("FATAL_ERROR Marshalling game state for client %s (player ID %s): %v. Payload: %+v", c.conn.RemoteAddr(), playerIDForClient, err, payload)
			continue
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Printf("FATAL_ERROR writing game state to client %s (player ID %s): %v", c.conn.RemoteAddr(), playerIDForClient, err)
		} else {
			log.Printf("DEBUG: Successfully sent gameState to client %s (player ID %s)", c.conn.RemoteAddr(), playerIDForClient)
		}
	}
	log.Println("Broadcasted game state update.")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	currentWsClient := &client{conn: conn}
	var assignedPlayer *Player

	// Assign player (critical section, uses gameInstanceMutex and clientsMu)
	gameInstanceMutex.Lock()
	clientsMu.Lock()
	if gameInstance != nil && gameInstance.Players != nil {
		for _, p := range gameInstance.Players {
			isSlotTaken := false
			for _, cl := range clients {
				if cl.player == p {
					isSlotTaken = true
					break
				}
			}
			if !isSlotTaken {
				assignedPlayer = p
				currentWsClient.player = p
				clients[conn] = currentWsClient
				break
			}
		}
	}
	clientsMu.Unlock()
	gameInstanceMutex.Unlock() // Unlock after initial player assignment setup

	defer func() {
		var disconnectedPlayerName string
		clientsMu.Lock() // Lock for reading and then modifying clients map
		clientInfo := clients[conn]
		if clientInfo != nil && clientInfo.player != nil {
			disconnectedPlayerName = clientInfo.player.Name
			log.Printf("Player %s (%s) WebSocket disconnecting.", clientInfo.player.ID, disconnectedPlayerName)
			// No gameInstance state modification for player.IsConnected here needed in defer
		}
		delete(clients, conn)
		// Unlock clientsMu before broadcasting, as broadcastMessage will acquire it again.
		// However, the client is already deleted, so the broadcast won't reach them (which is fine).
		// The list of recipients for broadcastMessage is determined when it runs.
		clientsMu.Unlock()

		conn.Close()
		log.Println("Client connection closed and removed:", conn.RemoteAddr())

		// Broadcast player disconnect system message if a player was associated
		if disconnectedPlayerName != "" {
			disconnectionMsg := fmt.Sprintf("%s has disconnected.", disconnectedPlayerName)
			chatPayload := map[string]string{"type": "chat", "sender": "System", "content": disconnectionMsg}
			jsonMsg, _ := json.Marshal(chatPayload)

			// We need to lock clientsMu again for broadcastMessage if it iterates the live map.
			// Or, ensure broadcastMessage uses a snapshot taken while locked.
			// Given broadcastMessage's current implementation, it locks itself.
			broadcastMessage(websocket.TextMessage, jsonMsg, nil) // Sender is nil (System)
		}
	}()

	if assignedPlayer == nil {
		log.Println("No available player slot for new client or game not ready. Disconnecting client:", conn.RemoteAddr())
		errMsg := `{"type": "error", "content": "Sorry, the game is full or not available."}`
		if connErr := conn.WriteMessage(websocket.TextMessage, []byte(errMsg)); connErr != nil {
			log.Printf("Error sending game full message to %s: %v", conn.RemoteAddr(), connErr)
		}
		return // Return directly, defer will handle cleanup
	}

	log.Printf("Client %s connected and assigned to Player %s (%s)", conn.RemoteAddr(), assignedPlayer.ID, assignedPlayer.Name)

	// Broadcast player connection system message
	connectionMsg := fmt.Sprintf("%s has connected.", assignedPlayer.Name)
	chatPayload := map[string]string{"type": "chat", "sender": "System", "content": connectionMsg}
	jsonMsg, _ := json.Marshal(chatPayload)
	// clientsMu.Lock() // broadcastMessage handles its own locking
	broadcastMessage(websocket.TextMessage, jsonMsg, nil)
	// clientsMu.Unlock()

	// Send initial game state to this newly connected player
	log.Printf("DEBUG: About to send initial game state to player %s. Game instance players: %d. Client player ID: %s", assignedPlayer.ID, len(gameInstance.Players), currentWsClient.player.ID)
	gameInstanceMutex.Lock()
	broadcastGameState(gameInstance)
	gameInstanceMutex.Unlock()

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Read error (client %s, player %s): %v", conn.RemoteAddr(), assignedPlayer.ID, err)
			} else {
				log.Printf("Client %s (player %s) initiated disconnect.", conn.RemoteAddr(), assignedPlayer.ID)
			}
			break // Exit loop, defer will clean up client
		}

		var receivedMsg map[string]interface{}
		if err := json.Unmarshal(msgBytes, &receivedMsg); err != nil {
			log.Printf("Error unmarshalling message from Player %s: %v. Message: %s", assignedPlayer.ID, err, string(msgBytes))
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Malformed JSON."}`))
			continue
		}

		msgType, typeOk := receivedMsg["type"].(string)
		if !typeOk {
			log.Printf("Message from Player %s missing 'type'. Msg: %s", assignedPlayer.ID, string(msgBytes))
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Message missing 'type' field."}`))
			continue
		}

		log.Printf("Parsed message type \"%s\" from Player %s", msgType, assignedPlayer.ID)

		gameInstanceMutex.Lock() // Lock game state for the duration of the action processing

		if gameInstance == nil || gameInstance.Players == nil || gameInstance.CurrentTurnPlayerIndex < 0 || gameInstance.CurrentTurnPlayerIndex >= len(gameInstance.Players) {
			log.Printf("Game not ready or invalid turn index for %s from Player %s", msgType, assignedPlayer.ID)
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Game not ready to process action."}`))
			log.Println("DEBUG: Explicit unlock before 'game not ready' continue in main loop")
			gameInstanceMutex.Unlock()
			continue
		}
		currentPlayerInGame := gameInstance.Players[gameInstance.CurrentTurnPlayerIndex]

		actionCtx := &ActionContext{
			Game:         gameInstance,
			Clients:      clients,    // Pass the global clients map
			ClientsMu:    &clientsMu, // Pass the mutex for it
			AssignedConn: conn,
		}

		var shouldContinueLoop, needsBroadcast bool

		switch msgType {
		case "chat":
			processChatAction(actionCtx, assignedPlayer, receivedMsg)
			needsBroadcast = false     // Chat doesn't trigger game state broadcast
			shouldContinueLoop = false // Chat doesn't make the main loop continue
			log.Println("DEBUG: chat - explicit unlock at end of case")
			gameInstanceMutex.Unlock()

		case "playCards":
			shouldContinueLoop, needsBroadcast = processPlayCardsAction(actionCtx, assignedPlayer, currentPlayerInGame, receivedMsg)
			if shouldContinueLoop {
				log.Println("DEBUG: playCards - explicit unlock because handler signaled continue")
				gameInstanceMutex.Unlock()
			} else {
				log.Println("DEBUG: playCards - explicit unlock at end of successful processing by handler")
				gameInstanceMutex.Unlock()
			}

		case "passTurn":
			shouldContinueLoop, needsBroadcast = processPassTurnAction(actionCtx, assignedPlayer, currentPlayerInGame, receivedMsg)
			if shouldContinueLoop {
				log.Println("DEBUG: passTurn - explicit unlock because handler signaled continue")
				gameInstanceMutex.Unlock()
			} else {
				log.Println("DEBUG: passTurn - explicit unlock at end of successful processing by handler")
				gameInstanceMutex.Unlock()
			}

		case "newGame":
			// No specific player context needed for newGame, but actionCtx provides gameInstance
			// assignedPlayer and currentPlayerInGame are not strictly used by processNewGameAction
			shouldContinueLoop, needsBroadcast = processNewGameAction(actionCtx)
			// processNewGameAction always returns shouldContinueLoop = false
			log.Println("DEBUG: newGame - explicit unlock at end of processing by handler")
			gameInstanceMutex.Unlock()

		case "setAlias":
			log.Printf("Received setAlias action from %s", assignedPlayer.ID)
			var aliasData struct { // Define struct for parsing alias
				Alias string `json:"alias"`
			}
			if err := json.Unmarshal(msgBytes, &aliasData); err != nil {
				log.Printf("Error unmarshalling setAlias payload: %v", err)
				// Optionally send an error back to the client
				shouldContinueLoop = false // Ensure loop continues
				gameInstanceMutex.Unlock()
				break
			}

			// Sanitize alias basic
			alias := strings.TrimSpace(aliasData.Alias)
			if len(alias) == 0 {
				alias = assignedPlayer.ID // Default to ID if empty after trim
			} else if len(alias) > 20 { // Max length example
				alias = alias[:20]
			}
			assignedPlayer.Name = alias
			log.Printf("Player %s set alias to %s", assignedPlayer.ID, assignedPlayer.Name)
			shouldContinueLoop = true
			needsBroadcast = true
			gameInstanceMutex.Unlock()

		default:
			log.Printf("Received unhandled message type \"%s\" from Player %s", msgType, assignedPlayer.ID)
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type": "error", "content": "Unknown message type: %s"}`, msgType)))
			needsBroadcast = false
			shouldContinueLoop = false
			log.Println("DEBUG: default case - explicit unlock")
			gameInstanceMutex.Unlock()
		}

		// Post-action processing based on handler results
		if needsBroadcast {
			// The lock for gameInstance was released by the case block before this point.
			// broadcastGameState needs to acquire it if it reads gameInstance directly.
			// Our current broadcastGameState does NOT lock gameInstanceMutex; it expects caller to.
			// So, we MUST re-acquire the lock here for the broadcast if gameInstance is read inside broadcastGameState.
			// This is crucial for data consistency during broadcast.
			gameInstanceMutex.Lock()
			broadcastGameState(gameInstance)
			gameInstanceMutex.Unlock()
		}

		if shouldContinueLoop {
			continue
		}
		// If not continuing, the loop naturally iterates to read the next message.
	}
}

func broadcastMessage(messageType int, message []byte, sender *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for conn := range clients {
		// if conn == sender { continue } // Uncomment to avoid sending echo to original sender for some message types
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("Broadcast write error:", err)
		}
	}
}

func main() {
	fmt.Println("Starting Big Two game server...")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleWebSocket)

	go func() {
		log.Println("Web server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	fmt.Println("Initializing game instance for web...")
	deck := NewDeck()
	deck.Shuffle()

	// === Single Player Debug Mode: Initialize only one player ===
	const singlePlayerDebug = false // Set to false for multiplayer
	var players []*Player
	if singlePlayerDebug {
		log.Println("INFO: Initializing in SINGLE PLAYER debug mode.")
		players = []*Player{
			NewPlayer(1, "Player1"), NewPlayer(2, "P2"),
		}
	} else {
		players = []*Player{
			NewPlayer(1, "P1"), NewPlayer(2, "P2"), NewPlayer(3, "P3"), NewPlayer(4, "P4"),
		}
	}
	// ==========================================================

	cardsPerPlayer := 0
	if len(players) > 0 {
		cardsPerPlayer = len(deck) / len(players) // Deal all cards if single player, or 13 for 4 players
	} else {
		log.Fatalln("No players initialized!")
	}

	if len(deck) < cardsPerPlayer*len(players) && !singlePlayerDebug { // Check only needed for multiplayer with fixed deal
		log.Fatalf("Not enough cards for %d players with %d cards each.", len(players), cardsPerPlayer)
	}

	for i, player := range players {
		hand, dealt := deck.Deal(cardsPerPlayer)
		if !dealt {
			log.Fatalf("Failed to deal %d cards to player %s", cardsPerPlayer, player.ID)
		}
		player.Hand = hand
		player.Hand.Sort()
		player.OrderInTurn = i // Assign turn order index explicitly
	}

	var startingPlayerIndex int
	if singlePlayerDebug && len(players) == 1 {
		startingPlayerIndex = 0
		log.Println("SINGLE PLAYER MODE: Player1 starts.")
		// Optional: Check if player 0 has 3D for log consistency, though they get all cards.
		found3D := false
		cardToFind := Card{Rank: Rank3, Suit: Diamonds}
		for _, card := range players[0].Hand {
			if card == cardToFind {
				found3D = true
				break
			}
		}
		if found3D {
			log.Println("SINGLE PLAYER MODE: Player1 has the 3 of Diamonds.")
		} else {
			log.Println("SINGLE PLAYER MODE: Player1 does not have 3 of Diamonds (should have with full deck), but starts anyway.")
		}
	} else if len(players) > 0 { // Multiplayer logic
		startingPlayerIndex = -1
		cardToFind := Card{Rank: Rank3, Suit: Diamonds}
		for i, player := range players {
			for _, card := range player.Hand {
				if card == cardToFind {
					startingPlayerIndex = i
					break
				}
			}
			if startingPlayerIndex != -1 {
				break
			}
		}
		if startingPlayerIndex == -1 {
			log.Println("CRITICAL: 3 of Diamonds not found in any hand. Defaulting to Player 0 to start.")
			startingPlayerIndex = 0
		}
	} else {
		log.Fatalln("No players to determine starting index.")
	}

	initialGameState := &GameState{
		Players:                players,
		CurrentTurnPlayerIndex: startingPlayerIndex,
		LastPlayedHand:         nil,
		RuleEngine:             NewBigTwoRuleEngine(),
		PassCount:              0,
		IsGameOver:             false, // Round is not over initially
		WinnerID:               "",
		Scores:                 make(map[string]int), // Overall scores, init empty or to 0 for players
		RoundNumber:            1,
		TargetScore:            100, // Default target score (penalty limit)
		IsMatchOver:            false,
		OverallWinnerID:        "",
	}
	gameInstance = initialGameState

	// Initialize scores to 0 for all players
	for _, p := range gameInstance.Players {
		gameInstance.Scores[p.ID] = 0
	}

	log.Println("Game state initialized. Server running. Waiting for WebSocket connections...")
	if startingPlayerIndex != -1 && len(gameInstance.Players) > 0 {
		log.Printf("Player %s (%s) should start. CurrentTurnPlayerIndex: %d",
			gameInstance.Players[startingPlayerIndex].Name,
			gameInstance.Players[startingPlayerIndex].ID,
			gameInstance.CurrentTurnPlayerIndex)
	} else {
		log.Println("Game starting condition not fully met (e.g. no players or no 3D found and no default). Check logs.")
	}

	select {}
}

// resetRoundState re-initializes the provided game state for a new ROUND.
// It uses the existing player objects but deals new hands and resets round-specific game variables.
// Overall scores, RoundNumber, TargetScore, and IsMatchOver are NOT reset here.
func resetRoundState(game *GameState) {
	log.Println("Resetting state for next round...")
	if game == nil || game.Players == nil {
		log.Println("ERROR: Cannot reset round state for nil game instance or game with nil players.")
		return
	}

	newDeck := NewDeck()
	newDeck.Shuffle()

	cardsPerPlayer := 0
	if len(game.Players) > 0 {
		// Determine cards per player (e.g. 13 for 4p, 52/n for other counts)
		if len(game.Players) == 4 {
			cardsPerPlayer = 13
		} else if len(game.Players) > 0 { // Ensure no division by zero if player count is manipulated
			cardsPerPlayer = len(newDeck) / len(game.Players)
		} else {
			log.Println("ERROR: No players to deal cards to during round reset.")
			return
		}
	} else {
		log.Println("ERROR: No players in game instance to deal cards to during round reset.")
		return
	}

	// Reset player-specific states and deal new hands
	for _, player := range game.Players {
		hand, dealt := newDeck.Deal(cardsPerPlayer)
		if !dealt {
			log.Printf("ERROR: Failed to deal %d cards to player %s during round reset.", cardsPerPlayer, player.ID)
			player.Hand = Deck{}
		} else {
			player.Hand = hand
			player.Hand.Sort()
		}
		player.HasPassed = false
	}

	// Determine starting player (e.g., with 3 of Diamonds)
	startingPlayerIndex := 0 // Default
	// Logic for finding 3 of Diamonds (same as initial setup)
	// This part of your existing resetGameInstance was good.
	if len(game.Players) == 1 { // Special case for single player debug/testing
		startingPlayerIndex = 0
	} else if len(game.Players) > 1 {
		found3D := false
		cardToFind := Card{Rank: Rank3, Suit: Diamonds}
		for i, p := range game.Players {
			for _, card := range p.Hand {
				if card == cardToFind {
					startingPlayerIndex = i
					found3D = true
					break
				}
			}
			if found3D {
				break
			}
		}
		if !found3D {
			log.Println("WARNING: 3 of Diamonds not found in any hand after reset for new round. Defaulting to Player 0 to start.")
			startingPlayerIndex = 0 // Fallback if 3D somehow isn't dealt
		}
	}

	// Reset round-specific game variables
	game.CurrentTurnPlayerIndex = startingPlayerIndex
	game.LastPlayedHand = nil
	game.PassCount = 0
	game.IsGameOver = false // Round is starting
	game.WinnerID = ""      // No round winner yet
	// game.Scores are overall scores and are NOT reset here
	// game.RoundNumber is incremented by caller (processNewGameAction)
	// game.TargetScore, game.IsMatchOver, game.OverallWinnerID are NOT reset here

	log.Printf("Round state reset. Player %s (%s) to start. CurrentTurnPlayerIndex: %d",
		game.Players[startingPlayerIndex].Name,
		game.Players[startingPlayerIndex].ID,
		game.CurrentTurnPlayerIndex)
}

// resetMatchState resets the game to a brand new match state.
// This includes resetting overall scores, round number, etc.
func resetMatchState(game *GameState) {
	log.Println("Resetting full match state...")
	game.RoundNumber = 1
	game.IsGameOver = false
	game.IsMatchOver = false
	game.WinnerID = ""
	game.OverallWinnerID = ""
	game.LastPlayedHand = nil
	game.PassCount = 0
	game.Scores = make(map[string]int)
	game.RoundScoresHistory = make([]map[string]int, 0) // Clear history for a new match

	// Initialize scores for all players to 0 for the new match
	for _, p := range game.Players {
		game.Scores[p.ID] = 0
	}

	// Now reset for the first round of the new match
	resetRoundState(game)
	log.Println("New match ready.")
}
