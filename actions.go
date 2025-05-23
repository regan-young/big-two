package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// MessageHandler interface for different message types
// Deprecated: Prefer direct functions for simplicity in this context unless polymorphism is truly needed.
// type MessageHandler interface {
// 	Handle(assignedPlayer *Player, currentPlayerInGame *Player, receivedMsg map[string]interface{}, conn *websocket.Conn, game *GameState, clients map[*websocket.Conn]*client, clientsMu *sync.Mutex) (shouldContinue bool, broadcastStateNeeded bool)
// }

// ActionContext holds dependencies for action handlers
// This helps in reducing the number of arguments passed to handler functions.
type ActionContext struct {
	Game         *GameState
	Clients      map[*websocket.Conn]*client // For direct client interactions if needed beyond broadcast
	ClientsMu    *sync.Mutex                 // To protect Clients map if directly accessed
	AssignedConn *websocket.Conn             // The connection of the player making the action
}

// processPlayCardsAction handles the logic for a "playCards" message.
// Assumes gameInstanceMutex is held by the caller (handleWebSocket).
func processPlayCardsAction(ctx *ActionContext, assignedPlayer *Player, currentPlayerInGame *Player, receivedMsg map[string]interface{}) (shouldContinue bool, broadcastStateNeeded bool) {
	if ctx.Game.IsGameOver {
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Game is over."}`))
		return true, false // continue listening for messages, no broadcast needed
	}

	if assignedPlayer != currentPlayerInGame {
		errMsg := fmt.Sprintf(`{"type": "error", "content": "It's not your turn. Currently Player %s's turn."}`,
			currentPlayerInGame.Name)
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		return true, false // continue, no broadcast
	}

	playedCardsData, dataOk := receivedMsg["cards"]
	if !dataOk {
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Play message missing card data."}`))
		return true, false
	}

	parsedDeck, parseErr := parseCardsFromClientData(playedCardsData) // parseCardsFromClientData remains a global helper in main.go
	if parseErr != nil {
		errMsg := fmt.Sprintf(`{"type": "error", "content": "Invalid card data: %s"}`, parseErr.Error())
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		return true, false
	}

	canPlayCards := true
	tempHandCheck := make(map[Card]int)
	for _, c := range assignedPlayer.Hand {
		tempHandCheck[c]++
	}
	for _, c := range parsedDeck {
		if tempHandCheck[c] > 0 {
			tempHandCheck[c]--
		} else {
			canPlayCards = false
			break
		}
	}
	if !canPlayCards {
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Invalid play: You do not possess all the cards you are trying to play."}`))
		return true, false
	}

	determinedHand, errDet := ctx.Game.RuleEngine.DeterminePlayedHand(parsedDeck)
	if errDet != nil {
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type": "error", "content": "Invalid hand: %s"}`, errDet.Error())))
		return true, false
	}
	determinedHand.PlayerID = assignedPlayer.ID
	determinedHand.HandTypeString = determinedHand.HandType.String()

	if !ctx.Game.RuleEngine.BeatsLastHand(determinedHand, ctx.Game.LastPlayedHand) {
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Your hand does not beat the hand on the table."}`))
		return true, false
	}

	if !assignedPlayer.RemoveCards(parsedDeck) {
		log.Printf("CRITICAL: Failed to remove cards %s from player %s hand %s after validation.", parsedDeck.String(), assignedPlayer.ID, assignedPlayer.Hand.String())
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Server error: could not remove cards from hand. Play aborted."}`))
		return true, false
	}

	ctx.Game.LastPlayedHand = determinedHand
	ctx.Game.PassCount = 0
	assignedPlayer.HasPassed = false
	log.Printf("Player %s (%s) played: %s. Cards remaining: %d", assignedPlayer.ID, assignedPlayer.Name, determinedHand.Cards, len(assignedPlayer.Hand))

	if len(assignedPlayer.Hand) == 0 {
		// Player has won the round
		ctx.Game.IsGameOver = true
		ctx.Game.WinnerID = assignedPlayer.ID
		log.Printf("Player %s (%s) has won Round %d!", assignedPlayer.Name, assignedPlayer.ID, ctx.Game.RoundNumber)

		// Calculate scores for the round
		roundScores := CalculateScores(ctx.Game)
		log.Printf("Round %d scores calculated: %v", ctx.Game.RoundNumber, roundScores)

		// Append round scores to history
		if ctx.Game.RoundScoresHistory == nil {
			ctx.Game.RoundScoresHistory = make([]map[string]int, 0)
		}
		ctx.Game.RoundScoresHistory = append(ctx.Game.RoundScoresHistory, roundScores)

		// Update overall scores and check for match end
		matchShouldEnd := false
		for _, p := range ctx.Game.Players {
			if roundScore, ok := roundScores[p.ID]; ok {
				ctx.Game.Scores[p.ID] += roundScore // Add round score to overall score
			}
			// Check if any player (not just the winner of the round) has reached/exceeded target score
			if ctx.Game.Scores[p.ID] >= ctx.Game.TargetScore {
				matchShouldEnd = true
			}
		}

		if matchShouldEnd {
			ctx.Game.IsMatchOver = true
			log.Printf("Match ends after Round %d! A player reached or exceeded target score of %d.", ctx.Game.RoundNumber, ctx.Game.TargetScore)
			// Determine overall winner (lowest score)
			lowestScore := -1
			var overallWinner *Player = nil
			for _, p := range ctx.Game.Players {
				if overallWinner == nil || ctx.Game.Scores[p.ID] < lowestScore {
					lowestScore = ctx.Game.Scores[p.ID]
					overallWinner = p
				}
			}
			if overallWinner != nil {
				ctx.Game.OverallWinnerID = overallWinner.ID
				log.Printf("Overall Winner of the Match: %s (%s) with %d points!", overallWinner.Name, overallWinner.ID, lowestScore)
			} else {
				log.Println("ERROR: Could not determine overall winner despite match ending.")
			}
		} else {
			log.Printf("Round %d ended. Match continues. Current overall scores: %v", ctx.Game.RoundNumber, ctx.Game.Scores)
		}
		// No turn advancement here, the round/match is over.
	} else {
		// Round is not over, advance turn
		ctx.Game.CurrentTurnPlayerIndex = (ctx.Game.CurrentTurnPlayerIndex + 1) % len(ctx.Game.Players)
		// The next player might have passed in a previous round of betting on the same trick,
		// but for a new trick or continued play, their HasPassed should be reset when it becomes their turn to act.
		// However, HasPassed is more about the *current trick sequence*.
		// For simplicity, if we just advance turn, their HasPassed status from a previous trick might persist.
		// It's generally reset when a trick is won (PassCount clears, all HasPassed reset), or for the new turn player.
		// Let's ensure the next player to play is marked as not having passed FOR THIS TURN.
		if ctx.Game.CurrentTurnPlayerIndex < len(ctx.Game.Players) && ctx.Game.CurrentTurnPlayerIndex >= 0 {
			ctx.Game.Players[ctx.Game.CurrentTurnPlayerIndex].HasPassed = false
		} else {
			log.Printf("ERROR: CurrentTurnPlayerIndex out of bounds: %d", ctx.Game.CurrentTurnPlayerIndex)
		}
		log.Printf("Turn advances to Player %s (%s)", ctx.Game.Players[ctx.Game.CurrentTurnPlayerIndex].Name, ctx.Game.Players[ctx.Game.CurrentTurnPlayerIndex].ID)
	}

	// Debug log for game state after play
	/* gameInstanceJSON, err := json.MarshalIndent(ctx.Game, "", "  ")
	if err != nil {
		log.Printf("DEBUG: Error marshalling gameInstance for logging in processPlayCardsAction: %v", err)
	} else {
		log.Printf("DEBUG: gameInstance in processPlayCardsAction (JSON):\n%s", string(gameInstanceJSON))
	} */

	// The original unconditional block that set IsGameOver etc. is now inside the if len(assignedPlayer.Hand) == 0 block.

	return false, true // Broadcast state after any valid play (either turn advance or game/match end)
}

