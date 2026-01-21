/**
 * Langfuse Client Module
 *
 * Handles Langfuse SDK initialization and trace management
 * with support for sessions, scoring, and dynamic metadata.
 */

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

// Load .env from project root (3 levels up from lib/)
const projectRoot = resolve(__dirname, "..", "..", "..");
config({ path: resolve(projectRoot, ".env") });

export const HOOK_HANDLER_VERSION = "1.7.0";

/**
 * Hooks that should not modify the trace name
 * These are continuation hooks that occur after the initial request
 */
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

/**
 * Trace cache - prevents name overwrites
 * Maps generation_id to trace instance and metadata
 */
const traceCache = new Map();

/**
 * Get or initialize the Langfuse client
 * @returns {Langfuse} Langfuse client instance
 */
export function getLangfuseClient() {
  if (!langfuseInstance && !isInitializing) {
    isInitializing = true;
    try {
      langfuseInstance = new Langfuse({
        secretKey: process.env.LANGFUSE_SECRET_KEY,
        publicKey: process.env.LANGFUSE_PUBLIC_KEY,
        baseUrl: process.env.LANGFUSE_BASE_URL || "https://cloud.langfuse.com",
        // Optimal settings for hooks:
        flushAt: 15, // send every 15 events (was 1 - too frequent)
        flushInterval: 1000, // or once per second
        requestTimeout: 10000, // increased timeout
      });
    } catch (err) {
      console.error('[Langfuse] Init error:', err.message);
    } finally {
      isInitializing = false;
    }
  }
  return langfuseInstance;
}

/**
 * Get or create a trace for the current conversation
 * Implements caching to prevent trace name overwrites
 * @param {object} input - Hook input data
 * @returns {object} Langfuse trace object
 */
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

  // Check cache
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

  // ===== NAME SETTING LOGIC =====
  const hasPrompt = !!(input.prompt || input.input);
  const isContinuation = CONTINUATION_HOOKS.includes(input.hook_event_name);

  let shouldSetName = false;

  if (!cached) {
    // First trace creation
    if (hasPrompt) {
      shouldSetName = true;
    } else if (!isContinuation) {
      shouldSetName = true;
    }
  } else {
    // Trace already exists
    if (hasPrompt && !cached.hasSetName) {
      shouldSetName = true;
    }
  }

  // KEY POINT: only pass name if we need to set it
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

  // Save to cache
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

/**
 * Create a mock trace object for fallback when Langfuse is unavailable
 * @returns {object} Mock trace with no-op methods
 */
function createMockTrace() {
  return {
    update: () => {},
    generation: () => ({ end: () => {} }),
    span: () => ({ end: () => {} }),
    score: () => {},
    end: () => {}
  };
}

/**
 * Add completion status scores to a trace
 * @param {object} trace - Langfuse trace object
 * @param {object} input - Hook input data with status
 */
export function addCompletionScores(trace, input) {
  try {
    const scores = { completed: 1, aborted: 0.5, error: 0 };
    const value = scores[input.status] ?? 0.5;
    trace.score({ 
      name: "completion_status", 
      value: value, 
      comment: `Status: ${input.status}` 
    });
  } catch (err) {
    console.error('[Langfuse] Error adding scores:', err.message);
  }
}

/**
 * Flush all pending events to Langfuse
 * Includes timeout and cache cleanup
 */
export async function flushLangfuse() {
  const client = getLangfuseClient();
  if (!client) {
    return;
  }
  
  try {
    // Flush with 5 second timeout (Langfuse may make batch requests)
    await Promise.race([
      client.flushAsync(),
      new Promise((resolve) => setTimeout(resolve, 5000))
    ]);
  } catch (err) {
    console.error('[Langfuse] Flush error:', err.message);
  }
  
  // Clear cache after flush
  traceCache.clear();
}