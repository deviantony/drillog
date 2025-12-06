<script>
  import TreeNode from './TreeNode.svelte';

  export let tree;
  export let visibleLevels;
  export let searchQuery = '';
  export let searchResults = null;

  // Track expanded state for all spans
  let expanded = {};

  // Initialize all spans as expanded
  $: if (tree?.spans) {
    for (const spanId of Object.keys(tree.spans)) {
      if (expanded[spanId] === undefined) {
        expanded[spanId] = true;
      }
    }
  }

  // Get span IDs that match the search
  $: matchingSpans = new Set(
    searchResults?.matches?.map(m => m.span) || []
  );

  function toggleExpanded(spanId) {
    expanded = { ...expanded, [spanId]: !expanded[spanId] };
  }

  function expandAll() {
    const newExpanded = {};
    for (const spanId of Object.keys(tree.spans)) {
      newExpanded[spanId] = true;
    }
    expanded = newExpanded;
  }

  function collapseAll() {
    const newExpanded = {};
    for (const spanId of Object.keys(tree.spans)) {
      newExpanded[spanId] = false;
    }
    expanded = newExpanded;
  }
</script>

<div class="space-y-1">
  <!-- Expand/collapse controls -->
  <div class="flex items-center gap-2 mb-3 text-xs">
    <button
      on:click={expandAll}
      class="text-terminal-muted hover:text-terminal-text"
    >
      Expand all
    </button>
    <span class="text-terminal-border">|</span>
    <button
      on:click={collapseAll}
      class="text-terminal-muted hover:text-terminal-text"
    >
      Collapse all
    </button>
    {#if searchQuery && searchResults}
      <span class="text-terminal-border">|</span>
      <span class="text-terminal-muted">
        {searchResults.total} match{searchResults.total === 1 ? '' : 'es'}
      </span>
    {/if}
  </div>

  <!-- Tree roots -->
  {#each tree.roots as rootId}
    <TreeNode
      spanId={rootId}
      spans={tree.spans}
      {expanded}
      {visibleLevels}
      {searchQuery}
      {matchingSpans}
      depth={0}
      on:toggle={(e) => toggleExpanded(e.detail.spanId)}
    />
  {/each}
</div>
