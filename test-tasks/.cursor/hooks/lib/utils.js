/**
 * Utility functions for Cursor Langfuse hooks
 */

/**
 * Read and parse JSON input from stdin
 * Cursor hooks pass data via stdin as JSON
 * @returns {Promise<object>} Parsed JSON object from stdin
 */
export async function readStdin() {
  return new Promise((resolve, reject) => {
    let data = '';
    process.stdin.setEncoding('utf8');
    process.stdin.on('data', (chunk) => {
      data += chunk;
    });
    process.stdin.on('end', () => {
      try {
        resolve(JSON.parse(data));
      } catch (e) {
        reject(new Error(`Failed to parse JSON from stdin: ${e.message}`));
      }
    });
    process.stdin.on('error', reject);
  });
}

/**
 * Generate a descriptive trace name from the input
 * @param {object} input - The input data from the hook
 * @returns {string} A descriptive trace name
 */
export function generateTraceName(input) {
  const prompt = input.prompt || input.input;
  const hookName = input.hook_event_name;

  if (prompt) {
    // Extract first meaningful words from the prompt (max 60 chars)
    const cleaned = prompt
      .replace(/\s+/g, ' ')
      .trim();
    
    const maxLength = 60;
    return cleaned.length <= maxLength 
      ? cleaned 
      : cleaned.substring(0, maxLength) + '...';
  }

  // If no text available, return system name
  return `[SYSTEM] ${hookName || 'Action'}`;
}

/**
 * Extract chat title from input data
 * @param {object} input - The input data from the hook
 * @returns {string|null} The chat title or null
 */
export function getChatTitle(input) {
  return input.chat_title || input.conversation_title || input.metadata?.title || null;
}

/**
 * Generate a session ID from workspace roots and conversation
 * Groups all conversations in the same workspace together
 * @param {string[]} workspaceRoots - Array of workspace root paths
 * @param {string} conversationId - The conversation ID
 * @param {string} chatTitle - Optional chat title
 * @returns {string} Session ID
 */
export function generateSessionId(workspaceRoots, conversationId, chatTitle = null) {
  const root = workspaceRoots?.[0] || 'unknown-project';
  const projectName = root.split(/[\\/]/).pop() || root;
  const chatName = chatTitle || `chat:${conversationId.substring(0, 8)}`;
  return `${projectName} | ${chatName}`;
}

/**
 * Generate dynamic tags based on hook activity
 * @param {string} hookName - The name of the hook being executed
 * @param {object} input - The input data from the hook
 * @returns {string[]} Array of tags
 */
export function generateTags(hookName, input) {
  const tags = new Set(['cursor']);
  
  // Add model-specific tag
  if (input.model) {
    const modelTag = input.model
      .toLowerCase()
      .replace(/[^a-z0-9-]/g, '-');
    tags.add(modelTag);
  }
  
  // Add tab feature tag if applicable
  if (hookName && hookName.toLowerCase().includes('tab')) {
    tags.add('tab-feature');
  }
  
  return Array.from(tags);
}

/**
 * Calculate edit statistics from an array of edits
 * @param {Array<{old_string: string, new_string: string}>} edits - Array of edits
 * @returns {object} Edit statistics
 */
export function calculateEditStats(edits) {
  if (!edits || !Array.isArray(edits)) {
    return { editCount: 0, linesAdded: 0, linesRemoved: 0 };
  }
  
  let added = 0;
  let removed = 0;
  
  edits.forEach(edit => {
    const oldLines = (edit.old_string || '').split('\n').length;
    const newLines = (edit.new_string || '').split('\n').length;
    
    if (newLines > oldLines) {
      added += (newLines - oldLines);
    } else if (oldLines > newLines) {
      removed += (oldLines - newLines);
    }
  });
  
  return {
    editCount: edits.length,
    linesAdded: added,
    linesRemoved: removed,
  };
}

/**
 * Extract file extension from a file path
 * @param {string} filePath - The file path
 * @returns {string} The file extension (without dot) or 'unknown'
 */
export function getFileExtension(filePath) {
  if (!filePath) return 'unknown';
  return filePath.split('.').pop().toLowerCase();
}

/**
 * Format duration in milliseconds to a human-readable string
 * @param {number} ms - Duration in milliseconds
 * @returns {string} Formatted duration
 */
export function formatDuration(ms) {
  if (!ms || ms < 0) return '0ms';
  
  return ms < 1000 
    ? `${ms}ms` 
    : `${(ms / 1000).toFixed(1)}s`;
}