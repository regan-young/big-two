package main

import (
	"fmt"
	// "sort" // Not directly used at top level of this file, but methods might use it via Deck.Sort()
)

// BigTwoRuleEngine encapsulates the core game logic for Big Two.
type BigTwoRuleEngine struct {
	// No fields needed for now, as rules are stateless in this implementation.
}

// NewBigTwoRuleEngine creates a new instance of the rule engine.
func NewBigTwoRuleEngine() *BigTwoRuleEngine {
	return &BigTwoRuleEngine{}
}

// DeterminePlayedHand analyzes a set of cards and determines if they form a valid Big 2 hand.
// It returns the PlayedHand struct (with type, effective rank/suit) or an error if invalid.
func (re *BigTwoRuleEngine) DeterminePlayedHand(selectedCards Deck) (*PlayedHand, error) {
	numCards := len(selectedCards)
	if numCards == 0 {
		return nil, fmt.Errorf("no cards selected")
	}

	var handType HandType = InvalidHand
	var effectiveRank Rank = -1
	var effectiveSuit Suit = -1

	switch numCards {
	case 1:
		handType = Single
		effectiveRank = selectedCards[0].Rank
		effectiveSuit = selectedCards[0].Suit
	case 2:
		if selectedCards[0].Rank == selectedCards[1].Rank {
			handType = Pair
			effectiveRank = selectedCards[0].Rank
			// Assuming cards are sorted by rank then suit, selectedCards[1] has the defining suit for the pair.
			effectiveSuit = selectedCards[1].Suit
		} else {
			return nil, fmt.Errorf("not a valid pair (ranks differ)")
		}
	case 3:
		if selectedCards[0].Rank == selectedCards[1].Rank && selectedCards[1].Rank == selectedCards[2].Rank {
			handType = Triple
			effectiveRank = selectedCards[0].Rank
			effectiveSuit = -1 // Suit doesn't matter for comparing triples
		} else {
			return nil, fmt.Errorf("not a valid triple")
		}
	case 5:
		// Cards are assumed sorted as they come from a sorted player hand.
		// Check in order of strength for 5-card hands.
		if isSF, sfRank, sfSuit := re.isStraightFlush(selectedCards); isSF {
			handType = StraightFlush
			effectiveRank = sfRank
			effectiveSuit = sfSuit
			break
		}
		if isFOAK, foakRank := re.isFourOfAKindPlusOne(selectedCards); isFOAK {
			handType = FourOfAKindPlusOne
			effectiveRank = foakRank
			effectiveSuit = -1 // Suit doesn't matter for FOAK
			break
		}
		if isFH, fhRank := re.isFullHouse(selectedCards); isFH {
			handType = FullHouse
			effectiveRank = fhRank // Rank of the triple
			effectiveSuit = -1     // Suit doesn't matter for Full House
			break
		}
		if isF, fSuit, fRank := re.isFlush(selectedCards); isF {
			handType = Flush
			effectiveRank = fRank // Rank of the highest card
			effectiveSuit = fSuit // Suit of the flush
			break
		}
		if isS, sRank, sSuit := re.isStraight(selectedCards); isS {
			handType = Straight
			effectiveRank = sRank // Rank of the highest card
			effectiveSuit = sSuit // Suit of the highest card
			break
		}
		return nil, fmt.Errorf("selected 5 cards do not form a valid Big 2 hand (Straight, Flush, Full House, Bomb, or Straight Flush)")
	default:
		return nil, fmt.Errorf("invalid number of cards played: %d. Must be 1, 2, 3, or 5", numCards)
	}

	if handType == InvalidHand {
		return nil, fmt.Errorf("selected cards do not form a valid Big 2 hand type")
	}

	playedCardsCopy := make(Deck, len(selectedCards))
	copy(playedCardsCopy, selectedCards)
	playedCardsCopy.Sort() // Ensure cards in the returned PlayedHand are always sorted.

	return &PlayedHand{
		Cards:         playedCardsCopy,
		HandType:      handType,
		EffectiveRank: effectiveRank,
		EffectiveSuit: effectiveSuit,
	}, nil
}

