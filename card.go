package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// --- Constants for Suits and Ranks ---

// Suit represents the suit of a card.
type Suit int

const (
	Diamonds Suit = iota // 0
	Clubs                // 1
	Hearts               // 2
	Spades               // 3
)

// String returns a string representation of the suit.
func (s Suit) String() string {
	return []string{"Diamonds", "Clubs", "Hearts", "Spades"}[s]
}

// Rank represents the rank of a card in Big 2.
// 3 is the lowest, 2 is the highest.
type Rank int

const (
	Rank3  Rank = 3
	Rank4  Rank = 4
	Rank5  Rank = 5
	Rank6  Rank = 6
	Rank7  Rank = 7
	Rank8  Rank = 8
	Rank9  Rank = 9
	Rank10 Rank = 10
	Jack   Rank = 11
	Queen  Rank = 12
	King   Rank = 13
	Ace    Rank = 14
	Two    Rank = 15 // Highest rank in Big 2
)

// String returns a string representation of the rank.
func (r Rank) String() string {
	switch r {
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ace:
		return "A"
	case Two:
		return "2"
	default:
		return fmt.Sprintf("%d", r)
	}
}

// Card represents a single playing card.
type Card struct {
	Rank Rank `json:"rank"`
	Suit Suit `json:"suit"`
}

// String returns a string representation of the card (e.g., "AS" for Ace of Spades).
func (c Card) String() string {
	suitStr := c.Suit.String()
	if len(suitStr) > 0 {
		return fmt.Sprintf("%s%c", c.Rank.String(), suitStr[0])
	}
	return fmt.Sprintf("%s?", c.Rank.String())
}

// Deck represents a collection of cards.
type Deck []Card

// --- Deck Operations ---

// NewDeck creates a standard 52-card deck.
func NewDeck() Deck {
	deck := make(Deck, 0, 52)
	suits := []Suit{Diamonds, Clubs, Hearts, Spades}
	ranks := []Rank{
		Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10,
		Jack, Queen, King, Ace, Two,
	}

	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{Rank: rank, Suit: suit})
		}
	}
	return deck
}

// Shuffle randomizes the order of cards in the deck.
func (d Deck) Shuffle() {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	for i := len(d) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		d[i], d[j] = d[j], d[i]
	}
}

// Deal removes and returns the top 'n' cards from the deck.
// Returns the dealt cards and a boolean indicating success (e.g., enough cards).
func (d *Deck) Deal(n int) (Deck, bool) {
	if len(*d) < n {
		return nil, false
	}
	dealtCards := (*d)[:n]
	*d = (*d)[n:]
	return dealtCards, true
}

// Sorts a deck of cards (typically a player's hand).
// Big 2 sort order: Rank (3 low, 2 high), then Suit (Diamonds low, Spades high).
func (d Deck) Sort() {
	sort.SliceStable(d, func(i, j int) bool {
		if d[i].Rank == d[j].Rank {
			return d[i].Suit < d[j].Suit
		}
		return d[i].Rank < d[j].Rank
	})
}

// String returns a string representation of the deck.
func (d Deck) String() string {
	if len(d) == 0 {
		return "[]"
	}
	cardStrings := make([]string, len(d))
	for i, card := range d {
		cardStrings[i] = card.String()
	}
	return "[" + strings.Join(cardStrings, ", ") + "]"
}
