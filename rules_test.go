package main

import (
	"testing"
)

// Helper function to create a card for tests. Simplifies test case setup.
func C(rank Rank, suit Suit) Card { return Card{Rank: rank, Suit: suit} }

func TestBigTwoRuleEngine_DeterminePlayedHand_SinglesPairsTriples(t *testing.T) {
	re := NewBigTwoRuleEngine()

	tests := []struct {
		name         string
		cards        Deck
		wantHandType HandType
		wantEffRank  Rank
		wantEffSuit  Suit
		wantErr      bool
	}{
		// Singles
		{"Valid Single (3D)", Deck{C(Rank3, Diamonds)}, Single, Rank3, Diamonds, false},
		{"Valid Single (2S)", Deck{C(Two, Spades)}, Single, Two, Spades, false},

		// Pairs
		{"Valid Pair (3D, 3S)", Deck{C(Rank3, Diamonds), C(Rank3, Spades)}, Pair, Rank3, Spades, false},
		{"Valid Pair (AS, AH)", Deck{C(Ace, Spades), C(Ace, Hearts)}, Pair, Ace, Spades, false},
		{"Invalid Pair (3D, 4D)", Deck{C(Rank3, Diamonds), C(Rank4, Diamonds)}, InvalidHand, -1, -1, true},

		// Triples
		{"Valid Triple (7D, 7C, 7H)", Deck{C(Rank7, Diamonds), C(Rank7, Clubs), C(Rank7, Hearts)}, Triple, Rank7, -1, false},
		{"Invalid Triple (7D, 7C, 8H)", Deck{C(Rank7, Diamonds), C(Rank7, Clubs), C(Rank8, Hearts)}, InvalidHand, -1, -1, true},

		// Invalid counts
		{"Invalid - No cards", Deck{}, InvalidHand, -1, -1, true},
		{"Invalid - Four cards (not a bomb)", Deck{C(Rank4, Diamonds), C(Rank4, Clubs), C(Rank4, Hearts), C(Rank4, Spades)}, InvalidHand, -1, -1, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure cards are sorted as DeterminePlayedHand might rely on it for some internal checks or consistency
			// although for singles, pairs, triples, direct comparison is fine.
			// For actual gameplay, player hands are sorted, and selected cards are taken from it.
			tc.cards.Sort() // Sorting here to mimic sorted selection

			playedHand, err := re.DeterminePlayedHand(tc.cards)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeterminePlayedHand() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if playedHand.HandType != tc.wantHandType {
					t.Errorf("DeterminePlayedHand() HandType = %v, want %v", playedHand.HandType, tc.wantHandType)
				}
				if playedHand.EffectiveRank != tc.wantEffRank {
					t.Errorf("DeterminePlayedHand() EffectiveRank = %v, want %v", playedHand.EffectiveRank, tc.wantEffRank)
				}
				if playedHand.EffectiveSuit != tc.wantEffSuit {
					t.Errorf("DeterminePlayedHand() EffectiveSuit = %v, want %v", playedHand.EffectiveSuit, tc.wantEffSuit)
				}
				// Check if cards in playedHand are a sorted copy of input
				if len(playedHand.Cards) != len(tc.cards) {
					t.Errorf("DeterminePlayedHand() len(playedHand.Cards) = %d, want %d", len(playedHand.Cards), len(tc.cards))
				}
				// Further check for content equality if needed, assuming sort order is canonical.
			}
		})
	}
}

