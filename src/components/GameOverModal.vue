<template>
  <div v-if="isVisible" class="modal-backdrop game-over-modal-vue">
    <div class="modal-content">
      <h2 id="game-over-title">{{ title }}</h2>
      <p v-if="winnerName && !isMatchOver">Round Winner: <strong>{{ winnerName }}</strong></p>
      <p v-if="overallWinnerName && isMatchOver">CONGRATULATIONS, <strong>{{ overallWinnerName }}</strong>, YOU WON THE MATCH!</p>
      <p v-else-if="isMatchOver && !overallWinnerName">The match has ended. Calculating final results...</p>

      <div v-if="playersInfo && playersInfo.length > 0" class="final-scores-container">
        <h3>{{ isMatchOver ? 'Final Scores:' : 'Current Scores:' }}</h3>
        <ul class="player-scores-list">
          <li v-for="player in sortedPlayersByScore" :key="player.id">
            {{ player.name }}: {{ scores[player.id] || 0 }}
          </li>
        </ul>
      </div>
      
      <div v-if="isMatchOver && roundScoresHistory && roundScoresHistory.length > 0" class="match-round-history">
          <h4>Match Round Summary</h4>
          <table>
              <thead>
                  <tr>
                      <th>Round</th>
                      <th v-for="player in playersInfo" :key="player.id">{{ getPlayerName(player.id)?.substring(0,6) }}..</th>
                      <th>Round Winner</th>
                  </tr>
              </thead>
              <tbody>
                  <tr v-for="round in roundScoresHistory" :key="round.roundNumber">
                      <td>{{ round.roundNumber }}</td>
                      <td v-for="player in playersInfo" :key="player.id">
                          {{ round.scores[player.id] !== undefined ? round.scores[player.id] : '-' }}
                      </td>
                      <td>{{ round.winnerId ? getPlayerName(round.winnerId) : '-' }}</td>
                  </tr>
              </tbody>
          </table>
      </div>

      <button @click="emitNewGame" id="new-game-button-vue">
        {{ isMatchOver ? 'Start New Match' : 'Start Next Round' }}
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, computed } from 'vue';
import { PlayerInfo, Scores, RoundResult } from '@/types';

export default defineComponent({
  name: 'GameOverModal',
  props: {
    isVisible: { type: Boolean, required: true },
    winnerName: { type: String as PropType<string | null>, default: null },
    scores: { type: Object as PropType<Readonly<Scores>>, default: () => ({}) },
    playersInfo: { type: Array as PropType<readonly PlayerInfo[]>, default: () => [] },
    isMatchOver: { type: Boolean, default: false },
    overallWinnerName: { type: String as PropType<string | null>, default: null },
    roundScoresHistory: { type: Array as PropType<readonly RoundResult[]>, default: () => [] },
  },
  emits: ['new-game'],
  setup(props, { emit }) {
    const title = computed(() => (props.isMatchOver ? 'Match Over!' : 'Round Over!'));

    const playerNamesById = computed(() => {
      const names: Record<string, string> = {};
      props.playersInfo.forEach(p => { names[p.id] = p.name; });
      return names;
    });
    
    const getPlayerName = (playerId: string | null | undefined): string => {
        if (!playerId) return 'N/A';
        return playerNamesById.value[playerId] || playerId;
    };

    const sortedPlayersByScore = computed(() => {
      if (!props.playersInfo) return [];
      return [...props.playersInfo].sort((a, b) => {
        const scoreA = props.scores[a.id] || 0;
        const scoreB = props.scores[b.id] || 0;
        // For Big Two, lower scores are better if that's the game rule.
        // Assuming higher score is better for now or it's just a display of points.
        // Adjust if lower score = winner for sorting.
        return scoreB - scoreA; // Sort descending by score
      });
    });

    const emitNewGame = () => {
      emit('new-game');
    };

    return {
      title,
      sortedPlayersByScore,
      emitNewGame,
      getPlayerName,
    };
  },
});
</script>

<style scoped>
/* Using similar modal styles as AliasModal and RulesModal */
.modal-backdrop {
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  background-color: rgba(0,0,0,0.7); display: flex;
  justify-content: center; align-items: center; z-index: 1050; /* Higher z-index */
}
.modal-content {
  background-color: #fff; padding: 25px 30px; border-radius: 8px;
  box-shadow: 0 5px 20px rgba(0,0,0,0.4); text-align: center;
  min-width: 400px; max-width: 650px; /* Wider for scores */
  max-height: 90vh; overflow-y: auto;
}
.modal-content h2 {
  margin-top: 0; margin-bottom: 15px; font-size: 1.8em; color: #333;
}
.modal-content p { margin-bottom: 10px; font-size: 1.1em; }
.modal-content strong { color: #28a745; } /* Green for winner emphasis */

.final-scores-container { margin-top: 20px; margin-bottom: 25px; }
.final-scores-container h3 { font-size: 1.3em; margin-bottom: 10px; color: #444; }
.player-scores-list { list-style: none; padding: 0; margin: 0 auto; max-width: 300px; }
.player-scores-list li {
  font-size: 1.1em; padding: 6px 0; border-bottom: 1px solid #eee;
  display: flex; justify-content: space-between;
}
.player-scores-list li:last-child { border-bottom: none; }

.match-round-history { margin-top: 20px; margin-bottom: 25px; }
.match-round-history h4 { font-size: 1.2em; margin-bottom: 10px; }
.match-round-history table { width: 100%; border-collapse: collapse; font-size: 0.9em; }
.match-round-history th, .match-round-history td {
  border: 1px solid #ddd; padding: 5px 7px; text-align: center;
}
.match-round-history th { background-color: #f0f0f0; }

#new-game-button-vue {
  padding: 12px 25px; background-color: #007bff; color: white;
  border: none; border-radius: 5px; cursor: pointer; font-size: 1.1em;
  transition: background-color 0.2s; margin-top: 15px;
}
#new-game-button-vue:hover { background-color: #0056b3; }
</style> 