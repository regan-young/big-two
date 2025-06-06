// Interfaces for game objects
export interface Card {
    readonly rank: number;
    readonly suit: number;
}

export interface PlayerInfo {
    readonly id: string;
    readonly name: string;
    readonly cardCount: number;
    readonly hasPassed: boolean;
}

export interface PlayedHand {
    readonly cards: readonly Card[];
    readonly playerId: string;
    readonly handType: number;
    readonly handTypeString: string;
    readonly rank: number;
    readonly EffectiveRank: number;
    readonly EffectiveSuit: number;
}

// Server Message Interfaces
export interface GameStateMessage {
    readonly type: "gameState";
    readonly hand?: readonly Card[] | null;
    readonly lastPlayedHand: PlayedHand | null;
    readonly yourPlayerId: string | null;
    readonly currentPlayerId: string | null;
    readonly currentPlayerName?: string;
    readonly passCount: number;
    readonly playersInfo?: readonly PlayerInfo[];
    readonly isGameOver: boolean;
    readonly scores?: { readonly [playerId: string]: number };
    readonly roundNumber?: number;
    readonly targetScore?: number;
    readonly isMatchOver?: boolean;
    readonly overallWinnerId?: string | null;
    readonly roundScoresHistory?: { [playerId: string]: number }[];
    readonly winnerId: string | null; 
    readonly gameMessage?: string;
}

export interface ChatMessage {
    readonly type: "chat";
    readonly sender: string;
    readonly content: string;
}

export interface SystemMessage {
    readonly type: "systemMessage";
    readonly content: string;
}

export interface ErrorMessage {
    readonly type: "error";
    readonly content: string;
    readonly context?: string;
}

export interface ActionSuccessMessage {
    readonly type: "actionSuccess";
}

export type ServerMessage = GameStateMessage | ChatMessage | ErrorMessage | SystemMessage | ActionSuccessMessage; 