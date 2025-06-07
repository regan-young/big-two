import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import './assets/styles.css';
import svgSprite from './assets/cards.svg?raw';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.mount('#app');

document.addEventListener('DOMContentLoaded', () => {
    const container = document.getElementById('svg-sprite-container');
    if (container) {
        container.innerHTML = svgSprite;
    }
});