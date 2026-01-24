/**
 * Утилиты для обработки данных Cursor в Langfuse
 */

export async function readStdin() {
  return new Promise((resolve, reject) => {
    let data = '';
    process.stdin.setEncoding('utf8');
    process.stdin.on('data', (chunk) => { data += chunk; });
    process.stdin.on('end', () => {
      try { resolve(JSON.parse(data)); } 
      catch (e) { reject(new Error(`Ошибка парсинга JSON: ${e.message}`)); }
    });
    process.stdin.on('error', reject);
  });
}

/**
 * Генерирует имя трейса. 
 */
export function generateTraceName(input) {
  const prompt = input.prompt || input.input;
  const hookName = input.hook_event_name;

  if (prompt) {
    const cleaned = prompt.replace(/\s+/g, ' ').trim();
    const maxLength = 60;
    return cleaned.length <= maxLength ? cleaned : cleaned.substring(0, maxLength) + "...";
  }

  // Если текста нет, возвращаем системное имя
  return `[SYSTEM] ${hookName || 'Action'}`;
}

export function getChatTitle(input) {
  return input.chat_title || input.conversation_title || input.metadata?.title || null;
}

export function generateSessionId(workspaceRoots, conversationId, chatTitle = null) {
  const root = workspaceRoots?.[0] || 'unknown-project';
  const projectName = root.split(/[\\/]/).pop() || root;
  const chatName = chatTitle || `chat:${conversationId.substring(0, 8)}`;
  return `${projectName} | ${chatName}`;
}

/**
 * Генерирует теги. ТЕПЕРЬ ЯВНО ЭКСПОРТИРУЕТСЯ.
 */
export function generateTags(hookName, input) {
  const tags = new Set(['cursor']);
  if (input.model) {
    tags.add(input.model.toLowerCase().replace(/[^a-z0-9-]/g, '-'));
  }
  if (hookName && hookName.toLowerCase().includes('tab')) {
    tags.add('tab-feature');
  }
  return Array.from(tags);
}

export function calculateEditStats(edits) {
  if (!edits || !Array.isArray(edits)) return { editCount: 0, linesAdded: 0, linesRemoved: 0 };
  let added = 0, removed = 0;
  edits.forEach(edit => {
    const oldL = (edit.old_string || '').split('\n').length;
    const newL = (edit.new_string || '').split('\n').length;
    if (newL > oldL) added += (newL - oldL);
    else removed += (oldL - newL);
  });
  return { editCount: edits.length, linesAdded: added, linesRemoved: removed };
}

export function getFileExtension(path) {
  if (!path) return 'unknown';
  return path.split('.').pop().toLowerCase();
}

export function formatDuration(ms) {
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`;
}