<script>
  import { onMount } from 'svelte';
  import { fetchTree, fetchStats, searchLogs } from './lib/api.js';
  import Stats from './components/Stats.svelte';
  import Search from './components/Search.svelte';
  import Filters from './components/Filters.svelte';
  import Tree from './components/Tree.svelte';

  let tree = null;
  let stats = null;
  let loading = true;
  let error = null;

  // Filter state
  let visibleLevels = { DEBUG: true, INFO: true, WARN: true, ERROR: true };
  let searchQuery = '';
  let searchResults = null;

  onMount(async () => {
    try {
      [tree, stats] = await Promise.all([fetchTree(), fetchStats()]);
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  async function handleSearch(event) {
    searchQuery = event.detail.query;
    if (searchQuery.trim()) {
      try {
        searchResults = await searchLogs(searchQuery);
      } catch (e) {
        console.error('Search failed:', e);
        searchResults = null;
      }
    } else {
      searchResults = null;
    }
  }

  function handleFilterChange(event) {
    visibleLevels = event.detail.levels;
  }
</script>

<div class="min-h-screen flex flex-col">
  <!-- Header -->
  <header class="bg-terminal-surface border-b border-terminal-border px-4 py-3">
    <div class="flex items-center justify-between">
      <h1 class="text-lg font-semibold">
        <span class="text-terminal-muted">drillog</span> viewer
      </h1>
      {#if stats}
        <Stats {stats} />
      {/if}
    </div>
  </header>

  <!-- Toolbar -->
  <div class="bg-terminal-surface border-b border-terminal-border px-4 py-2 flex items-center gap-4">
    <Search on:search={handleSearch} />
    <Filters {visibleLevels} levels={stats?.levels} on:change={handleFilterChange} />
  </div>

  <!-- Main content -->
  <main class="flex-1 overflow-auto p-4">
    {#if loading}
      <div class="flex items-center justify-center h-64">
        <div class="text-terminal-muted">Loading...</div>
      </div>
    {:else if error}
      <div class="flex items-center justify-center h-64">
        <div class="text-log-error">Error: {error}</div>
      </div>
    {:else if tree}
      <Tree
        {tree}
        {visibleLevels}
        {searchQuery}
        {searchResults}
      />
    {/if}
  </main>
</div>
