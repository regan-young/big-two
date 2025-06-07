<template>
  <div id="app">
    <!-- Game UI Wrapper -->
    <div v-if="gamePhase !== 'loading' && gamePhase !== 'ended'" class="game-layout-wrapper">
      
      <div class="score-and-rules-wrapper">
        <!-- Score Table -->
        <ScoreTable 
          :players-info="players"
          :scores="scores"
          :round-scores-history="roundScoresHistory"
          :round-number="roundNumber"
          :target-score="targetScore"
          :your-player-id="yourPlayerId"
        />
        <button @click="openRulesModal" class="rules-button-vue">Rules</button>
      </div>

      <!-- Top Game Row -->
      <div class="top-game-row-vue">
        <PlayersArea 
          :players-info="players"
          :your-player-id="yourPlayerId"
          :current-player-id="currentPlayerId"
        />
        <TableDisplay 
          :last-played-hand="lastPlayedHand"
          :current-player-name="currentPlayerName"
          :pass-count="passCount"
        />
        <MessagingArea 
          :chat-messages="chatMessages"
          :system-messages="formattedSystemMessages"
          @send-chat-message="handleSendChatMessage"
        />
      </div>
      <!-- End Top Game Row -->

      <!-- Player Hand Display Area -->
      <div id="player-hand-display-area">
          <h2>Your Hand <span v-if="isMyTurn" class="turn-indicator">(Your Turn!)</span></h2>
          <div id="player-controls-area-vue"> 
            <div id="player-action-messages-vue" class="messages-area">
              <p v-if="validationError" class="validation-error">{{ validationError }}</p>
            </div>
            <PlayerControls 
              :is-my-turn="isMyTurn"
              :is-game-over="isGameOver"
              :is-match-over="isMatchOver"
              :selected-cards-count="selectedCardsForAction.length"
              :auto-pass-enabled="autoPassEnabled"
              @sort-hand="handleSortHand"
              @play-selection="handlePlaySelection"
              @pass-turn="handlePassTurn"
              @toggle-auto-pass="handleToggleAutoPass"
            />
          </div>
          <PlayerHand 
            :hand="yourHand" 
            :is-game-over="isGameOver"
            :is-match-over="isMatchOver"
            @update:selectedCards="handleSelectedCardsUpdate"
          />
      </div>

    </div> <!-- End Game UI Wrapper -->

    <!-- Modals are outside the gameUiVisibleRef wrapper -->
    <GameOverModal 
      :is-visible="gameOverModalVisible"
      :winner-name="winnerName"
      :scores="scores"
      :players-info="players"
      :is-match-over="isMatchOver"
      :overall-winner-name="overallWinnerName"
      :round-scores-history="roundScoresHistory"
      @new-game="handleNewGame"
    />

    <RulesModal 
      :is-visible="rulesModalVisible"
      @close-rules="closeRulesModal"
    />

    <AliasModal 
      :is-visible="aliasModalVisible"
      :error-message="aliasError"
      @submit-alias="handleAliasSubmit"
    />

  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, computed, ref, watch } from 'vue';
import { useGameStore } from '@/stores/game';
import { storeToRefs } from 'pinia';
import { Card } from '@/types';
import { useWebSocket } from '@/composables/useWebSocket';

// Import all components
import PlayerHand from '@/components/PlayerHand.vue';
import PlayersArea from '@/components/PlayersArea.vue';
import TableDisplay from '@/components/TableDisplay.vue';
import MessagingArea from '@/components/MessagingArea.vue';
import PlayerControls from '@/components/PlayerControls.vue';
import ScoreTable from '@/components/ScoreTable.vue';
import AliasModal from '@/components/AliasModal.vue';
import RulesModal from '@/components/RulesModal.vue';
import GameOverModal from '@/components/GameOverModal.vue';

