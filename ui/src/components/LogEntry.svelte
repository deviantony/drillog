<script>
  export let log;
  export let searchQuery = '';

  const levelColors = {
    DEBUG: 'text-log-debug',
    INFO: 'text-log-info',
    WARN: 'text-log-warn',
    ERROR: 'text-log-error',
  };

  $: levelColor = levelColors[log.level] || 'text-terminal-text';

  // Format timestamp to show only time portion
  $: time = log.time ? new Date(log.time).toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    fractionalSecondDigits: 3
  }) : '';

  // Check if this log matches the search
  $: isMatch = searchQuery && (
    log.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
    Object.values(log.attrs || {}).some(v =>
      String(v).toLowerCase().includes(searchQuery.toLowerCase())
    )
  );

  // Format attributes for display
  $: attrs = Object.entries(log.attrs || {})
    .filter(([k]) => k !== 'duration') // duration shown elsewhere
    .map(([k, v]) => `${k}=${v}`)
    .join(' ');
</script>

<div
  class="flex items-baseline gap-2 py-0.5 pl-4 text-sm hover:bg-terminal-surface/50
         {isMatch ? 'bg-terminal-surface' : ''}"
>
  <!-- Timestamp -->
  <span class="text-terminal-muted text-xs shrink-0">{time}</span>

  <!-- Level indicator -->
  <span class="w-1.5 h-1.5 rounded-full shrink-0 {levelColor.replace('text-', 'bg-')}"></span>

  <!-- Message -->
  <span class="{levelColor} {isMatch ? 'font-medium' : ''}">
    {log.message}
  </span>

  <!-- Attributes -->
  {#if attrs}
    <span class="text-terminal-muted text-xs">{attrs}</span>
  {/if}
</div>
