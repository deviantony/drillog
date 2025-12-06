/**
 * API client for the drillog viewer backend.
 */

/**
 * Fetch the tree structure with all spans.
 * @returns {Promise<{roots: string[], spans: Object}>}
 */
export async function fetchTree() {
  const res = await fetch('/api/tree');
  if (!res.ok) throw new Error(`Failed to fetch tree: ${res.status}`);
  return res.json();
}

/**
 * Fetch log entries for a specific span.
 * @param {string} spanId
 * @returns {Promise<{logs: Array}>}
 */
export async function fetchLogs(spanId) {
  const res = await fetch(`/api/logs?span=${encodeURIComponent(spanId)}`);
  if (!res.ok) throw new Error(`Failed to fetch logs: ${res.status}`);
  return res.json();
}

/**
 * Fetch aggregate statistics.
 * @returns {Promise<{totalSpans: number, totalLogs: number, levels: Object}>}
 */
export async function fetchStats() {
  const res = await fetch('/api/stats');
  if (!res.ok) throw new Error(`Failed to fetch stats: ${res.status}`);
  return res.json();
}

/**
 * Search log entries.
 * @param {string} query
 * @returns {Promise<{matches: Array, total: number}>}
 */
export async function searchLogs(query) {
  const res = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
  if (!res.ok) throw new Error(`Failed to search: ${res.status}`);
  return res.json();
}