export default defineComponent({
  name: 'App',
  components: {
    PlayerHand, PlayersArea, TableDisplay, MessagingArea, PlayerControls, ScoreTable, AliasModal, RulesModal, GameOverModal,
  },
  setup() {
    const gameStore = useGameStore();
    const { sendMessage, lastMessage, connect } = useWebSocket();

    // Reactive state from Pinia store
    const {
      players,
      currentPlayerId,
      lastPlayedHand,
      gamePhase,
      scores,
      roundScoresHistory,
      roundNumber,
      targetScore,
      passCount,
      isGameOver,
      isMatchOver,
      yourPlayerId,
      chatMessages,
      systemMessages,
      autoPassEnabled,
      errorMessages,
      validationError,
    } = storeToRefs(gameStore);

    // Local UI state
    const selectedCardsForAction = ref<Card[]>([]);
    const rulesModalVisible = ref(false);
    const gameOverModalVisible = ref(false);
    const aliasModalVisible = computed(() => !yourPlayerId.value);
    const aliasError = ref<string | null>(null);
    const winnerName = ref<string | null>(null);
    const overallWinnerName = ref<string | null>(null);

    // Computed properties derived from state
    const currentPlayer = computed(() => players.value.find(p => p.id === currentPlayerId.value));
    const currentPlayerName = computed(() => currentPlayer.value?.name || 'Unknown');
    const isMyTurn = computed(() => currentPlayerId.value === yourPlayerId.value && !isGameOver.value);
    const yourHand = computed(() => {
        const me = players.value.find(p => p.id === yourPlayerId.value);
        return me ? me.hand : [];
    });
    const formattedSystemMessages = computed(() => {
      const normalMessages = systemMessages.value.map(msg => ({
        content: msg.content,
        isError: false,
        type: 'general' as const // Use 'as const' for literal type
      }));
      const errorMessagesFormatted = errorMessages.value.map(msg => ({
        content: msg.content,
        isError: true,
        type: 'general' as const // Errors are also a type of system message
      }));
      return [...normalMessages, ...errorMessagesFormatted];
    });

    // --- WebSocket Logic ---
    onMounted(() => {
      // Construct the WebSocket URL dynamically
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      const wsUrl = `${protocol}//${host}/ws`;

      // Connect to WebSocket when component mounts
      connect(wsUrl);
    });

    // Watch for incoming messages and update the store
    watch(lastMessage, (message) => {
      if (!message) return;
      gameStore.processWebSocketMessage(message);
    });

    // Watch for game over state to show modal
    watch(isGameOver, (newVal) => {
        if (newVal) {
            const winner = players.value.find(p => p.hand.length === 0);
            winnerName.value = winner ? winner.name : 'Unknown';
            // Logic to determine overall winner if match is over
            if (isMatchOver.value) {
                // This logic should probably be in the store
                let maxScore = -1;
                let winner = null;
                for(const playerId in scores.value) {
                    if(scores.value[playerId] > maxScore) {
                        maxScore = scores.value[playerId];
                        winner = players.value.find(p => p.id === playerId);
                    }
                }
                overallWinnerName.value = winner ? winner.name : 'Unknown';
            }
            gameOverModalVisible.value = true;
        } else {
            gameOverModalVisible.value = false;
        }
    });

    // --- Event Handlers ---
    const handleSelectedCardsUpdate = (newSelectedCards: Card[]) => {
      if (validationError.value) {
        gameStore.clearValidationError();
      }
      selectedCardsForAction.value = newSelectedCards;
    };

    const handlePlaySelection = () => {
      if (selectedCardsForAction.value.length > 0) {
        sendMessage({ type: 'playCards', cards: selectedCardsForAction.value });
        selectedCardsForAction.value = []; // Clear selection after playing
      }
    };

    const handlePassTurn = () => sendMessage({ type: 'passTurn' });
    const handleSortHand = (payload: { preference: 'rank' | 'suit' }) => gameStore.sortHand(payload.preference);
    const handleToggleAutoPass = () => gameStore.toggleAutoPass();
    const handleSendChatMessage = (content: string) => sendMessage({ type: 'chat', content });
    const handleAliasSubmit = (alias: string) => sendMessage({ type: 'setAlias', alias });
    const handleNewGame = () => sendMessage({ type: 'newGame' });

    return {
      // State & Computed
      players, currentPlayerId, lastPlayedHand, gamePhase, scores, roundScoresHistory, roundNumber, targetScore,
      passCount, isGameOver, isMatchOver, yourPlayerId, chatMessages, systemMessages, autoPassEnabled,
      errorMessages, validationError,
      selectedCardsForAction, rulesModalVisible, gameOverModalVisible, aliasModalVisible, aliasError,
      winnerName, overallWinnerName, currentPlayerName, isMyTurn, yourHand,
      // Methods
      handleSelectedCardsUpdate, handlePlaySelection, handlePassTurn, handleSortHand, handleToggleAutoPass,
      handleSendChatMessage, handleAliasSubmit, handleNewGame,
      openRulesModal: () => rulesModalVisible.value = true,
      closeRulesModal: () => rulesModalVisible.value = false,
      formattedSystemMessages,
    };
  },
});
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  border: none;
  padding: 0 15px 15px 15px;
  gap: 20px;
}

.game-layout-wrapper {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.top-game-row-vue {
  display: flex;
  justify-content: space-between;
  gap: 15px;
}

.top-game-row-vue > * {
  flex: 1 1 0;
  min-width: 0;
}

.score-and-rules-wrapper {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  gap: 20px;
}
.rules-button-vue {
  margin-top: 5px; /* Adjust as needed */
  height: fit-content;
}
.validation-error {
  color: #dc3545; /* Red for errors */
  font-weight: bold;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
  100% {
    opacity: 1;
  }
}

.turn-indicator {
  color: #28a745;
  vertical-align: text-top;
}

#player-controls-area-vue {
  padding: 0 0 10px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
}
</style> 