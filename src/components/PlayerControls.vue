<template>
  <div class="player-controls-vue">
    <button @click="emitSort('rank')" :disabled="isGameOver || isMatchOver">Sort by Rank</button>
    <button @click="emitSort('suit')" :disabled="isGameOver || isMatchOver">Sort by Suit</button>
    <button 
      @click="emitPlay" 
      :disabled="!isMyTurn || selectedCardsCount === 0 || isGameOver || isMatchOver"
      :class="{ 'can-play': isMyTurn && selectedCardsCount > 0 && !isGameOver && !isMatchOver }"
    >
      Play Selected ({{ selectedCardsCount }})
    </button>
    <button 
      @click="emitPass" 
      :disabled="!isMyTurn || isGameOver || isMatchOver"
      :class="{ 'auto-pass-enabled': autoPassEnabled && isMyTurn }"
    >
      {{ autoPassEnabled && isMyTurn ? 'Auto-Passing...' : (isMyTurn ? 'Pass Turn' : 'Pass') }}
    </button>
    <!-- Basic toggle for auto-pass, can be improved -->
    <label class="auto-pass-toggle">
      <input type="checkbox" :checked="autoPassEnabled" @change="emitToggleAutoPass" :disabled="isGameOver || isMatchOver">
      Auto-Pass Next Turn
    </label>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

export default defineComponent({
  name: 'PlayerControls',
  props: {
    isMyTurn: { type: Boolean, default: false },
    isGameOver: { type: Boolean, default: false },
    isMatchOver: { type: Boolean, default: false },
    // canPlaySelection: { type: Boolean, default: false }, // Simplified for now
    selectedCardsCount: { type: Number, default: 0 },
    autoPassEnabled: { type: Boolean, default: false },
  },
  emits: ['sort-hand', 'play-selection', 'pass-turn', 'toggle-auto-pass'],
  setup(props, { emit }) {
    const emitSort = (preference: 'rank' | 'suit') => {
      emit('sort-hand', { preference });
    };
    const emitPlay = () => {
      emit('play-selection');
    };
    const emitPass = () => {
      emit('pass-turn');
    };
    const emitToggleAutoPass = (event: Event) => {
        const target = event.target as HTMLInputElement;
        emit('toggle-auto-pass', target.checked);
    }

    return {
      emitSort,
      emitPlay,
      emitPass,
      emitToggleAutoPass,
    };
  },
});
</script>

<style scoped>
.player-controls-vue {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 10px;
  border: 1px solid #28a745; /* Green border for distinction */
  background-color: #f0fff0;
  border-radius: 4px;
  align-items: center;
}
.player-controls-vue button {
  padding: 8px 12px;
  border: 1px solid #ccc;
  background-color: #f8f8f8;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s, border-color 0.2s;
}
.player-controls-vue button:hover:not(:disabled) {
  background-color: #e9e9e9;
  border-color: #bbb;
}
.player-controls-vue button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
.player-controls-vue button.can-play {
    border-color: #28a745;
    background-color: #d4edda;
    color: #155724;
}
.player-controls-vue button.auto-pass-enabled {
    border-color: #ffc107;
    background-color: #fff3cd;
}
.auto-pass-toggle {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 0.9em;
    cursor: pointer;
}
.auto-pass-toggle input[type="checkbox"] {
    cursor: pointer;
}
</style> 