// BeatsLastHand checks if the currentPlay can beat the lastPlayedHand according to Big 2 rules.
func (re *BigTwoRuleEngine) BeatsLastHand(currentPlay *PlayedHand, lastPlayedHand *PlayedHand) bool {
	if lastPlayedHand == nil {
		return currentPlay.HandType != InvalidHand
	}

	currentPlayerIsBomb := currentPlay.HandType == FourOfAKindPlusOne || currentPlay.HandType == StraightFlush
	lastPlayerIsBomb := lastPlayedHand.HandType == FourOfAKindPlusOne || lastPlayedHand.HandType == StraightFlush

	if currentPlayerIsBomb {
		if !lastPlayerIsBomb {
			return true // Bomb beats any non-bomb
		}
		// Both are bombs
		if currentPlay.HandType == StraightFlush && lastPlayedHand.HandType == FourOfAKindPlusOne {
			return true // SF beats FOAK
		}
		if currentPlay.HandType == FourOfAKindPlusOne && lastPlayedHand.HandType == StraightFlush {
			return false // FOAK doesn't beat SF
		}
		// Same type of bomb, compare by rank, then suit if SF
		if currentPlay.EffectiveRank > lastPlayedHand.EffectiveRank {
			return true
		}
		if currentPlay.EffectiveRank == lastPlayedHand.EffectiveRank {
			if currentPlay.HandType == StraightFlush { // SF ties broken by suit
				return currentPlay.EffectiveSuit > lastPlayedHand.EffectiveSuit
			}
			return false // FOAKs of same rank, or SFs of same rank & suit: cannot beat
		}
		return false // Current bomb rank is lower
	}

	if lastPlayerIsBomb { // Current play is not a bomb, but last was
		return false // Non-bomb cannot beat a bomb
	}

	// Standard hand comparison (neither is a bomb)
	if len(currentPlay.Cards) != len(lastPlayedHand.Cards) {
		return false // Must be same number of cards
	}

	// If hand types are different (and it's 5-card hands, and neither are bombs - handled above)
	if currentPlay.HandType != lastPlayedHand.HandType {
		// Only allow different hand types if they are both 5-card hands (non-bomb type)
		if len(currentPlay.Cards) == 5 { // Both must be 5 cards due to len check above
			// Allow stronger 5-card hand type to beat weaker 5-card hand type
			// This relies on HandType enum values being in order of strength for 5-card hands.
			// Straight < Flush < FullHouse (already covered by bomb logic: < FourOfAKind < StraightFlush)
			// We only need to compare Straight, Flush, FullHouse here as bombs are handled.
			// And FourOfAKindPlusOne and StraightFlush are already handled by the bomb logic above.
			// So, this comparison effectively applies to Straight, Flush, and FullHouse.
			return currentPlay.HandType > lastPlayedHand.HandType
		} else {
			// If not 5-card hands, types must match (e.g. pair vs pair, single vs single)
			return false
		}
	}

	// Hand types are the same, compare by effective rank, then suit if applicable.
	if currentPlay.EffectiveRank > lastPlayedHand.EffectiveRank {
		return true
	}
	if currentPlay.EffectiveRank < lastPlayedHand.EffectiveRank {
		return false
	}

	// Ranks are equal, compare by suit where applicable
	switch currentPlay.HandType {
	case Single, Pair, Straight, Flush:
		return currentPlay.EffectiveSuit > lastPlayedHand.EffectiveSuit
	case Triple, FullHouse:
		return false // Ranks are equal, suit doesn't break ties
	default:
		return false // Should not be reached
	}
}

// --- Helper functions for 5-card hand validation (methods of BigTwoRuleEngine) ---