// processPassTurnAction handles the logic for a "passTurn" message.
// Assumes gameInstanceMutex is held by the caller.
func processPassTurnAction(ctx *ActionContext, assignedPlayer *Player, currentPlayerInGame *Player, _ map[string]interface{}) (shouldContinue bool, broadcastStateNeeded bool) {
	if ctx.Game.IsGameOver {
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "Game is over."}`))
		return true, false // continue listening, no broadcast
	}

	if assignedPlayer != currentPlayerInGame {
		errMsg := fmt.Sprintf(`{"type": "error", "content": "It's not your turn to pass. Currently Player %s's turn."}`,
			currentPlayerInGame.Name)
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(errMsg))
		return true, false
	}
	if ctx.Game.LastPlayedHand == nil && ctx.Game.PassCount == 0 {
		ctx.AssignedConn.WriteMessage(websocket.TextMessage, []byte(`{"type": "error", "content": "You cannot pass when you are leading a new trick."}`))
		return true, false
	}

	assignedPlayer.HasPassed = true
	ctx.Game.PassCount++
	log.Printf("Player %s (%s) passed. PassCount: %d", assignedPlayer.ID, assignedPlayer.Name, ctx.Game.PassCount)

	if ctx.Game.PassCount >= (len(ctx.Game.Players) - 1) {
		log.Printf("All other players passed. Player %s wins the trick and starts new.", ctx.Game.Players[ctx.Game.CurrentTurnPlayerIndex].Name)
		ctx.Game.LastPlayedHand = nil
		ctx.Game.PassCount = 0
		for _, p := range ctx.Game.Players {
			p.HasPassed = false
		}
	}
	ctx.Game.CurrentTurnPlayerIndex = (ctx.Game.CurrentTurnPlayerIndex + 1) % len(ctx.Game.Players)
	ctx.Game.Players[ctx.Game.CurrentTurnPlayerIndex].HasPassed = false
	return false, true // Do not continue loop, broadcast needed
}

// processChatAction handles the logic for a "chat" message.
// This function does not modify game state directly protected by gameInstanceMutex,
// but it does use ctx.Clients and ctx.ClientsMu for broadcasting.
// The main gameInstanceMutex is assumed to be held by the caller for consistency of the overall request lifecycle.
func processChatAction(ctx *ActionContext, assignedPlayer *Player, receivedMsg map[string]interface{}) {
	content, contentOk := receivedMsg["content"].(string)
	if contentOk {
		broadcastMsgPayload := map[string]string{
			"type": "chat", "sender": fmt.Sprintf("%s (%s)", assignedPlayer.Name, assignedPlayer.ID), "content": content,
		}
		jsonBroadcast, _ := json.Marshal(broadcastMsgPayload)
		// broadcastMessage function needs to be updated to accept clients map and mutex, or use a method on ActionContext
		// For now, assuming a global broadcastMessage or one that can be called this way:
		broadcastMessage(websocket.TextMessage, jsonBroadcast, ctx.AssignedConn) // broadcastMessage will use global clients and clientsMu
	}
}

// processNewGameAction handles the logic for a "newGame" message.
// Assumes gameInstanceMutex is held by the caller.
func processNewGameAction(ctx *ActionContext) (shouldContinue bool, broadcastStateNeeded bool) {
	if ctx.Game.IsMatchOver {
		log.Printf("Processing 'newGame' action: Starting a New Match because current match is over.")
		resetMatchState(ctx.Game) // Resets everything for a new match (defined in main.go)
	} else if ctx.Game.IsGameOver { // Current round is over, but match continues
		log.Printf("Processing 'newGame' action: Starting Next Round (Round %d).", ctx.Game.RoundNumber+1)
		ctx.Game.RoundNumber++
		resetRoundState(ctx.Game) // Resets only for the next round (defined in main.go)
	} else {
		// "New Game" clicked during an active round (neither round nor match is over).
		// Typically, this means the user wants to abandon the current match and start a fresh one.
		log.Printf("Processing 'newGame' action: Starting a New Match (abandoning current active round/match).")
		resetMatchState(ctx.Game)
	}

	// Comment about resetGameInstance is no longer relevant here.
	log.Println("Game state has been reset for new game/round. Broadcasting new game state.")
	return false, true // Always broadcast after a new game/round action
}
