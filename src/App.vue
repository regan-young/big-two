<template>
  <div id="app">
    <!-- Game UI Wrapper -->
    <div v-if="gameUiVisibleRef" class="game-layout-wrapper">
      
      <div class="score-and-rules-wrapper">
        <!-- Score Table -->
        <ScoreTable 
          :players-info="playersInfoRef"
          :scores="scoresRef"
          :round-scores-history="roundScoresHistoryRef"
          :round-number="roundNumberRef"
          :target-score="targetScoreRef"
          :your-player-id="yourPlayerIdRef"
        />
        <button @click="openRulesModal" class="rules-button-vue">Rules</button>
      </div>

      <!-- Top Game Row -->
      <div class="top-game-row-vue" style="display: flex; justify-content: space-around; gap: 15px; margin-bottom: 15px;">
        <PlayersArea 
          :players-info="playersInfoRef"
          :your-player-id="yourPlayerIdRef"
          :current-player-id="currentPlayerIdRef"
          style="flex: 1;"
        />
        <TableDisplay 
          :last-played-hand="lastPlayedHandRef"
          :current-player-name="currentPlayerNameRef"
          :pass-count="passCountRef"
          style="flex: 1;"
        />
        <MessagingArea 
          :chat-messages="chatMessagesRef"
          :system-messages="systemMessagesRef"
          @send-chat-message="handleSendChatMessage"
          style="flex: 1;"
        />
      </div>
      <!-- End Top Game Row -->

      <!-- Player Hand Display Area -->
      <div id="player-hand-display-area" style="margin-bottom: 15px;">
          <h2>Your Hand</h2>
          <PlayerHand 
            :hand="currentPlayerHandRef" 
            :is-game-over="isGameOverRef"
            :is-match-over="isMatchOverRef"
            @update:selectedCards="handleSelectedCardsUpdate"
          />
      </div>

      <!-- Player Controls Area -->
      <div id="player-controls-area-vue" style="margin-top: 15px;"> 
        <h2>Controls</h2>
        <div id="player-action-messages-vue" class="messages-area" style="color: red; min-height: 1.2em; margin-bottom: 5px;">
        </div>
        <PlayerControls 
          :is-my-turn="isMyTurnRef"
          :is-game-over="isGameOverRef"
          :is-match-over="isMatchOverRef"
          :selected-cards-count="selectedCardsForAction.length"
          :auto-pass-enabled="autoPassEnabledRef"
          @sort-hand="handleSortHand"
          @play-selection="handlePlaySelection"
          @pass-turn="handlePassTurn"
          @toggle-auto-pass="handleToggleAutoPass"
        />
      </div>
    </div> <!-- End Game UI Wrapper -->

    <!-- Modals are outside the gameUiVisibleRef wrapper -->
    <GameOverModal 
      :is-visible="gameOverModalVisibleRef"
      :winner-name="winnerNameRef"
      :scores="scoresRef"
      :players-info="playersInfoRef"
      :is-match-over="isMatchOverRef"
      :overall-winner-name="overallWinnerNameRef"
      :round-scores-history="roundScoresHistoryRef"
      @new-game="handleNewGame"
    />

    <RulesModal 
      :is-visible="rulesModalVisibleRef"
      @close-rules="closeRulesModal"
    />

    <AliasModal 
      :is-visible="aliasModalVisibleRef"
      :error-message="aliasErrorRef"
      @submit-alias="handleAliasSubmit"
    />

  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, watch, ref } from 'vue';
import { useWebSocket } from '@/composables/useWebSocket';
import { Card, GameStateMessage, ChatMessage, ErrorMessage, ServerMessage, PlayerInfo, PlayedHand, SystemMessage } from '@/types';
import PlayerHand from '@/components/PlayerHand.vue';
import PlayersArea from '@/components/PlayersArea.vue';
import TableDisplay from '@/components/TableDisplay.vue';
import MessagingArea from '@/components/MessagingArea.vue';
import PlayerControls from '@/components/PlayerControls.vue';
import ScoreTable from '@/components/ScoreTable.vue';
import AliasModal from '@/components/AliasModal.vue';
import RulesModal from '@/components/RulesModal.vue';
import GameOverModal from '@/components/GameOverModal.vue';