func TestBigTwoRuleEngine_BeatsLastHand_BasicScenarios(t *testing.T) {
	re := NewBigTwoRuleEngine()

	// Test hands (already determined and valid)
	phSingle3D := &PlayedHand{Cards: Deck{C(Rank3, Diamonds)}, HandType: Single, EffectiveRank: Rank3, EffectiveSuit: Diamonds}
	phSingle4C := &PlayedHand{Cards: Deck{C(Rank4, Clubs)}, HandType: Single, EffectiveRank: Rank4, EffectiveSuit: Clubs}
	phSingleKH := &PlayedHand{Cards: Deck{C(King, Hearts)}, HandType: Single, EffectiveRank: King, EffectiveSuit: Hearts}
	phSingleKS := &PlayedHand{Cards: Deck{C(King, Spades)}, HandType: Single, EffectiveRank: King, EffectiveSuit: Spades}
	phSingleAS := &PlayedHand{Cards: Deck{C(Ace, Spades)}, HandType: Single, EffectiveRank: Ace, EffectiveSuit: Spades}
	phSingle2D := &PlayedHand{Cards: Deck{C(Two, Diamonds)}, HandType: Single, EffectiveRank: Two, EffectiveSuit: Diamonds}

	phPair3 := &PlayedHand{Cards: Deck{C(Rank3, Diamonds), C(Rank3, Spades)}, HandType: Pair, EffectiveRank: Rank3, EffectiveSuit: Spades}  // 3D, 3S
	phPair4 := &PlayedHand{Cards: Deck{C(Rank4, Clubs), C(Rank4, Hearts)}, HandType: Pair, EffectiveRank: Rank4, EffectiveSuit: Hearts}     // 4C, 4H
	phPairK := &PlayedHand{Cards: Deck{C(King, Diamonds), C(King, Clubs)}, HandType: Pair, EffectiveRank: King, EffectiveSuit: Clubs}       // KD, KC
	phPairA_LowSuit := &PlayedHand{Cards: Deck{C(Ace, Diamonds), C(Ace, Clubs)}, HandType: Pair, EffectiveRank: Ace, EffectiveSuit: Clubs}  // AD, AC
	phPairA_HighSuit := &PlayedHand{Cards: Deck{C(Ace, Hearts), C(Ace, Spades)}, HandType: Pair, EffectiveRank: Ace, EffectiveSuit: Spades} // AH, AS

	phTriple6 := &PlayedHand{Cards: Deck{C(Rank6, Diamonds), C(Rank6, Clubs), C(Rank6, Hearts)}, HandType: Triple, EffectiveRank: Rank6, EffectiveSuit: -1}
	phTripleQ := &PlayedHand{Cards: Deck{C(Queen, Spades), C(Queen, Hearts), C(Queen, Diamonds)}, HandType: Triple, EffectiveRank: Queen, EffectiveSuit: -1}

	// 5-card hands (Rank & Suit details are important for tie-breaking where applicable)
	// Straights
	straight37D := &PlayedHand{Cards: Deck{C(Rank3, Clubs), C(Rank4, Hearts), C(Rank5, Spades), C(Rank6, Diamonds), C(Rank7, Diamonds)}, HandType: Straight, EffectiveRank: Rank7, EffectiveSuit: Diamonds}
	straight48H := &PlayedHand{Cards: Deck{C(Rank4, Clubs), C(Rank5, Spades), C(Rank6, Diamonds), C(Rank7, Hearts), C(Rank8, Hearts)}, HandType: Straight, EffectiveRank: Rank8, EffectiveSuit: Hearts}
	straight10AS := &PlayedHand{Cards: Deck{C(Rank10, Clubs), C(Jack, Hearts), C(Queen, Diamonds), C(King, Spades), C(Ace, Spades)}, HandType: Straight, EffectiveRank: Ace, EffectiveSuit: Spades}
	straightJQKA2D := &PlayedHand{Cards: Deck{C(Jack, Clubs), C(Queen, Hearts), C(King, Spades), C(Ace, Diamonds), C(Two, Diamonds)}, HandType: Straight, EffectiveRank: Two, EffectiveSuit: Diamonds}

	// Flushes (Suit of flush, Rank of highest card)
	flushKDH := &PlayedHand{Cards: Deck{C(Rank3, Diamonds), C(Rank5, Diamonds), C(Rank8, Diamonds), C(Jack, Diamonds), C(King, Diamonds)}, HandType: Flush, EffectiveRank: King, EffectiveSuit: Diamonds}
	flushAD_SpadeH := &PlayedHand{Cards: Deck{C(Rank4, Spades), C(Rank6, Spades), C(Rank9, Spades), C(Queen, Spades), C(Ace, Spades)}, HandType: Flush, EffectiveRank: Ace, EffectiveSuit: Spades}

	// Full Houses (Rank of triple)
	fh3o2 := &PlayedHand{Cards: Deck{C(Rank3, Diamonds), C(Rank3, Clubs), C(Rank3, Hearts), C(Two, Spades), C(Two, Diamonds)}, HandType: FullHouse, EffectiveRank: Rank3, EffectiveSuit: -1}
	fhAoK := &PlayedHand{Cards: Deck{C(Ace, Spades), C(Ace, Hearts), C(King, Diamonds), C(King, Clubs), C(King, Spades)}, HandType: FullHouse, EffectiveRank: King, EffectiveSuit: -1}

	// Four of a Kind + One (Bombs - Rank of Quads)
	foak7s3 := &PlayedHand{Cards: Deck{C(Rank7, Diamonds), C(Rank7, Clubs), C(Rank7, Hearts), C(Rank7, Spades), C(Rank3, Diamonds)}, HandType: FourOfAKindPlusOne, EffectiveRank: Rank7, EffectiveSuit: -1}
	foakAsK := &PlayedHand{Cards: Deck{C(Ace, Diamonds), C(Ace, Clubs), C(King, Hearts), C(Ace, Spades), C(Ace, Hearts)}, HandType: FourOfAKindPlusOne, EffectiveRank: Ace, EffectiveSuit: -1}

	// Straight Flushes (Bombs - Rank & Suit of highest card in straight component)
	sf37D := &PlayedHand{Cards: Deck{C(Rank3, Diamonds), C(Rank4, Diamonds), C(Rank5, Diamonds), C(Rank6, Diamonds), C(Rank7, Diamonds)}, HandType: StraightFlush, EffectiveRank: Rank7, EffectiveSuit: Diamonds}
	sf10AS := &PlayedHand{Cards: Deck{C(Rank10, Spades), C(Jack, Spades), C(Queen, Spades), C(King, Spades), C(Ace, Spades)}, HandType: StraightFlush, EffectiveRank: Ace, EffectiveSuit: Spades}
	sfJ2H := &PlayedHand{Cards: Deck{C(Jack, Hearts), C(Queen, Hearts), C(King, Hearts), C(Ace, Hearts), C(Two, Hearts)}, HandType: StraightFlush, EffectiveRank: Two, EffectiveSuit: Hearts}

	tests := []struct {
		name           string
		currentPlay    *PlayedHand
		lastPlayedHand *PlayedHand
		wantBeats      bool
	}{
		// Playing to an empty table
		{"Play Single to empty table", phSingle3D, nil, true},
		{"Play Pair to empty table", phPair4, nil, true},

		// Single vs Single
		{"Single 4C beats 3D", phSingle4C, phSingle3D, true},
		{"Single 3D does not beat 4C", phSingle3D, phSingle4C, false},
		{"Single KH beats 4C (rank)", phSingleKH, phSingle4C, true},
		{"Single 4C does not beat KH (rank)", phSingle4C, phSingleKH, false},
		{"Single KH beats KD (suit)",
			&PlayedHand{Cards: Deck{C(King, Hearts)}, HandType: Single, EffectiveRank: King, EffectiveSuit: Hearts},
			&PlayedHand{Cards: Deck{C(King, Diamonds)}, HandType: Single, EffectiveRank: King, EffectiveSuit: Diamonds},
			true},
		{"Single KD does not beat KH (suit)",
			&PlayedHand{Cards: Deck{C(King, Diamonds)}, HandType: Single, EffectiveRank: King, EffectiveSuit: Diamonds},
			&PlayedHand{Cards: Deck{C(King, Hearts)}, HandType: Single, EffectiveRank: King, EffectiveSuit: Hearts},
			false},
		{"Single same card does not beat itself", phSingle4C, phSingle4C, false},
		{"Single KS beats KH (suit)", phSingleKS, phSingleKH, true},
		{"Single KH does not beat KS (suit)", phSingleKH, phSingleKS, false},
		{"Single AS beats KS (rank)", phSingleAS, phSingleKS, true},
		{"Single KS does not beat AS (rank)", phSingleKS, phSingleAS, false},

		// Pair vs Pair
		{"Pair4 beats Pair3", phPair4, phPair3, true},
		{"Pair3 does not beat Pair4", phPair3, phPair4, false},
		{"PairA_HighS beats PairA_LowS (suit)", phPairA_HighSuit, phPairA_LowSuit, true},
		{"PairA_LowS does not beat PairA_HighS (suit)", phPairA_LowSuit, phPairA_HighSuit, false},
		{"PairK same rank different high suit, higher wins",
			&PlayedHand{Cards: Deck{C(King, Spades), C(King, Hearts)}, HandType: Pair, EffectiveRank: King, EffectiveSuit: Spades}, // KH, KS
			phPairK, // KD, KC (effSuit Club)
			true},

		// Triple vs Triple
		{"TripleQ beats Triple6", phTripleQ, phTriple6, true},
		{"Triple6 does not beat TripleQ", phTriple6, phTripleQ, false},
		{"TripleQ vs same TripleQ", phTripleQ, phTripleQ, false},

		// Straight vs Straight
		{"Straight 4-8H beats 3-7D (rank)", straight48H, straight37D, true},
		{"Straight 3-7D does not beat 4-8H (rank)", straight37D, straight48H, false},
		{"Straight 10-AS beats 4-8H (rank, Ace high)", straight10AS, straight48H, true},
		{"Straight J-2D beats 10-AS (rank, Two high)", straightJQKA2D, straight10AS, true},
		{"Straight J-2D vs J-2D (same highest card rank and suit)", straightJQKA2D, straightJQKA2D, false},
		{"Straight 3-7D vs 3-7S (same rank, higher suit wins)",
			&PlayedHand{Cards: Deck{C(Rank3, Clubs), C(Rank4, Diamonds), C(Rank5, Hearts), C(Rank6, Spades), C(Rank7, Spades)}, HandType: Straight, EffectiveRank: Rank7, EffectiveSuit: Spades},
			straight37D, //effSuit Diamonds
			true},

		// Flush vs Flush
		{"Flush AD_SpadeH beats KD_DiamondH (rank of highest card)", flushAD_SpadeH, flushKDH, true},
		{"Flush KD_DiamondH does not beat AD_SpadeH (rank)", flushKDH, flushAD_SpadeH, false},
		{"Flush AS_SameSuit beats KS_SameSuit (rank, suit same)",
			&PlayedHand{Cards: Deck{C(Two, Spades), C(Rank4, Spades), C(Rank6, Spades), C(Rank8, Spades), C(Ace, Spades)}, HandType: Flush, EffectiveRank: Ace, EffectiveSuit: Spades},
			&PlayedHand{Cards: Deck{C(Rank3, Spades), C(Rank5, Spades), C(Rank7, Spades), C(Rank9, Spades), C(King, Spades)}, HandType: Flush, EffectiveRank: King, EffectiveSuit: Spades},
			true},
		{"Flush KS_Spades beats KS_Hearts (suit of flush, ranks same)",
			&PlayedHand{Cards: Deck{C(Two, Spades), C(Rank4, Spades), C(Rank6, Spades), C(Rank8, Spades), C(King, Spades)}, HandType: Flush, EffectiveRank: King, EffectiveSuit: Spades},
			&PlayedHand{Cards: Deck{C(Two, Hearts), C(Rank4, Hearts), C(Rank6, Hearts), C(Rank8, Hearts), C(King, Hearts)}, HandType: Flush, EffectiveRank: King, EffectiveSuit: Hearts},
			true},

		// FullHouse vs FullHouse
		{"FH AoK (K triple) beats FH 3o2 (3 triple)", fhAoK, fh3o2, true},
		{"FH 3o2 does not beat FH AoK", fh3o2, fhAoK, false},
		{"FH AoK vs FH AoK (same rank triple)", fhAoK, fhAoK, false},

		// 5-card Hand Strength Order (Non-Bombs)
		{"Actual: Full House (3o2) beats Flush (KDH)", fh3o2, flushKDH, true},
		{"Actual: Flush (KDH) beats Straight (3-7D)", flushKDH, straight37D, true},
		{"Actual: Straight (3-7D) does not beat Flush (KDH)", straight37D, flushKDH, false},
		{"Actual: Flush (KDH) does not beat Full House (3o2)", flushKDH, fh3o2, false},

		// Bomb Logic - FOAK
		{"FOAK 7s beats Single 2D", foak7s3, phSingle2D, true},
		{"FOAK 7s beats Pair As", foak7s3, phPairA_HighSuit, true},
		{"FOAK 7s beats Triple Qs", foak7s3, phTripleQ, true},
		{"FOAK 7s beats Straight J-2D", foak7s3, straightJQKA2D, true},
		{"FOAK 7s beats Flush AD_SpadeH", foak7s3, flushAD_SpadeH, true},
		{"FOAK 7s beats FullHouse AoK", foak7s3, fhAoK, true},
		{"FOAK AsK beats FOAK 7s3 (rank)", foakAsK, foak7s3, true},
		{"FOAK 7s3 does not beat FOAK AsK (rank)", foak7s3, foakAsK, false},
		{"FOAK AsK vs FOAK AsK (same)", foakAsK, foakAsK, false},

		// Bomb Logic - Straight Flush
		{"SF 3-7D beats Single 2D", sf37D, phSingle2D, true},
		{"SF 3-7D beats Pair As", sf37D, phPairA_HighSuit, true},
		{"SF 3-7D beats Triple Qs", sf37D, phTripleQ, true},
		{"SF 3-7D beats Straight J-2D", sf37D, straightJQKA2D, true},
		{"SF 3-7D beats Flush AD_SpadeH", sf37D, flushAD_SpadeH, true},
		{"SF 3-7D beats FullHouse AoK", sf37D, fhAoK, true},
		{"SF 3-7D beats FOAK AsK", sf37D, foakAsK, true},
		{"FOAK AsK does not beat SF 3-7D", foakAsK, sf37D, false},
		{"SF 10-AS beats SF 3-7D (rank)", sf10AS, sf37D, true},
		{"SF 3-7D does not beat SF 10-AS (rank)", sf37D, sf10AS, false},
		{"SF J-2H beats SF 10-AS (rank)", sfJ2H, sf10AS, true},
		{"SF J-2H vs SF J-2H (same)", sfJ2H, sfJ2H, false},

		// Invalid Comparisons (card count mismatch, non-bomb)
		{"Pair vs Single (invalid count)", phPair4, phSingleKH, false},
		{"Single vs Pair (invalid count)", phSingleKH, phPair4, false},
		{"Straight vs Triple (invalid count)", straight37D, phTripleQ, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if gotBeats := re.BeatsLastHand(tc.currentPlay, tc.lastPlayedHand); gotBeats != tc.wantBeats {
				t.Errorf("BeatsLastHand() for %s: got %v, want %v", tc.name, gotBeats, tc.wantBeats)
			}
		})
	}
}

