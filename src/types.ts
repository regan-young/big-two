// Interfaces for game objects
export interface Card {
    rank: number;
    suit: number;
}

export interface PlayerInfo {
    id: string;
    name: string;
    cardCount: number;
    hasPassed: boolean;
}

export interface PlayedHand {
    cards: Card[];
    playerId: string;
    handType: number;
    handTypeString: string;
    rank: number;
    EffectiveRank: number;
    EffectiveSuit: number;
}

// Server Message Interfaces
export interface GameStateMessage {
    type: "gameState";
    hand?: Card[] | null; 
    lastPlayedHand: PlayedHand | null;
    yourPlayerId: string | null;
    currentPlayerId: string | null;
    currentPlayerName?: string;
    passCount: number;
    playersInfo?: PlayerInfo[];
    isGameOver: boolean;
    scores?: { [playerId: string]: number };
    roundNumber?: number;
    targetScore?: number;
    isMatchOver?: boolean;
    overallWinnerId?: string | null;
    roundScoresHistory?: { [playerId: string]: number }[];
    winnerId: string | null; 
    gameMessage?: string;
}

export interface ChatMessage {
    type: "chat";
    sender: string;
    content: string;
}

export interface ErrorMessage {
    type: "error";
    content: string;
}

export type ServerMessage = GameStateMessage | ChatMessage | ErrorMessage; 