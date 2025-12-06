<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { fetchLogs } from '../lib/api.js';
  import LogEntry from './LogEntry.svelte';

  export let spanId;
  export let spans;
  export let expanded;
  export let visibleLevels;
  export let searchQuery;
  export let matchingSpans;
  export let depth = 0;

  const dispatch = createEventDispatcher();

  let logs = null;
  let logsLoaded = false;

  $: span = spans[spanId];
  $: isExpanded = expanded[spanId];
  $: hasChildren = span?.children?.length > 0;
  $: hasMatch = matchingSpans.has(spanId);

  // Load logs when expanded for the first time
  $: if (isExpanded && !logsLoaded) {
    loadLogs();
  }

  async function loadLogs() {
    try {
      const result = await fetchLogs(spanId);
      logs = result.logs;
      logsLoaded = true;
    } catch (e) {
      console.error(`Failed to load logs for ${spanId}:`, e);
    }
  }

  function toggle() {
    dispatch('toggle', { spanId });
  }

  // Filter logs by visible levels
  $: filteredLogs = logs?.filter(log => visibleLevels[log.level]) || [];
</script>

<div class="select-none">
  <!-- Span header -->
  <div
    class="flex items-center gap-1 py-0.5 rounded hover:bg-terminal-surface cursor-pointer group
           {hasMatch ? 'bg-terminal-surface' : ''}"
    style="padding-left: {depth * 20}px"
    on:click={toggle}
    on:keydown={(e) => e.key === 'Enter' && toggle()}
    role="button"
    tabindex="0"
  >
    <!-- Expand/collapse indicator -->
    <span class="w-4 h-4 flex items-center justify-center text-terminal-muted">
      {#if hasChildren || (logs && logs.length > 0)}
        <svg
          class="w-3 h-3 transition-transform {isExpanded ? 'rotate-90' : ''}"
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path d="M6 6L14 10L6 14V6Z" />
        </svg>
      {:else}
        <span class="w-3"></span>
      {/if}
    </span>

    <!-- Span name -->
    <span class="font-medium {hasMatch ? 'text-log-warn' : 'text-terminal-text'}">
      {span?.name || spanId}
    </span>

    <!-- Duration badge -->
    {#if span?.duration}
      <span class="text-xs text-terminal-muted bg-terminal-bg px-1.5 py-0.5 rounded">
        {span.duration}
      </span>
    {/if}

    <!-- Log count -->
    <span class="text-xs text-terminal-muted opacity-0 group-hover:opacity-100">
      {span?.logCount || 0} logs
    </span>

    <!-- Span ID (dimmed) -->
    <span class="text-xs text-terminal-border opacity-0 group-hover:opacity-100 ml-auto mr-2">
      {spanId}
    </span>
  </div>

  <!-- Expanded content -->
  {#if isExpanded}
    <div class="ml-4 border-l border-terminal-border" style="margin-left: {depth * 20 + 8}px">
      <!-- Log entries for this span -->
      {#if filteredLogs.length > 0}
        <div class="py-1">
          {#each filteredLogs as log}
            <LogEntry {log} {searchQuery} />
          {/each}
        </div>
      {/if}

      <!-- Child spans (recursive) -->
      {#if hasChildren}
        {#each span.children as childId}
          <svelte:self
            spanId={childId}
            {spans}
            {expanded}
            {visibleLevels}
            {searchQuery}
            {matchingSpans}
            depth={depth + 1}
            on:toggle
          />
        {/each}
      {/if}
    </div>
  {/if}
</div>
