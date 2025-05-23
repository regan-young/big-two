package main

// HandType represents the type of a 5-card poker hand or other valid Big 2 play.
type HandType int

const (
	InvalidHand HandType = iota
	Single
	Pair
	Triple
	Straight           // 5 cards
	Flush              // 5 cards
	FullHouse          // 5 cards
	FourOfAKindPlusOne // 5 cards (Bomb)
	StraightFlush      // 5 cards
)

// String returns a string representation of the hand type.
func (ht HandType) String() string {
	return []string{
		"Invalid", "Single", "Pair", "Triple", "Straight", "Flush",
		"Full House", "Four of a Kind", "Straight Flush",
	}[ht]
}

// PlayedHand represents the cards played in a turn and their type.
type PlayedHand struct {
	Cards          []Card   `json:"cards"`                    // Already had []Card which is good.
	PlayerID       string   `json:"playerId"`                 // Was playerID, making consistent camelCase
	HandType       HandType `json:"handType"`                 // Good.
	HandTypeString string   `json:"handTypeString,omitempty"` // Good.
	Rank           Rank     `json:"rank"`                     // Good.
	EffectiveRank  Rank     // For straights/flushes, the rank of the highest card. For pairs/triples/quads, the rank of the set. For full house, rank of the triple.
	EffectiveSuit  Suit     // For tie-breaking pairs or highest card in flushes/straights.
}

// GameState represents the overall state of the Big Two game.
type GameState struct {
	Players                []*Player         `json:"players"`
	CurrentTurnPlayerIndex int               `json:"currentPlayerIndex"`
	LastPlayedHand         *PlayedHand       `json:"lastPlayedHand"` // Pointer to allow nil
	RuleEngine             *BigTwoRuleEngine `json:"-"`              // Not serialized directly
	PassCount              int               `json:"passCount"`
	IsGameOver             bool              `json:"isGameOver"`         // True if the current ROUND is over
	WinnerID               string            `json:"winnerId,omitempty"` // Winner of the current ROUND
	Scores                 map[string]int    `json:"scores,omitempty"`   // Overall accumulated scores for the MATCH

	// New fields for multi-round/match play
	RoundNumber        int              `json:"roundNumber"`
	TargetScore        int              `json:"targetScore"` // Max penalty points before match ends
	IsMatchOver        bool             `json:"isMatchOver"`
	OverallWinnerID    string           `json:"overallWinnerId,omitempty"`
	RoundScoresHistory []map[string]int `json:"roundScoresHistory,omitempty"` // History of scores for each round
}

// --- Game Initialization & Helper Functions ---

// FindPlayerWith3D finds the player with the 3 of Diamonds to start the game.
// Returns the index of the player.
func FindPlayerWith3D(players []*Player) int {
	for i, p := range players {
		for _, card := range p.Hand {
			if card.Rank == Rank3 && card.Suit == Diamonds {
				return i
			}
		}
	}
	return 0 // Fallback, though in a real game 3D must exist.
}

// Helper function to check if a deck contains a specific card.
func containsCard(deck Deck, target Card) bool {
	for _, c := range deck {
		if c.Rank == target.Rank && c.Suit == target.Suit {
			return true
		}
	}
	return false
}

// NewGameState initializes a new game state.
// ... existing code ...
