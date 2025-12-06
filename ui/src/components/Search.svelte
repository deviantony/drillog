<script>
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let query = '';
  let debounceTimer;

  function handleInput() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      dispatch('search', { query });
    }, 200);
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      query = '';
      dispatch('search', { query: '' });
    }
  }
</script>

<div class="relative flex-1 max-w-md">
  <input
    type="text"
    bind:value={query}
    on:input={handleInput}
    on:keydown={handleKeydown}
    placeholder="Search logs... (Esc to clear)"
    class="w-full bg-terminal-bg border border-terminal-border rounded px-3 py-1.5 text-sm
           placeholder-terminal-muted focus:outline-none focus:border-terminal-muted"
  />
  {#if query}
    <button
      on:click={() => { query = ''; dispatch('search', { query: '' }); }}
      class="absolute right-2 top-1/2 -translate-y-1/2 text-terminal-muted hover:text-terminal-text"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  {/if}
</div>
