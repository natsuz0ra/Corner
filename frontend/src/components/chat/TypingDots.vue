<template>
  <div class="flex items-center gap-1 py-1">
    <span class="typing-dot" />
    <span class="typing-dot" style="animation-delay: 0.15s" />
    <span class="typing-dot" style="animation-delay: 0.30s" />
  </div>
</template>

<style scoped>
.typing-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 9999px;
  background: #9ca3af;
  animation: bounce-dot 1s ease-in-out infinite;
}

@keyframes bounce-dot {
  0%, 80%, 100% {
    transform: translateY(0);
    opacity: 0.5;
  }
  40% {
    transform: translateY(-5px);
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .typing-dot { animation: none; opacity: 0.7; }
}
</style>