// Define interfaces for message entries if not already in types.ts
interface ChatMessageEntry {
  sender: string;
  content: string;
  // id?: string; // Optional: if you need a unique key from server
}

interface SystemMessageEntry {
  content: string;
  isError: boolean;
  type: 'general' | 'action';
  // id?: string; // Optional: if you need a unique key
}

// Define type aliases for complex types from GameStateMessage
type Scores = { readonly [playerId: string]: number };
type RoundResult = { readonly [playerId: string]: number };

// Import UI functions and NEW SETTER functions from main.ts
import {
    updatePassButtonAppearance,
    setPlayersInfo,
    setYourPlayerId,
    setCurrentPlayerId,
    setIsMyTurn as setMainTsIsMyTurn,
    setRoundNumber as setMainTsRoundNumber,
    setTargetScore as setMainTsTargetScore,
    setIsMatchOver as setMainTsIsMatchOver,
    setScores as setMainTsScores,
    setRoundScoresHistory as setMainTsRoundScoresHistory,
    setLastPlayedHand,
    setPassCount as setMainTsPassCount,
    setIsGameOver as setMainTsIsGameOver,
    setAutoPassEnabled as setMainTsAutoPassEnabled,
    setCurrentPlayerHand,
    setPreviousCurrentPlayerId,
    gCurrentSortPreference
} from './main';