func TestBigTwoRuleEngine_DeterminePlayedHand_FiveCardHands(t *testing.T) {
	re := NewBigTwoRuleEngine()

	tests := []struct {
		name         string
		cards        Deck
		wantHandType HandType
		wantEffRank  Rank
		wantEffSuit  Suit
		wantErr      bool
	}{
		// Straights
		{"Straight 3-7", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank5, Hearts), C(Rank6, Spades), C(Rank7, Diamonds)}, Straight, Rank7, Diamonds, false},
		{"Straight 10-A", Deck{C(Rank10, Spades), C(Jack, Diamonds), C(Queen, Clubs), C(King, Hearts), C(Ace, Spades)}, Straight, Ace, Spades, false},
		{"Straight J-2 (highest)", Deck{C(Jack, Hearts), C(Queen, Spades), C(King, Diamonds), C(Ace, Clubs), C(Two, Hearts)}, Straight, Two, Hearts, false},
		{"Invalid Straight A-5 (Ace low not standard)", Deck{C(Ace, Hearts), C(Two, Spades), C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank5, Hearts)}, InvalidHand, -1, -1, true},
		{"Invalid Straight K-A-2-3-4 (wrap around)", Deck{C(King, Hearts), C(Ace, Spades), C(Two, Diamonds), C(Rank3, Clubs), C(Rank4, Hearts)}, InvalidHand, -1, -1, true},
		{"Not a straight (skip rank)", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank5, Hearts), C(Rank7, Spades), C(Rank8, Diamonds)}, InvalidHand, -1, -1, true},

		// Flushes
		{"Flush Diamonds (K high)", Deck{C(Rank3, Diamonds), C(Rank5, Diamonds), C(Rank7, Diamonds), C(Jack, Diamonds), C(King, Diamonds)}, Flush, King, Diamonds, false},
		{"Flush Spades (Ace high)", Deck{C(Rank4, Spades), C(Rank6, Spades), C(Rank9, Spades), C(Queen, Spades), C(Ace, Spades)}, Flush, Ace, Spades, false},
		{"Not a flush (one suit wrong)", Deck{C(Rank3, Diamonds), C(Rank5, Diamonds), C(Rank7, Clubs), C(Jack, Diamonds), C(King, Diamonds)}, InvalidHand, -1, -1, true},

		// Full Houses
		{"Full House (3s over 2s)", Deck{C(Rank3, Diamonds), C(Rank3, Clubs), C(Rank3, Hearts), C(Rank4, Spades), C(Rank4, Diamonds)}, FullHouse, Rank3, -1, false},
		{"Full House (As over Ks)", Deck{C(Ace, Spades), C(Ace, Hearts), C(King, Diamonds), C(King, Clubs), C(King, Spades)}, FullHouse, King, -1, false},
		{"Not Full House (two pairs)", Deck{C(Rank3, Diamonds), C(Rank3, Clubs), C(Rank4, Hearts), C(Rank4, Spades), C(Rank5, Diamonds)}, InvalidHand, -1, -1, true},

		// Four of a Kind + One (Bomb)
		{"Four of a Kind (7s and a 3)", Deck{C(Rank7, Diamonds), C(Rank7, Clubs), C(Rank7, Hearts), C(Rank7, Spades), C(Rank3, Diamonds)}, FourOfAKindPlusOne, Rank7, -1, false},
		{"Four of a Kind (As and a K)", Deck{C(Ace, Diamonds), C(Ace, Clubs), C(King, Hearts), C(Ace, Spades), C(Ace, Hearts)}, FourOfAKindPlusOne, Ace, -1, false},

		// Straight Flushes (Bomb)
		{"Straight Flush Diamonds 3-7", Deck{C(Rank3, Diamonds), C(Rank4, Diamonds), C(Rank5, Diamonds), C(Rank6, Diamonds), C(Rank7, Diamonds)}, StraightFlush, Rank7, Diamonds, false},
		{"Straight Flush Spades 10-A", Deck{C(Rank10, Spades), C(Jack, Spades), C(Queen, Spades), C(King, Spades), C(Ace, Spades)}, StraightFlush, Ace, Spades, false},
		{"Straight Flush Hearts J-2", Deck{C(Jack, Hearts), C(Queen, Hearts), C(King, Hearts), C(Ace, Hearts), C(Two, Hearts)}, StraightFlush, Two, Hearts, false},
		{"Invalid SF A-5 (Ace low not standard)", Deck{C(Ace, Hearts), C(Two, Hearts), C(Rank3, Hearts), C(Rank4, Hearts), C(Rank5, Hearts)}, InvalidHand, -1, -1, true},
		{"Not a Straight Flush (Straight, but not Flush)", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank5, Diamonds), C(Rank6, Diamonds), C(Rank7, Diamonds)}, Straight, Rank7, Diamonds, false},
		{"Not a Straight Flush (Flush, but not Straight)", Deck{C(Rank3, Diamonds), C(Rank5, Diamonds), C(Rank7, Diamonds), C(Jack, Diamonds), C(King, Diamonds)}, Flush, King, Diamonds, false},
		{"Flush Clubs (7 high)", Deck{C(Two, Clubs), C(Rank3, Clubs), C(Rank4, Clubs), C(Rank5, Clubs), C(Rank7, Clubs)}, Flush, Rank7, Clubs, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cards.Sort() // Essential for 5-card hand detection logic
			playedHand, err := re.DeterminePlayedHand(tc.cards)

			if (err != nil) != tc.wantErr {
				t.Errorf("DeterminePlayedHand() error = %v, wantErr %v for cards %v", err, tc.wantErr, tc.cards)
				return
			}
			if !tc.wantErr {
				if playedHand.HandType != tc.wantHandType {
					t.Errorf("DeterminePlayedHand() HandType = %s, want %s for cards %v", playedHand.HandType, tc.wantHandType, tc.cards)
				}
				if playedHand.EffectiveRank != tc.wantEffRank {
					t.Errorf("DeterminePlayedHand() EffectiveRank = %s, want %s for cards %v", playedHand.EffectiveRank, tc.wantEffRank, tc.cards)
				}
				if playedHand.EffectiveSuit != tc.wantEffSuit {
					// For hand types where suit doesn't matter (e.g., FullHouse, FOAK), wantEffSuit is -1.
					// The actual effectiveSuit in PlayedHand might also be -1 or a card's suit; comparison should be fine.
					t.Errorf("DeterminePlayedHand() EffectiveSuit = %s, want %s for cards %v", playedHand.EffectiveSuit, tc.wantEffSuit, tc.cards)
				}
			}
		})
	}
}