func (re *BigTwoRuleEngine) isStraight(cards Deck) (bool, Rank, Suit) {
	if len(cards) != 5 {
		return false, -1, -1
	}
	// Cards are assumed to be sorted by Rank then Suit.

	// Case 1: 10, J, Q, K, A (Ace-high straight)
	if cards[0].Rank == Rank10 && cards[1].Rank == Jack && cards[2].Rank == Queen && cards[3].Rank == King && cards[4].Rank == Ace {
		return true, cards[4].Rank, cards[4].Suit // Effective rank is Ace, suit of Ace
	}

	// Case 2: J, Q, K, A, 2 (Two-high straight for Big Two)
	// Sorted by game rank: J, Q, K, A, 2
	if cards[0].Rank == Jack && cards[1].Rank == Queen && cards[2].Rank == King && cards[3].Rank == Ace && cards[4].Rank == Two {
		return true, cards[4].Rank, cards[4].Suit // Effective rank is Two, suit of Two
	}

	// Case 3: A, 2, 3, 4, 5 (Five-high straight, Ace low)
	// Sorted by game rank: 3, 4, 5, A, 2
	if cards[0].Rank == Rank3 && cards[1].Rank == Rank4 && cards[2].Rank == Rank5 && cards[3].Rank == Ace && cards[4].Rank == Two {
		return true, cards[2].Rank, cards[2].Suit // Effective rank is Five, suit of the Five card
	}

	// Case 4: General consecutive ranks (e.g., 3-4-5-6-7 or 7-8-9-10-J)
	// This will not catch 10-A, J-A-2, or the A-2-3-4-5 (3,4,5,A,2) sequence due to rank values of Ace and Two.
	isConsecutive := true
	for i := 0; i < 4; i++ {
		if cards[i+1].Rank != cards[i].Rank+1 {
			isConsecutive = false
			break
		}
	}
	if isConsecutive {
		return true, cards[4].Rank, cards[4].Suit // Highest card of the sequence determines rank and suit
	}

	return false, -1, -1
}

func (re *BigTwoRuleEngine) isFlush(cards Deck) (bool, Suit, Rank) {
	if len(cards) != 5 {
		return false, -1, -1
	}
	firstSuit := cards[0].Suit
	for i := 1; i < 5; i++ {
		if cards[i].Suit != firstSuit {
			return false, -1, -1
		}
	}
	// For a flush, the effective rank is the rank of the highest card,
	// and the effective suit is the suit of that card (which is the suit of the flush).
	// We explicitly use the highest card (cards[4]) for both properties for clarity.
	return true, cards[4].Suit, cards[4].Rank
}

func (re *BigTwoRuleEngine) isFullHouse(cards Deck) (bool, Rank) {
	if len(cards) != 5 {
		return false, -1
	}
	// Cards sorted: XXX YY or XX YYY
	if cards[0].Rank == cards[1].Rank && cards[1].Rank == cards[2].Rank && cards[3].Rank == cards[4].Rank && cards[2].Rank != cards[3].Rank {
		return true, cards[0].Rank // Triple XXX YY
	}
	if cards[0].Rank == cards[1].Rank && cards[2].Rank == cards[3].Rank && cards[3].Rank == cards[4].Rank && cards[1].Rank != cards[2].Rank {
		return true, cards[2].Rank // Triple XX YYY
	}
	return false, -1
}

func (re *BigTwoRuleEngine) isFourOfAKindPlusOne(cards Deck) (bool, Rank) {
	if len(cards) != 5 {
		return false, -1
	}
	// Cards sorted: XXXX Y or X YYYY
	if cards[0].Rank == cards[1].Rank && cards[1].Rank == cards[2].Rank && cards[2].Rank == cards[3].Rank {
		return true, cards[0].Rank // Rank of the XXXX.
	}
	if cards[1].Rank == cards[2].Rank && cards[2].Rank == cards[3].Rank && cards[3].Rank == cards[4].Rank {
		return true, cards[1].Rank // Rank of the YYYY.
	}
	return false, -1
}

func (re *BigTwoRuleEngine) isStraightFlush(cards Deck) (bool, Rank, Suit) {
	if len(cards) != 5 {
		return false, -1, -1
	}
	isFlushBool, _, _ := re.isFlush(cards)
	if !isFlushBool {
		return false, -1, -1
	}
	isStraightBool, highestRankStraight, highestSuitStraight := re.isStraight(cards)
	if isStraightBool {
		return true, highestRankStraight, highestSuitStraight
	}
	return false, -1, -1
}