export default defineComponent({
  name: 'App',
  components: {
    PlayerHand,
    PlayersArea,
    TableDisplay,
    MessagingArea,
    PlayerControls,
    ScoreTable,
    AliasModal,
    RulesModal,
    GameOverModal,
  },
  setup() {
    const { 
      isConnected,
      error: wsError,
      lastMessage,
      connect,
      sendMessage,
    } = useWebSocket();

    // Game State Refs
    const yourPlayerIdRef = ref<string | null>(null);
    const currentPlayerHandRef = ref<readonly Card[]>([]);
    const isGameOverRef = ref<boolean>(false);
    const isMatchOverRef = ref<boolean>(false);
    const selectedCardsForAction = ref<Card[]>([]);
    const playersInfoRef = ref<readonly PlayerInfo[]>([]);
    const lastPlayedHandRef = ref<PlayedHand | null>(null);
    const currentPlayerNameRef = ref<string | null>(null);
    const currentPlayerIdRef = ref<string | null>(null); 
    const passCountRef = ref<number>(0);
    const isMyTurnRef = ref<boolean>(false);
    const autoPassEnabledRef = ref<boolean>(false);

    // ScoreTable Refs
    const scoresRef = ref<Readonly<Scores>>({});
    const roundScoresHistoryRef = ref<RoundResult[]>([]);
    const roundNumberRef = ref<number>(1);
    const targetScoreRef = ref<number>(100);

    // MessagingArea Refs
    const chatMessagesRef = ref<ChatMessageEntry[]>([]);
    const systemMessagesRef = ref<SystemMessageEntry[]>([]);

    // AliasModal Refs
    const aliasModalVisibleRef = ref<boolean>(true);
    const aliasErrorRef = ref<string | null>(null);

    // RulesModal Ref
    const rulesModalVisibleRef = ref<boolean>(false);

    // GameOverModal Refs
    const gameOverModalVisibleRef = ref<boolean>(false);
    const winnerNameRef = ref<string | null>(null);
    const overallWinnerNameRef = ref<string | null>(null);

    // UI Visibility Ref
    const gameUiVisibleRef = ref<boolean>(false);

    const handleSelectedCardsUpdate = (newSelectedCards: Card[]) => {
      selectedCardsForAction.value = newSelectedCards;
    };

    // Event Handlers for PlayerControls
    const handleSortHand = (payload: { preference: 'rank' | 'suit' }) => {
      const hand = [...currentPlayerHandRef.value];
      if (payload.preference === 'suit') {
        hand.sort((a, b) => {
            if (a.suit !== b.suit) {
                return a.suit - b.suit;
            }
            return a.rank - b.rank;
        });
      } else {
        hand.sort((a, b) => a.rank - b.rank);
      }
      currentPlayerHandRef.value = hand;
      localStorage.setItem('sortPreference', payload.preference);
    };

    const handlePlaySelection = () => {
      if (selectedCardsForAction.value.length > 0) {
        sendMessage({ type: 'playCards', cards: selectedCardsForAction.value });
      } else {
        systemMessagesRef.value.push({ content: 'No cards selected to play.', isError: true, type: 'action' });
      }
    };

    const handlePassTurn = () => {
      sendMessage({ type: 'passTurn' });
    };

    const handleToggleAutoPass = () => {
      const newAutoPassState = !autoPassEnabledRef.value;
      autoPassEnabledRef.value = newAutoPassState;
      setMainTsAutoPassEnabled(newAutoPassState);
      updatePassButtonAppearance();
    };

    // Event Handler for MessagingArea
    const handleSendChatMessage = (messageContent: string) => {
      sendMessage({ type: 'chat', content: messageContent });
    };

    // AliasModal Handler
    const handleAliasSubmit = (alias: string) => {
      if (alias && alias.trim().length > 0) {
        sendMessage({ type: 'setAlias', alias: alias.trim() });
      } else {
        aliasErrorRef.value = 'Alias cannot be empty.';
      }
    };

    // RulesModal Handlers
    const openRulesModal = () => {
      rulesModalVisibleRef.value = true;
    };
    const closeRulesModal = () => {
      rulesModalVisibleRef.value = false;
    };

    // GameOverModal Handler
    const handleNewGame = () => {
      sendMessage({ type: 'newGame' });
      // Reset relevant UI state immediately
      gameOverModalVisibleRef.value = false;
      winnerNameRef.value = null;
      isGameOverRef.value = false;
      setMainTsIsGameOver(false);
    };

    watch(lastMessage, (newMessage: ServerMessage | null) => {
      if (!newMessage) {
        return;
      }

      switch (newMessage.type) {
        case 'gameState':
          {
            const gameState = newMessage as GameStateMessage;
            console.log('gameState', gameState);
            const players = gameState.playersInfo || [];
            setPlayersInfo(players);
            playersInfoRef.value = players;

            setYourPlayerId(gameState.yourPlayerId);
            yourPlayerIdRef.value = gameState.yourPlayerId;

            setRoundNumber(gameState.roundNumber || 1);
            setTargetScore(gameState.targetScore || 100);
            setIsMatchOver(gameState.isMatchOver || false);

            if (gameState.yourPlayerId) {
              const me = players.find(p => p.id === gameState.yourPlayerId);
              if (me) {
                setCurrentPlayerHand(gameState.hand || []);
                currentPlayerHandRef.value = gameState.hand || [];
              }
              if (!gameUiVisibleRef.value && aliasModalVisibleRef.value) {
                aliasModalVisibleRef.value = false;
                gameUiVisibleRef.value = true;
              }
            }
            
            setLastPlayedHand(gameState.lastPlayedHand);
            lastPlayedHandRef.value = gameState.lastPlayedHand;
            
            setCurrentPlayerId(gameState.currentPlayerId);
            setPreviousCurrentPlayerId(currentPlayerIdRef.value);
            currentPlayerIdRef.value = gameState.currentPlayerId;
            const currentPlayer = players.find(p => p.id === gameState.currentPlayerId);
            currentPlayerNameRef.value = currentPlayer ? currentPlayer.name : 'Unknown';

            setPassCount(gameState.passCount);
            setScores(gameState.scores || {});
            setRoundScoresHistory(gameState.roundScoresHistory || []);

            setIsGameOver(gameState.isGameOver);
            isGameOverRef.value = gameState.isGameOver;

            setIsMyTurn(yourPlayerIdRef.value === currentPlayerIdRef.value);

            if (gameState.isGameOver) {
              const winner = players.find(p => p.id === gameState.winnerId);
              winnerNameRef.value = winner ? winner.name : 'Unknown';
              const overallWinnerInfo = players.find(p => p.id === gameState.overallWinnerId);
              overallWinnerNameRef.value = overallWinnerInfo ? overallWinnerInfo.name : null;
              gameOverModalVisibleRef.value = true;
            } else {
              gameOverModalVisibleRef.value = false;
            }

            // Update refs that are not directly used in main.ts logic but needed for components
            scoresRef.value = gameState.scores || {};
            roundScoresHistoryRef.value = (gameState.roundScoresHistory || []).map(item => ({ ...item }));
            roundNumberRef.value = gameState.roundNumber || 1;
            targetScoreRef.value = gameState.targetScore || 100;
            isMatchOverRef.value = gameState.isMatchOver || false;
            passCountRef.value = gameState.passCount;
          }
          break;
        case 'chat':
          {
            const chatMessage = newMessage as ChatMessage;
            chatMessagesRef.value.push({ sender: chatMessage.sender, content: chatMessage.content });
          }
          break;
        case 'systemMessage':
          {
            const systemMessage = newMessage as SystemMessage;
            systemMessagesRef.value.push({ content: systemMessage.content, isError: false, type: 'general' });
          }
          break;
        case 'error':
          {
            const error = newMessage as ErrorMessage;
            if (error.context === 'alias') {
              aliasErrorRef.value = error.content;
            } else {
              systemMessagesRef.value.push({ content: `Error: ${error.content}`, isError: true, type: 'action' });
            }
          }
          break;
        case 'actionSuccess':
          selectedCardsForAction.value = [];
          if (document.querySelector('.player-hand-vue')) {
              const handElement = document.querySelector('.player-hand-vue');
              const cardElements = handElement?.querySelectorAll('.card-vue.selected');
              cardElements?.forEach(card => card.classList.remove('selected'));
          }
          break;
      }
    });

    const setIsMyTurn = (isTurn: boolean) => {
      isMyTurnRef.value = isTurn;
      setMainTsIsMyTurn(isTurn);
    };

    const setPassCount = (count: number) => {
      passCountRef.value = count;
      setMainTsPassCount(count);
    }
    
    const setScores = (newScores: Scores) => {
        scoresRef.value = newScores;
        setMainTsScores(newScores);
    }
    
    const setRoundScoresHistory = (history: RoundResult[]) => {
        roundScoresHistoryRef.value = history;
        setMainTsRoundScoresHistory(history);
    }
    
    const setRoundNumber = (roundNum: number) => {
        roundNumberRef.value = roundNum;
        setMainTsRoundNumber(roundNum);
    }

    const setTargetScore = (score: number) => {
        targetScoreRef.value = score;
        setMainTsTargetScore(score);
    }
    
    const setIsMatchOver = (isMatchOver: boolean) => {
        isMatchOverRef.value = isMatchOver;
        setMainTsIsMatchOver(isMatchOver);
    }

    const setIsGameOver = (isGameOver: boolean) => {
        isGameOverRef.value = isGameOver;
        setMainTsIsGameOver(isGameOver);
    }

    onMounted(() => {
      connect('ws://localhost:8080/ws');
    });

    return {
      isConnected,
      wsError,
      lastMessage,
      yourPlayerIdRef,
      currentPlayerHandRef,
      isGameOverRef,
      isMatchOverRef,
      selectedCardsForAction,
      playersInfoRef,
      lastPlayedHandRef,
      currentPlayerNameRef,
      currentPlayerIdRef,
      passCountRef,
      isMyTurnRef,
      autoPassEnabledRef,
      scoresRef,
      roundScoresHistoryRef,
      roundNumberRef,
      targetScoreRef,
      chatMessagesRef,
      systemMessagesRef,
      handleSelectedCardsUpdate,
      handleSortHand,
      handlePlaySelection,
      handlePassTurn,
      handleToggleAutoPass,
      handleSendChatMessage,
      aliasModalVisibleRef,
      aliasErrorRef,
      handleAliasSubmit,
      rulesModalVisibleRef,
      openRulesModal,
      closeRulesModal,
      gameOverModalVisibleRef,
      winnerNameRef,
      overallWinnerNameRef,
      handleNewGame,
      gameUiVisibleRef, // Expose for v-if in template
    };
  }
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
  padding: 15px;
}

.game-layout-wrapper {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.score-and-rules-wrapper {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  gap: 20px;
}

.rules-button-vue {
  margin-top: 20px; /* Adjust as needed */
  height: fit-content;
}
</style> 