// TODO: Add individual tests for helper methods like isStraight, isFlush, etc., if desired for granularity.

func TestBigTwoRuleEngine_isStraight(t *testing.T) {
	re := NewBigTwoRuleEngine()

	tests := []struct {
		name           string
		cards          Deck
		wantIsStraight bool
		wantEffRank    Rank
		wantEffSuit    Suit // Suit of the highest card that determines the straight's rank
	}{
		{"Empty deck", Deck{}, false, -1, -1},
		{"Too few cards for straight", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank5, Hearts), C(Rank6, Spades)}, false, -1, -1},
		{"Valid Straight 3-7 (7D)", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank5, Hearts), C(Rank6, Spades), C(Rank7, Diamonds)}, true, Rank7, Diamonds},
		{"Valid Straight 10-A (AS)", Deck{C(Rank10, Spades), C(Jack, Diamonds), C(Queen, Clubs), C(King, Hearts), C(Ace, Spades)}, true, Ace, Spades},
		{"Valid Straight A-5 (5H, Ace low - specific rule for some games, Big Two typically doesn't use this, but testing helper behavior)", Deck{C(Ace, Clubs), C(Two, Spades), C(Rank3, Diamonds), C(Rank4, Hearts), C(Rank5, Hearts)}, true, Rank5, Hearts}, // Assumes A-5 is treated as 5-high
		{"Valid Straight J-2 (2H, 2 highest)", Deck{C(Jack, Hearts), C(Queen, Spades), C(King, Diamonds), C(Ace, Clubs), C(Two, Hearts)}, true, Two, Hearts},
		{"Invalid Straight (duplicates)", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank4, Hearts), C(Rank5, Spades), C(Rank6, Diamonds)}, false, -1, -1},
		{"Invalid Straight (gap)", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank6, Hearts), C(Rank7, Spades), C(Rank8, Diamonds)}, false, -1, -1},
		{"Invalid Straight (wrap around K-A-2-3-4)", Deck{C(King, Hearts), C(Ace, Spades), C(Two, Diamonds), C(Rank3, Clubs), C(Rank4, Hearts)}, false, -1, -1},
		// {"Valid Straight but with 6 cards (should detect 5-card straight within)", Deck{C(Rank3, Diamonds), C(Rank4, Clubs), C(Rank5, Hearts), C(Rank6, Spades), C(Rank7, Diamonds), C(Rank8, Clubs)}, true, Rank7, Diamonds}, // Depends on isStraight only looking at first 5 sorted cards
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cards.Sort() // Helper function expects sorted cards
			isStraight, effRank, effSuit := re.isStraight(tc.cards)

			if isStraight != tc.wantIsStraight {
				t.Errorf("isStraight() gotIsStraight = %v, want %v for cards %v", isStraight, tc.wantIsStraight, tc.cards)
			}
			if isStraight { // Only check rank and suit if it's meant to be a straight
				if effRank != tc.wantEffRank {
					t.Errorf("isStraight() gotEffRank = %s, want %s for cards %v", effRank, tc.wantEffRank, tc.cards)
				}
				if effSuit != tc.wantEffSuit {
					t.Errorf("isStraight() gotEffSuit = %s, want %s for cards %v", effSuit, tc.wantEffSuit, tc.cards)
				}
			}
		})
	}
}

