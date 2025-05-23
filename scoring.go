package main

// CalculateScores calculates the scores for each player at the end of the game.
// For now, it's a placeholder: winner gets 0, others get the number of cards remaining.
func CalculateScores(game *GameState) map[string]int {
	scores := make(map[string]int)
	if game == nil || game.Players == nil || game.WinnerID == "" {
		return scores // Return empty scores if game state is invalid for scoring
	}

	for _, player := range game.Players {
		if player.ID == game.WinnerID {
			scores[player.ID] = 0
		} else {
			// Basic scoring: number of cards left.
			// More complex scoring (e.g., double for 10+ cards) can be added later.
			score := len(player.Hand)
			if score >= 10 && score < 13 {
				score *= 2 // Double if 10, 11, 12 cards
			} else if score == 13 {
				score *= 3 // Triple if all 13 cards left
			}
			scores[player.ID] = score
		}
	}
	return scores
}
