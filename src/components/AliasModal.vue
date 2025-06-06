<template>
  <div v-if="isVisible" class="modal-backdrop alias-modal-vue">
    <div class="modal-content">
      <h2>Welcome! Please Enter Your Name</h2>
      <input type="text" v-model="aliasInputValue" @keyup.enter="handleSubmit" placeholder="Your Name">
      <button @click="handleSubmit">Join Game</button>
      <p v-if="errorMessage" class="alias-error">{{ errorMessage }}</p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, watch } from 'vue';

export default defineComponent({
  name: 'AliasModal',
  props: {
    isVisible: {
      type: Boolean,
      required: true,
    },
    errorMessage: {
      type: String as () => string | null,
      default: null,
    },
  },
  emits: ['submit-alias'],
  setup(props, { emit }) {
    const aliasInputValue = ref('');

    const handleSubmit = () => {
      if (aliasInputValue.value.trim()) {
        emit('submit-alias', aliasInputValue.value.trim());
      }
      // Do not clear input or hide modal here; parent (App.vue) will control visibility
      // and can clear error/input if submission is successful.
    };

    // Optional: Clear input when modal becomes visible after being hidden
    // or if an error is cleared.
    watch(() => props.isVisible, (newValue) => {
      if (newValue) {
        // Potentially clear input or error if needed when re-shown,
        // but often we want to preserve input unless explicitly cleared by parent.
        // aliasInputValue.value = ''; // Example if you want to reset input on show
      }
    });

     watch(() => props.errorMessage, (newError) => {
        // If error is cleared by parent, maybe also clear input? Or let user retry.
        // if (!newError) aliasInputValue.value = '';
    });


    return {
      aliasInputValue,
      handleSubmit,
    };
  },
});
</script>

<style scoped>
/* Styles are similar to original modal styles in style.css */
/* Ensure these don't conflict or override if style.css is still loaded globally */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000; /* Ensure it's above other content */
}

.modal-content {
  background-color: #fff;
  padding: 25px 30px;
  border-radius: 8px;
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
  text-align: center;
  min-width: 300px;
  max-width: 400px;
}

.modal-content h2 {
  margin-top: 0;
  margin-bottom: 20px;
  font-size: 1.5em;
  color: #333;
}

.modal-content input[type="text"] {
  width: calc(100% - 22px); /* Account for padding/border */
  padding: 10px;
  margin-bottom: 15px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 1em;
}

.modal-content button {
  padding: 10px 20px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1em;
  transition: background-color 0.2s;
}

.modal-content button:hover {
  background-color: #0056b3;
}

.alias-error {
  color: #dc3545; /* Red for errors */
  margin-top: 10px;
  font-size: 0.9em;
  min-height: 1.2em; /* Prevent layout shift */
}
</style> 