func TestBigTwoRuleEngine_isFlush(t *testing.T) {
	re := NewBigTwoRuleEngine()

	tests := []struct {
		name        string
		cards       Deck
		wantIsFlush bool
		wantEffRank Rank // Rank of the highest card in the flush
		wantEffSuit Suit // Suit of the flush
	}{
		{"Empty deck", Deck{}, false, -1, -1},
		{"Too few cards for flush", Deck{C(Rank3, Diamonds), C(Rank4, Diamonds), C(Rank5, Diamonds), C(Rank6, Diamonds)}, false, -1, -1},
		{"Valid Flush Diamonds (K high)", Deck{C(Rank3, Diamonds), C(Rank5, Diamonds), C(Rank7, Diamonds), C(Jack, Diamonds), C(King, Diamonds)}, true, King, Diamonds},
		{"Valid Flush Spades (Ace high)", Deck{C(Rank4, Spades), C(Rank6, Spades), C(Rank9, Spades), C(Queen, Spades), C(Ace, Spades)}, true, Ace, Spades},
		{"Invalid Flush (mixed suits)", Deck{C(Rank3, Diamonds), C(Rank5, Diamonds), C(Rank7, Clubs), C(Jack, Diamonds), C(King, Diamonds)}, false, -1, -1},
		{"Invalid Flush (4 Diamonds, 1 Club)", Deck{C(Two, Diamonds), C(Rank4, Diamonds), C(Rank6, Diamonds), C(Rank8, Diamonds), C(King, Clubs)}, false, -1, -1},
		//{"Valid Flush with 6 cards (should detect 5-card flush within)", Deck{C(Two, Hearts), C(Rank3, Hearts), C(Rank5, Hearts), C(Rank7, Hearts), C(Jack, Hearts), C(King, Clubs)}, true, Jack, Hearts}, // Depends on isFlush only looking at first 5 sorted cards
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.cards.Sort() // Helper function expects sorted cards
			isFlush, effSuit, effRank := re.isFlush(tc.cards)

			if isFlush != tc.wantIsFlush {
				t.Errorf("isFlush() gotIsFlush = %v, want %v for cards %v", isFlush, tc.wantIsFlush, tc.cards)
			}
			if isFlush { // Only check rank and suit if it's meant to be a flush
				if effSuit != tc.wantEffSuit {
					t.Errorf("isFlush() gotEffRank = %s, want %s for cards %v", effRank, tc.wantEffRank, tc.cards)
				}
				if effRank != tc.wantEffRank {
					t.Errorf("isFlush() gotEffSuit = %s, want %s for cards %v", effSuit, tc.wantEffSuit, tc.cards)
				}
			}
		})
	}
}
