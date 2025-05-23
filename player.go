package main

import "fmt"

// Player represents a player in the game.
type Player struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Hand        Deck   `json:"hand"` // Deck type is []Card
	IsConnected bool
	Score       int
	OrderInTurn int  // To determine play sequence
	HasPassed   bool `json:"hasPassed"` // Tracks if player passed in the current round of plays
}

// NewPlayer creates and returns a new player.
func NewPlayer(id int, name string) *Player {
	return &Player{
		ID:          fmt.Sprintf("player%d", id),
		Name:        name,
		Hand:        Deck{},
		IsConnected: true, // Assume connected on creation
		Score:       0,
		OrderInTurn: -1, // Will be set during game setup
		HasPassed:   false,
	}
}

// Contains checks if a card is present in a deck.
func (d Deck) Contains(card Card) bool {
	for _, c := range d {
		if c == card {
			return true
		}
	}
	return false
}

// RemoveCards removes a sub-deck of cards from a player's hand.
// Returns true if all cards were successfully found and removed, false otherwise.
// Modifies p.Hand directly if successful.
func (p *Player) RemoveCards(cardsToRemove Deck) bool {
	if len(cardsToRemove) == 0 {
		return true // Nothing to remove
	}
	newHand := make(Deck, 0, len(p.Hand))

	// Create a map of cards to remove for efficient lookup and tracking counts
	toRemoveMap := make(map[Card]int)
	for _, card := range cardsToRemove {
		toRemoveMap[card]++
	}

	originalHandSize := len(p.Hand)
	cardsSuccessfullyProcessedForRemoval := 0

	for _, cardInHand := range p.Hand {
		if count, found := toRemoveMap[cardInHand]; found && count > 0 {
			toRemoveMap[cardInHand]-- // Decrement count for this card
			cardsSuccessfullyProcessedForRemoval++
			// Do not add this card to newHand (it's being removed)
		} else {
			newHand = append(newHand, cardInHand) // Keep this card
		}
	}

	if cardsSuccessfullyProcessedForRemoval == len(cardsToRemove) && len(newHand) == originalHandSize-len(cardsToRemove) {
		p.Hand = newHand
		p.Hand.Sort() // Keep hand sorted
		return true
	}

	// If not all cards were found or counts didn't match, the removal is considered failed.
	// The player's hand is not modified in this case to maintain consistency.
	return false
}
