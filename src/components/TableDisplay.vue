<template>
  <div class="table-display-vue">
    <h2>Table</h2>
    <div class="last-played-hand-info-vue">
      <span v-if="lastPlayedHand && lastPlayedHand.playerId">
        Last played by: {{ lastPlayedHand.playerId }} ({{ lastPlayedHand.handType }})
      </span>
      <span v-else-if="lastPlayedHand && lastPlayedHand.cards && lastPlayedHand.cards.length > 0">
        Last played by: {{ lastPlayedHand.playerId }} ({{ lastPlayedHand.handTypeString }})
      </span>
      <span v-else>
        Table is clear.
      </span>
    </div>
    <div class="last-played-cards-vue cards-display-area">
      <svg
        v-if="lastPlayedHand && lastPlayedHand.cards && lastPlayedHand.cards.length > 0"
        v-for="card in lastPlayedHand.cards"
        :key="`${card.rank}-${card.suit}`"
        class="card"
        :data-rank="card.rank"
        :data-suit="card.suit"
        viewBox="0 0 169.075 244.64"
      >
        <use :xlink:href="`#${suitMap[card.suit]}_${rankMap[card.rank]}`"></use>
      </svg>
      <svg v-else viewBox="0 0 169.075 244.64" class="card card-back">
          <use xlink:href="#back"></use>
      </svg>
    </div>
    <div class="turn-info-vue">
      <span v-if="currentPlayerName">Current Turn: {{ currentPlayerName }}</span>
      <span v-else>Waiting for turn information...</span>
    </div>
    <div class="pass-count-info-vue">
      <span>Passes: {{ passCount }}</span>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, computed } from 'vue';
import { PlayedHand, Card } from '@/types';
import { rankMap, suitMap } from '@/utils/cardUtils';

export default defineComponent({
  name: 'TableDisplay',
  props: {
    lastPlayedHand: {
      type: Object as PropType<PlayedHand | null>,
      default: null,
    },
    currentPlayerName: {
      type: String as PropType<string | null>,
      default: null,
    },
    passCount: {
      type: Number,
      default: 0,
    },
  },
  setup(props) {
    return {
      rankMap, // Expose to template
      suitMap, // Expose to template
    };
  },
});
</script>

<style scoped>
.table-display-vue {
  border: 1px solid #ddd;
  padding: 10px;
  margin-bottom: 10px;
  background-color: #f0f0f0; /* Slightly different background */
}
.table-display-vue h2 {
  margin-top: 0;
  font-size: 1.2em;
  color: #333;
  margin-bottom: 8px;
}
.last-played-hand-info-vue, .turn-info-vue, .pass-count-info-vue {
  font-size: 0.9em;
  margin-bottom: 5px;
  min-height: 1.2em; /* Ensure space even if empty */
}
.last-played-cards-vue {
  /* This will use global .cards-display-area styles, plus specific ones below */
  min-height: 100px; /* Adjust as needed */
  border: 1px solid #ccc; /* Keep border from original for visual grouping */
  background-color: #e9e9e9; /* Slightly darker than parent */
}
.card {
  width: 120px; /* Smaller cards for table display */
  height: auto;
  margin: 2px 2px 2px -40px;
}
.card:first-child {
  margin-left: 0;
}
.card-back {
    margin-left: 0; /* Center the card back */
    fill: #555; /* Example, if your card_back SVG needs a fill */
    /* Or use an <image> tag within the SVG definition if it's a raster image */
}
</style> 