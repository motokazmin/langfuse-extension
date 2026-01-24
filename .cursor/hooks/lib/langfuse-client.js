import { Langfuse } from "langfuse";
import { config } from "dotenv";
import { resolve, dirname } from "path";
import { fileURLToPath } from "url";
import { 
  generateTraceName, 
  generateSessionId, 
  generateTags, 
  getChatTitle 
} from "./utils.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const projectRoot = resolve(__dirname, "..", "..", "..");
config({ path: resolve(projectRoot, ".env") });

export const HOOK_HANDLER_VERSION = "1.7.0";

// Список хуков, которые НЕ должны менять имя трейса
const CONTINUATION_HOOKS = [
  'afterAgentThought',
  'afterAgentResponse',
  'afterFileEdit',
  'afterMCPExecution',
  'afterShellExecution',
  'afterTabFileEdit',
  'stop'
];

let langfuseInstance = null;
let isInitializing = false;

// ===== КЭШ ТРЕЙСОВ — предотвращает перезапись имени =====
const traceCache = new Map();

export function getLangfuseClient() {
  if (!langfuseInstance && !isInitializing) {
    isInitializing = true;
    try {
      langfuseInstance = new Langfuse({
        secretKey: process.env.LANGFUSE_SECRET_KEY,
        publicKey: process.env.LANGFUSE_PUBLIC_KEY,
        baseUrl: process.env.LANGFUSE_BASE_URL || "https://cloud.langfuse.com",
        // Оптимальные настройки для hooks:
        flushAt: 15, // отправка каждые 15 событий (был 1 - слишком часто)
        flushInterval: 1000, // или раз в секунду
        requestTimeout: 10000, // увеличен таймаут
      });
    } catch (err) {
      console.error('[Langfuse] Init error:', err.message);
    } finally {
      isInitializing = false;
    }
  }
  return langfuseInstance;
}

export function getOrCreateTrace(input) {
  let langfuse;
  try {
    langfuse = getLangfuseClient();
    if (!langfuse) {
      return createMockTrace();
    }
  } catch (err) {
    console.error('[Langfuse] Error getting client:', err.message);
    return createMockTrace();
  }

  const chatTitle = getChatTitle(input);
  const sessionId = generateSessionId(input.workspace_roots, input.conversation_id, chatTitle);
  const generationId = input.generation_id;

  // Проверяем кэш
  const cached = traceCache.get(generationId);

  const traceParams = {
    id: generationId,
    sessionId: sessionId,
    sessionProperties: chatTitle ? { name: chatTitle } : undefined,
    userId: input.user_email || undefined,
    metadata: {
      cursor_version: input.cursor_version,
      model: input.model,
      hook: input.hook_event_name
    },
    tags: generateTags(input.hook_event_name, input),
  };

  // ===== ЛОГИКА УСТАНОВКИ ИМЕНИ =====
  const hasPrompt = !!(input.prompt || input.input);
  const isContinuation = CONTINUATION_HOOKS.includes(input.hook_event_name);

  let shouldSetName = false;

  if (!cached) {
    // Первое создание трейса
    if (hasPrompt) {
      shouldSetName = true;
    } else if (!isContinuation) {
      shouldSetName = true;
    }
  } else {
    // Трейс уже существует
    if (hasPrompt && !cached.hasSetName) {
      shouldSetName = true;
    }
  }

  // КЛЮЧЕВОЙ МОМЕНТ: передаём name только если нужно установить
  if (shouldSetName) {
    traceParams.name = generateTraceName(input);
  }

  let trace;
  try {
    trace = langfuse.trace(traceParams);
  } catch (err) {
    console.error('[Langfuse] Error creating trace:', err.message);
    return createMockTrace();
  }

  // Сохраняем в кэш
  if (!cached) {
    traceCache.set(generationId, { 
      trace, 
      hasSetName: shouldSetName 
    });
  } else if (shouldSetName) {
    cached.hasSetName = true;
  }

  return trace;
}

// Заглушка для случаев, когда Langfuse недоступен
function createMockTrace() {
  return {
    update: () => {},
    generation: () => ({ end: () => {} }),
    span: () => ({ end: () => {} }),
    score: () => {},
    end: () => {}
  };
}

export function addCompletionScores(trace, input) {
  try {
    const scores = { completed: 1, aborted: 0.5, error: 0 };
    const val = scores[input.status] ?? 0.5;
    trace.score({ 
      name: "completion_status", 
      value: val, 
      comment: `Status: ${input.status}` 
    });
  } catch (err) {
    console.error('[Langfuse] Error adding scores:', err.message);
  }
}

export async function flushLangfuse() {
  const client = getLangfuseClient();
  if (!client) {
    return;
  }
  
  try {
    // Flush с таймаутом 5 секунд (Langfuse может делать batch requests)
    await Promise.race([
      client.flushAsync(),
      new Promise((resolve) => setTimeout(resolve, 5000))
    ]);
  } catch (err) {
    console.error('[Langfuse] Flush error:', err.message);
  }
  
  // Очищаем кэш после flush
  traceCache.clear();
}