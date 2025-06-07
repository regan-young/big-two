<template>
  <div class="player-hand-vue cards-display-area">
    <p v-if="!hand || hand.length === 0">No cards in hand.</p>
    <svg
      v-for="(card, index) in hand"
      :key="`${card.rank}-${card.suit}`"
      class="card"
      :class="{ selected: isSelected(card) }"
      :data-rank="card.rank"
      :data-suit="card.suit"
      viewBox="0 0 169.075 244.64" 
      @click="toggleCardSelection(card)"
    >
      <use :xlink:href="`#${suitMap[card.suit]}_${rankMap[card.rank]}`"></use>
    </svg>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, ref, watch, computed } from 'vue';
import { Card } from '@/types';
import { rankMap, suitMap } from '@/utils/cardUtils';

export default defineComponent({
  name: 'PlayerHand',
  props: {
    hand: {
      type: Array as PropType<readonly Card[]>,
      required: true,
    },
    isGameOver: {
        type: Boolean,
        default: false,
    },
    isMatchOver: {
        type: Boolean,
        default: false,
    }
  },
  emits: ['update:selectedCards'], // To inform parent about selection changes

  setup(props, { emit }) {
    const selectedCardsRef = ref<Card[]>([]); // Store actual Card objects

    const canSelectCards = computed(() => !props.isGameOver && !props.isMatchOver);

    const toggleCardSelection = (card: Card) => {
      if (!canSelectCards.value) return;

      const index = selectedCardsRef.value.findIndex(
        (c) => c.rank === card.rank && c.suit === card.suit
      );
      if (index === -1) {
        selectedCardsRef.value.push(card);
      } else {
        selectedCardsRef.value.splice(index, 1);
      }
      emit('update:selectedCards', [...selectedCardsRef.value]); // Emit a copy
    };

    const isSelected = (card: Card): boolean => {
      return selectedCardsRef.value.some((c) => c.rank === card.rank && c.suit === card.suit);
    };

    // Watch for changes in the hand prop to clear selections if the hand itself changes drastically
    // (e.g., new round, cards played). This new logic preserves selection during sorting.
    watch(() => props.hand, (newHand, oldHand) => {
      if (newHand.length !== oldHand.length) {
        selectedCardsRef.value = [];
        emit('update:selectedCards', []);
        return;
      }

      // If lengths are the same, check if it's the same set of cards.
      // This distinguishes sorting from other actions like playing cards.
      const oldCardIds = oldHand.map(c => `${c.rank}-${c.suit}`).sort();
      const newCardIds = newHand.map(c => `${c.rank}-${c.suit}`).sort();

      if (JSON.stringify(oldCardIds) !== JSON.stringify(newCardIds)) {
        selectedCardsRef.value = [];
        emit('update:selectedCards', []);
      }
    }, { deep: true });

    return {
      toggleCardSelection,
      isSelected,
      selectedCards: selectedCardsRef, // Expose for parent if needed (though event is preferred for changes)
      rankMap, // Use rankMap in template
      suitMap, // Use suitMap in template
    };
  },
});
</script>

<style scoped>
.player-hand-vue {
  /* Styles from original #player-hand can be migrated here or kept global */
  /* display: flex; flex-wrap: wrap; justify-content: center; etc. */
  min-height: 150px; /* Example, adjust as needed */
  padding-top: 30px;
}

.card {
  width: 120px; /* Example size, adjust */
  height: auto;
  margin: 2px;
  margin-left: -50px;
  cursor: pointer;
  transition: transform 0.1s ease-out, box-shadow 0.1s ease-out;
  transform: translateY(0px) scale(1.03);
  border-radius: 5px; /* Optional: if you want rounded borders on the SVG container */
}

.player-hand-vue > .card:first-child {
  margin-left: 0;
}

.card:hover {
  transform: translateY(-5px) scale(1.03);
  box-shadow: 0px 4px 8px rgba(0,0,0,0.2);
}

.card.selected {
  transform: translateY(-20px) scale(1.05);
  box-shadow: 0 0 10px 3px gold; 
  /* border: 2px solid gold; */ /* Alternative selection indicator */
}

/* Ensure cards-display-area styles from style.css are considered */
/* cards-display-area in original style.css:
.cards-display-area {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: flex-start; 
    padding: 5px;
    border: 1px solid #ccc;
    border-radius: 4px;
    background-color: #f9f9f9;
    min-height: 120px; 
    max-height: 280px; 
    overflow-y: auto;
}
*/
</style> 