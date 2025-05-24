// Card Rank and Suit mapping (from card.go enums)
//Go Ranks: Three=3, Four=4, ..., Queen=12, King=13, Ace=14, Two=15
//SVG Ranks: 1 (Ace), 2-10, jack, queen, king
export const rankMap: { [key: number]: string } = {
    3: '3', 4: '4', 5: '5', 6: '6', 7: '7', 8: '8', 9: '9', 10: '10',
    11: 'jack', 12: 'queen', 13: 'king',
    14: '1',  // Ace (game value 14) is '1' in SVG
    15: '2'   // Two (game value 15) is '2' in SVG
};

//Go Suits: Diamonds=0, Clubs=1, Hearts=2, Spades=3
//SVG Suits: diamond, club, heart, spade
export const suitMap: { [key: number]: string } = {
    0: 'diamond',
    1: 'club',
    2: 'heart',
    3: 'spade'
}; 