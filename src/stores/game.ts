import { defineStore } from 'pinia';
import { Card, PlayerInfo as ServerPlayerInfo, PlayedHand, SystemMessage, ChatMessage, GameStateMessage, ErrorMessage } from '@/types';

// Extend the server's PlayerInfo for our client-side needs
export interface Player extends ServerPlayerInfo {
  hand: Card[];
}

// Define the shape of your game state
export interface GameState {
  players: Player[];
  yourPlayerId: string | null;
  currentPlayerId: string | null;
  lastPlayedHand: PlayedHand | null;
  gamePhase: string; // e.g., 'selecting-cards', 'waiting-for-turn', 'ended'
  scores: Record<string, number>;
  roundScoresHistory: readonly Record<string, number>[];
  roundNumber: number;
  targetScore: number;
  passCount: number;
  isGameOver: boolean;
  isMatchOver: boolean;
  systemMessages: SystemMessage[];
  errorMessages: ErrorMessage[];
  validationError: string | null;
  chatMessages: ChatMessage[];
  autoPassEnabled: boolean;
  sortPreference: 'rank' | 'suit';
}

export const useGameStore = defineStore('game', {
  state: (): GameState => ({
    // Initialize your state here
    players: [],
    yourPlayerId: null,
    currentPlayerId: null,
    lastPlayedHand: null,
    gamePhase: 'loading', // Initial phase
    scores: {},
    roundScoresHistory: [],
    roundNumber: 1,
    targetScore: 100,
    passCount: 0,
    isGameOver: false,
    isMatchOver: false,
    systemMessages: [],
    errorMessages: [],
    validationError: null,
    chatMessages: [],
    autoPassEnabled: false,
    sortPreference: 'rank',
  }),

  getters: {
    // get a specific player by ID
    getPlayerById: (state) => (id: string): Player | undefined => {
      return state.players.find(p => p.id === id);
    },
    
    // get the current player
    getCurrentPlayer: (state): Player | undefined => {
      if (!state.currentPlayerId) return undefined;
      return state.players.find(p => p.id === state.currentPlayerId);
    }
  },

  actions: {
    // Action to process incoming WebSocket messages
    processWebSocketMessage(message: any) {
        this.validationError = null; // Clear previous validation error on any new message
        switch (message.type) {
            case 'gameState':
                this.handleGameStateUpdate(message);
                break;
            case 'error':
                if (message.context === 'validation') {
                    this.validationError = message.content;
                } else {
                    this.errorMessages.push(message);
                }
                break;
            case 'chat':
                this.chatMessages.push({ type: 'chat', sender: message.sender, content: message.content });
                break;
            case 'system':
                 this.systemMessages.push({ type: 'systemMessage', content: message.content });
                break;
            // Add other message types as needed
        }
    },

    handleGameStateUpdate(state: GameStateMessage) {
        // Update yourPlayerId first, as it's needed to correctly identify the player's hand.
        this.yourPlayerId = state.yourPlayerId ?? this.yourPlayerId;

        // Update players, preserving their hands if they exist
        if (state.playersInfo) {
            this.players = state.playersInfo.map(serverPlayer => {
                const existingPlayer = this.players.find(p => p.id === serverPlayer.id);
                return { ...serverPlayer, hand: existingPlayer?.hand || [] };
            });
        }
        
        // Update your own hand if the server sends it
        const me = this.players.find(p => p.id === this.yourPlayerId);
        if (me && state.hand) {
            me.hand = [...state.hand]; // Ensure reactivity
            this.sortHand(this.sortPreference); // Re-apply sorting
        }

        // Update other game state properties, providing defaults for optional ones
        this.currentPlayerId = state.currentPlayerId ?? this.currentPlayerId;
        this.lastPlayedHand = state.lastPlayedHand;
        this.passCount = state.passCount ?? this.passCount;
        this.isGameOver = state.isGameOver ?? this.isGameOver;
        this.isMatchOver = state.isMatchOver ?? this.isMatchOver;
        this.roundNumber = state.roundNumber ?? this.roundNumber;
        this.scores = state.scores ?? this.scores;
        this.roundScoresHistory = state.roundScoresHistory ?? this.roundScoresHistory;
        this.targetScore = state.targetScore ?? this.targetScore;

        if (this.gamePhase === 'loading' && this.yourPlayerId) {
            this.gamePhase = 'playing';
        }
        if (this.isGameOver) {
            this.gamePhase = 'ended';
        }
    },

    sortHand(preference: 'rank' | 'suit') {
        const player = this.players.find(p => p.id === this.yourPlayerId);
        if (!player?.hand) return;

        const hand = [...player.hand];
        if (preference === 'suit') {
            hand.sort((a, b) => {
                if (a.suit !== b.suit) return a.suit - b.suit;
                return a.rank - b.rank;
            });
        } else { // 'rank'
            hand.sort((a, b) => a.rank - b.rank);
        }
        
        player.hand = hand;
        this.sortPreference = preference;
    },

    toggleAutoPass() {
        this.autoPassEnabled = !this.autoPassEnabled;
    },

    clearValidationError() {
        this.validationError = null;
    }
  },
}); 