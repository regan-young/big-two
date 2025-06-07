import { ref, shallowRef, readonly, type Ref } from 'vue';
import type { ServerMessage } from '@/types'; // Assuming your types are in src/types.ts and you have path alias

// Type guard to check if an object is a valid ServerMessage
function isServerMessage(data: any): data is ServerMessage {
  if (data && typeof data === 'object' && typeof data.type === 'string') {
    const validTypes = ["gameState", "chat", "error", "system", "actionSuccess"];
    return validTypes.includes(data.type);
  }
  return false;
}

// Define the shape of the composable's return value
export interface UseWebSocketReturn {
  socket: Readonly<Ref<WebSocket | null>>;
  isConnected: Readonly<Ref<boolean>>;
  error: Readonly<Ref<any | null>>;
  lastMessage: Readonly<Ref<ServerMessage | null>>; // For now, will just store the last message
  connect: (url: string) => void;
  disconnect: () => void;
  sendMessage: (data: object) => boolean; // Returns true if message was sent, false otherwise
}

export function useWebSocket(): UseWebSocketReturn {
  const socket = shallowRef<WebSocket | null>(null); // Use shallowRef for non-deep reactivity on the socket object itself
  const isConnected = ref(false);
  const error = ref<any | null>(null);
  const lastMessage = ref<ServerMessage | null>(null);

  const connect = (url: string) => {
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected.');
      return;
    }

    // Clean up any existing socket before creating a new one
    if (socket.value) {
        socket.value.close();
    }

    console.log(`Attempting to connect WebSocket to ${url}...`);
    const newSocket = new WebSocket(url);

    newSocket.onopen = () => {
      console.log('WebSocket connection established.');
      isConnected.value = true;
      error.value = null;
      socket.value = newSocket; // Assign after successful connection
    };

    newSocket.onmessage = (event: MessageEvent) => {
      console.log('WebSocket message received:', event.data);
      try {
        const parsedData = JSON.parse(event.data as string);
        
        // Use the type guard to validate the message
        if (isServerMessage(parsedData)) {
            lastMessage.value = parsedData;
        } else {
            console.error('Received malformed message:', parsedData);
            error.value = new Error('Received malformed WebSocket message.');
        }
      } catch (e) {
        console.error('Error parsing WebSocket message:', e);
        error.value = e; // Or a more specific error message
      }
    };

    newSocket.onerror = (event: Event) => {
      console.error('WebSocket error:', event);
      error.value = event; // Or a more specific error object/message
      isConnected.value = false;
      // socket.value will be null or the failing socket, no need to set it to null here
      // as onclose will handle final cleanup if it was ever opened.
    };

    newSocket.onclose = (event: CloseEvent) => {
      console.log('WebSocket connection closed.', event.reason);
      isConnected.value = false;
      // socket.value = null; // Clear the socket ref on close
      // It's important to decide if we set socket.value to null here.
      // If connect can be called again, it will create a new one.
      // For now, let's keep the socket instance for potential inspection of close event details,
      // but new connections will replace it.
      if (socket.value === newSocket) { // Only set to null if it's the one we're tracking
        socket.value = null;
      }
    };
    
    // Note: We don't assign to socket.value here immediately.
    // It's assigned onopen. If the connection fails immediately, 
    // newSocket might not be the one we want to keep in socket.value.
    // However, to allow immediate calls to disconnect on a failing attempt,
    // one might assign it here and rely on onclose/onerror to nullify or update.
    // For now, strict assignment onopen.
  };

  const disconnect = () => {
    if (socket.value) {
      console.log('Disconnecting WebSocket...');
      socket.value.close();
      // State updates (isConnected = false, socket = null) will be handled by onclose
    } else {
      console.log('No WebSocket connection to disconnect.');
    }
  };

  const sendMessage = (data: object): boolean => {
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      try {
        socket.value.send(JSON.stringify(data));
        console.log('WebSocket message sent:', data);
        return true;
      } catch (e) {
        console.error('Error sending WebSocket message:', e);
        error.value = e;
        return false;
      }
    } else {
      console.warn('WebSocket not connected. Cannot send message:', data);
      // Optionally set an error or queue the message, for now just log and return false
      if (!socket.value) {
          error.value = new Error("Socket instance is null. Cannot send message.");
      } else {
          error.value = new Error(`Socket not open. ReadyState: ${socket.value.readyState}. Cannot send message.`);
      }
      return false;
    }
  };

  // Expose reactive state and methods
  return {
    socket: readonly(socket), // Expose socket as readonly shallowRef
    isConnected: readonly(isConnected),
    error: readonly(error),
    lastMessage: readonly(lastMessage),
    connect,
    disconnect,
    sendMessage,
  };
} 