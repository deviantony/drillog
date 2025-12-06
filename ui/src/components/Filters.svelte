<script>
  import { createEventDispatcher } from 'svelte';

  export let visibleLevels;
  export let levels = {};

  const dispatch = createEventDispatcher();

  const levelConfig = [
    { key: 'DEBUG', color: 'bg-log-debug', label: 'DBG' },
    { key: 'INFO', color: 'bg-log-info', label: 'INF' },
    { key: 'WARN', color: 'bg-log-warn', label: 'WRN' },
    { key: 'ERROR', color: 'bg-log-error', label: 'ERR' },
  ];

  function toggle(level) {
    visibleLevels = { ...visibleLevels, [level]: !visibleLevels[level] };
    dispatch('change', { levels: visibleLevels });
  }
</script>

<div class="flex items-center gap-1">
  {#each levelConfig as { key, color, label }}
    {@const count = levels?.[key] || 0}
    {@const active = visibleLevels[key]}
    <button
      on:click={() => toggle(key)}
      class="flex items-center gap-1.5 px-2 py-1 rounded text-xs font-medium transition-opacity
             {active ? 'opacity-100' : 'opacity-40'}"
      title="{key}: {count} entries"
    >
      <span class="w-2 h-2 rounded-full {color}"></span>
      <span>{label}</span>
      {#if count > 0}
        <span class="text-terminal-muted">({count})</span>
      {/if}
    </button>
  {/each}
</div>
