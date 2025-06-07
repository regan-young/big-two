<template>
  <div class="players-area-vue">
    <h2>Players</h2>
    <ul v-if="playersInfo && playersInfo.length > 0">
      <li v-for="player in playersInfo" :key="player.id" :class="{
          'current-player': player.id === currentPlayerId,
          'is-you': player.id === yourPlayerId,
          'has-passed': player.hasPassed,
      }">
        <span>{{ player.name }} ({{ player.cardCount }} cards)</span>
        <span v-if="player.id === yourPlayerId"> (You)</span>
        <span v-if="player.hasPassed && player.id !== currentPlayerId"> - PASSED</span>
      </li>
    </ul>
    <p v-else>No player information available.</p>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { PlayerInfo } from '@/types';

export default defineComponent({
  name: 'PlayersArea',
  props: {
    playersInfo: {
      type: Array as PropType<readonly PlayerInfo[]>,
      required: true,
    },
    yourPlayerId: {
      type: String as PropType<string | null>,
      default: null,
    },
    currentPlayerId: {
      type: String as PropType<string | null>,
      default: null,
    },
  },
  setup() {
    // No reactive logic needed within this component for now, it just displays props.
    return {};
  },
});
</script>

<style scoped>
.players-area-vue {
  border: 1px solid #ddd;
  padding: 10px;
  margin-bottom: 10px;
  background-color: #f9f9f9;
}
.players-area-vue h2 {
  margin-top: 0;
  font-size: 1.2em;
  color: #333;
  margin-bottom: 8px;
}
.players-area-vue ul {
  list-style-type: none;
  padding: 0;
  margin: 0;
}
.players-area-vue li {
  padding: 6px 3px;
  border-bottom: 1px dotted #eee;
  font-size: 0.9em;
}
.players-area-vue li:last-child {
  border-bottom: none;
}
.current-player {
  font-weight: bold;
  background-color: #d4edda; /* Light green for current player's turn */
  border-left: 3px solid #28a745;
  padding-left: 5px; /* Add some padding to make the border visible */

}
.is-you:not(.current-player) { /* Style for 'you' if it's not your turn */
  background-color: #e6ffed; 
}
.is-you.current-player { /* Specific style if it is 'you' AND your turn */
    background-color: #c8e6c9; /* A slightly different shade or more emphasis */
    font-weight: bold;
}
.has-passed {
  color: #777;
  font-style: italic;
}
</style> 