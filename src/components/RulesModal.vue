<template>
  <div v-if="isVisible" class="modal-backdrop rules-modal-vue">
    <div class="modal-content">
      <span @click="closeModal" class="modal-close-button">&times;</span>
      <h2>Big Two Rules</h2>
      <div>
        <p><strong>Objective:</strong> Be the first player to get rid of all your cards.</p>
        
        <h3>Card Ranking (High to Low):</h3>
        <p>2 > A > K > Q > J > 10 > 9 > 8 > 7 > 6 > 5 > 4 > 3</p>
        
        <h3>Suit Ranking (High to Low):</h3>
        <p>Spades (♠) > Hearts (♥) > Clubs (♣) > Diamonds (♦)</p>
        
        <h3>Starting Play:</h3>
        <p>The player holding the 3 of Diamonds (3♦) starts the first trick by playing it (usually as part of a valid hand).</p>
        
        <h3>Valid Hands to Play:</h3>
        <ul>
          <li><strong>Single:</strong> One card.</li>
          <li><strong>Pair:</strong> Two cards of the same rank.</li>
          <li><strong>Triple:</strong> Three cards of the same rank.</li>
          <li><strong>5-Card Hands (Poker Hands):</strong>
            <ul>
              <li><strong>Straight:</strong> Five cards in sequence (e.g., 3-4-5-6-7). A-2-3-4-5 is a valid straight (5 is high card). 2-3-4-5-6 is the highest straight.</li>
              <li><strong>Flush:</strong> Five cards of the same suit, not in sequence. Highest rank card determines strength, then suit of highest card.</li>
              <li><strong>Full House:</strong> A triple and a pair (e.g., 7-7-7-Q-Q). Rank of the triple determines strength.</li>
              <li><strong>Four of a Kind + One Card (Bomb):</strong> Four cards of the same rank, plus any fifth card (e.g., K-K-K-K-5). Beats any smaller hand except a Straight Flush.</li>
              <li><strong>Straight Flush (Bomb):</strong> Five cards in sequence of the same suit. The ultimate hand, beats all other hands.</li>
            </ul>
          </li>
        </ul>

        <h3>Playing the Game:</h3>
        <p>Subsequent players must play the same number of cards as the previous hand, forming a higher-ranking hand of the same type. If a player cannot or chooses not to play, they must pass. If all other players pass, the player who played the last hand starts a new trick with any valid hand.</p>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'RulesModal',
  props: {
    isVisible: {
      type: Boolean,
      required: true,
    },
  },
  emits: ['close-rules'],
  setup(props, { emit }) {
    const closeModal = () => {
      emit('close-rules');
    };

    return {
      closeModal,
    };
  },
});
</script>

<style scoped>
/* Re-using modal-backdrop and modal-content styles from AliasModal or global styles.css */
/* Specific styles for RulesModal content if needed */
.modal-backdrop { /* Copied from AliasModal for standalone functionality */
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000; 
}

.modal-content { /* Copied from AliasModal */
  background-color: #fff;
  padding: 20px 25px;
  border-radius: 8px;
  box-shadow: 0 5px 15px rgba(0,0,0,0.3);
  max-width: 600px; /* Rules can be wider */
  max-height: 80vh; /* Limit height and allow scroll */
  overflow-y: auto;
  text-align: left; /* Rules are better left-aligned */
  position: relative; /* For close button positioning */
}

.modal-close-button {
  position: absolute;
  top: 10px;
  right: 15px;
  font-size: 1.8em;
  font-weight: bold;
  color: #aaa;
  cursor: pointer;
  line-height: 1;
}
.modal-close-button:hover {
  color: #777;
}

.modal-content h2 {
  margin-top: 0;
  margin-bottom: 15px;
  font-size: 1.6em;
  color: #333;
  text-align: center;
}
.modal-content h3 {
  margin-top: 20px;
  margin-bottom: 8px;
  font-size: 1.2em;
  color: #444;
}
.modal-content p, .modal-content li {
  font-size: 0.95em;
  line-height: 1.6;
  color: #555;
}
.modal-content ul {
  padding-left: 20px; /* Indent lists */
  margin-bottom: 10px;
}
.modal-content ul ul {
  margin-top: 5px;
  margin-bottom: 5px;
}
.modal-content strong {
    color: #333;
}
</style> 