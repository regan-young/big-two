<template>
  <div class="score-table-vue">
    <h3>Scores (Round {{ roundNumber }} / Target: {{ targetScore }})</h3>
    <table class="current-scores-table">
      <thead>
        <tr>
          <th>Player</th>
          <th>Current Score</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="!playerNamesById || Object.keys(playerNamesById).length === 0">
          <td colspan="2">Waiting for player data...</td>
        </tr>
        <tr v-for="player in playersInfo" :key="player.id">
          <td>{{ player.name }} <span v-if="player.id === yourPlayerId">(You)</span></td>
          <td>{{ scores[player.id] || 0 }}</td>
        </tr>
      </tbody>
    </table>

    <div v-if="roundScoresHistory && roundScoresHistory.length > 0" class="round-history-container">
      <h4>Round History</h4>
      <table class="round-history-table">
        <thead>
          <tr>
            <th>Round</th>
            <th v-for="player in playersInfo" :key="player.id">{{ player.name.substring(0, 5) }}..</th>
            <th>Winner</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="roundResult in roundScoresHistory" :key="roundResult.roundNumber">
            <td>{{ roundResult.roundNumber }}</td>
            <td v-for="player in playersInfo" :key="player.id">
              {{ roundResult.scores[player.id] !== undefined ? roundResult.scores[player.id] : '-' }}
            </td>
            <td>{{ roundResult.winnerId ? playerNamesById[roundResult.winnerId] || 'N/A' : 'N/A' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <p v-else>No round history yet.</p>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, computed } from 'vue';
import { PlayerInfo, RoundResult, Scores } from '@/types'; // Assuming Scores is Record<string, number>

export default defineComponent({
  name: 'ScoreTable',
  props: {
    playersInfo: {
      type: Array as PropType<readonly PlayerInfo[]>,
      default: () => [],
    },
    scores: { // Current overall scores
      type: Object as PropType<Readonly<Scores>>,
      default: () => ({}),
    },
    roundScoresHistory: {
      type: Array as PropType<readonly RoundResult[]>,
      default: () => [],
    },
    roundNumber: {
      type: Number,
      default: 1,
    },
    targetScore: {
      type: Number,
      default: 100,
    },
    yourPlayerId: { // To highlight "You"
        type: String as PropType<string | null>,
        default: null
    }
  },
  setup(props) {
    const playerNamesById = computed(() => {
      const names: Record<string, string> = {};
      if (props.playersInfo) {
        props.playersInfo.forEach(p => {
          names[p.id] = p.name;
        });
      }
      return names;
    });

    return {
      playerNamesById,
    };
  },
});
</script>

<style scoped>
.score-table-vue {
  border: 1px solid #ffc107; /* Yellow/Orange border for distinction */
  padding: 5px;
  background-color: #fff9e6;
  border-radius: 4px;
  margin-bottom: 10px;
}
.score-table-vue h3, .score-table-vue h4 {
  margin-top: 0;
  font-size: 1em;
  color: #555;
  margin-bottom: 5px;
  border-bottom: 1px solid #eee;
  padding-bottom: 3px;
}
.score-table-vue h4 {
    font-size: 0.9em;
    margin-top: 10px;
}

table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85em;
}
th, td {
  border: 1px solid #ddd;
  padding: 4px 6px;
  text-align: left;
}
th {
  background-color: #f2f2f2;
  font-weight: bold;
}
.current-scores-table td:nth-child(2), .round-history-table td {
    text-align: right;
}
.round-history-container {
    margin-top: 10px;
}
.round-history-table th, .round-history-table td {
    font-size: 0.8em;
    padding: 3px 5px;
}
</style> 