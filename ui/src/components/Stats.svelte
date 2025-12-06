<script>
  export let stats;
</script>

<div class="flex items-center gap-4 text-sm">
  <div class="flex items-center gap-1.5">
    <span class="text-terminal-muted">Spans:</span>
    <span class="font-medium">{stats.totalSpans.toLocaleString()}</span>
  </div>
  <div class="flex items-center gap-1.5">
    <span class="text-terminal-muted">Logs:</span>
    <span class="font-medium">{stats.totalLogs.toLocaleString()}</span>
  </div>
  {#if stats.levels.ERROR > 0}
    <div class="flex items-center gap-1.5">
      <span class="w-2 h-2 rounded-full bg-log-error"></span>
      <span class="text-log-error font-medium">{stats.levels.ERROR} errors</span>
    </div>
  {/if}
</div>
