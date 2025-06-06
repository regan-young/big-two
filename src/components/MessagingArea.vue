<template>
  <div class="messaging-area-vue">
    <div class="system-messages-container">
      <h3>Game & System Messages</h3>
      <div class="messages-log system-log">
        <p v-if="!systemMessages || systemMessages.length === 0">No system messages.</p>
        <div v-for="(msg, index) in systemMessages" :key="`system-${index}`"
             :class="{ 'error-message': msg.isError, 'action-message': msg.type === 'action', 'general-message': msg.type === 'general' }">
          {{ msg.content }}
        </div>
      </div>
    </div>

    <div class="chat-container">
      <h3>Chat</h3>
      <div class="messages-log chat-log">
        <p v-if="!chatMessages || chatMessages.length === 0">No chat messages.</p>
        <div v-for="(chat, index) in chatMessages" :key="`chat-${index}`" class="chat-entry">
          <strong>{{ chat.sender }}:</strong> {{ chat.content }}
        </div>
      </div>
      <div class="chat-input-controls">
        <input type="text" v-model="chatInputMessage" @keyup.enter="handleSendChatMessage" placeholder="Type message...">
        <button @click="handleSendChatMessage">Send</button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, ref, nextTick, watch } from 'vue';

interface ChatMessageEntry {
  sender: string;
  content: string;
}

interface SystemMessageEntry {
  content: string;
  isError: boolean;
  type: 'general' | 'action'; // To distinguish styling or placement
}

export default defineComponent({
  name: 'MessagingArea',
  props: {
    chatMessages: {
      type: Array as PropType<readonly ChatMessageEntry[]>,
      default: () => [],
    },
    systemMessages: {
      type: Array as PropType<readonly SystemMessageEntry[]>,
      default: () => [],
    },
  },
  emits: ['send-chat-message'],
  setup(props, { emit }) {
    const chatInputMessage = ref('');

    const handleSendChatMessage = () => {
      if (chatInputMessage.value.trim()) {
        emit('send-chat-message', chatInputMessage.value.trim());
        chatInputMessage.value = '';
      }
    };

    // Auto-scroll chat and system logs
    const chatLogRef = ref<HTMLDivElement | null>(null); // For direct DOM manipulation if needed for scrolling
    const systemLogRef = ref<HTMLDivElement | null>(null);

    watch(() => props.chatMessages, async () => {
      await nextTick();
      const chatLogDiv = document.querySelector('.chat-log'); // Simple querySelector for now
      if (chatLogDiv) chatLogDiv.scrollTop = chatLogDiv.scrollHeight;
    }, { deep: true });

    watch(() => props.systemMessages, async () => {
      await nextTick();
      const systemLogDiv = document.querySelector('.system-log'); // Simple querySelector for now
      if (systemLogDiv) systemLogDiv.scrollTop = systemLogDiv.scrollHeight;
    }, { deep: true });


    return {
      chatInputMessage,
      handleSendChatMessage,
      // chatLogRef, // if we use ref binding in template
      // systemLogRef,
    };
  },
});
</script>

<style scoped>
.messaging-area-vue {
  display: flex;
  flex-direction: column;
  gap: 15px;
  padding: 10px;
  background-color: #f8f9fa;
  min-width: 280px; /* Ensure it has some base width */
}

.system-messages-container, .chat-container {
  border: 1px solid #ccc;
  padding: 10px;
  background-color: #fff;
  border-radius: 4px;
}

.messaging-area-vue h3 {
  margin-top: 0;
  font-size: 1.1em;
  color: #333;
  margin-bottom: 8px;
  border-bottom: 1px solid #eee;
  padding-bottom: 5px;
}

.messages-log {
  height: 120px; /* Adjust as needed */
  overflow-y: auto;
  border: 1px solid #e0e0e0;
  padding: 8px;
  background-color: #fdfdfd;
  font-size: 0.9em;
  line-height: 1.4;
}
.messages-log p {
    color: #888;
    font-style: italic;
}

.system-log div {
  margin-bottom: 4px;
  padding: 3px 5px;
  border-radius: 3px;
}
.error-message {
  color: #721c24;
  background-color: #f8d7da;
  border-left: 3px solid #f5c6cb;
}
.action-message { /* Could be same as error or different */
  color: #004085;
  background-color: #cce5ff;
  border-left: 3px solid #b8daff;
}
.general-message {
  color: #383d41;
  background-color: #e2e3e5;
  border-left: 3px solid #d6d8db;
}


.chat-entry {
  margin-bottom: 4px;
  text-align: left;
}
.chat-entry strong {
  color: #0056b3; /* Darker blue for sender name */
}

.chat-input-controls {
  display: flex;
  margin-top: 8px;
}
.chat-input-controls input[type="text"] {
  flex-grow: 1;
  padding: 8px;
  border: 1px solid #ccc;
  border-radius: 4px 0 0 4px;
}
.chat-input-controls button {
  padding: 8px 12px;
  border: 1px solid #007bff;
  background-color: #007bff;
  color: white;
  cursor: pointer;
  border-radius: 0 4px 4px 0;
  border-left: none;
}
.chat-input-controls button:hover {
  background-color: #0056b3;
}
</style> 