#!/usr/bin/env node

/**
 * Cursor Hooks Langfuse Integration
 * 
 * Main entry point for Cursor hooks that sends traces to Langfuse.
 * 
 * Features:
 * - All 12 Cursor hooks supported (Agent + Tab)
 * - Traces grouped by conversation_id
 * - Sessions grouped by workspace
 * - Dynamic tags based on activity
 * - Completion scores and efficiency metrics
 * - Rich metadata and edit statistics
 * 
 * @version 1.1.0
 * @see https://cursor.com/docs/agent/hooks
 * @see https://langfuse.com/docs
 */

import { readStdin } from './lib/utils.js';
import { 
  getOrCreateTrace, 
  flushLangfuse,
  HOOK_HANDLER_VERSION,
} from './lib/langfuse-client.js';
import { routeHookHandler } from './lib/handlers.js';

/**
 * Main handler function
 * Reads hook data from stdin, creates Langfuse trace, and routes to handler
 */
async function main() {
  let input;
  
  try {
    // Read JSON input from stdin
    input = await readStdin();
    console.error(`[Hook] Processing: ${input.hook_event_name}, gen_id: ${input.generation_id?.substring(0, 8)}`);
  } catch (error) {
    // Log error but don't crash - we don't want to block Cursor
    console.error(`[Langfuse Hook v${HOOK_HANDLER_VERSION}] Error reading stdin: ${error.message}`);
    console.log(JSON.stringify({ continue: true }));
    return;
  }

  // ========================================
  // CRITICAL: Respond to Cursor immediately
  // ========================================
  // This ensures the hook doesn't block operations if something goes wrong
  let hookResponse = { continue: true };
  
  try {
    // Get or create a trace for this conversation
    const trace = getOrCreateTrace(input);
    
    if (trace) {
      // Route to the appropriate handler based on hook type
      const response = routeHookHandler(input.hook_event_name, trace, input);
      if (response) {
        hookResponse = response;
      }
    }
  } catch (error) {
    console.error(`[Langfuse Hook v${HOOK_HANDLER_VERSION}] Error in handler: ${error.message}`);
  }

  // Output response to Cursor immediately
  console.log(JSON.stringify(hookResponse));
  console.error(`[Hook] Response sent, starting flush...`);

  // ========================================
  // Flush all pending events to Langfuse AFTER responding
  // ========================================
  // Now we can wait for flush since response is already sent
  flushLangfuse()
    .then(() => {
      console.error(`[Hook] Flush completed successfully`);
    })
    .catch(err => {
      console.error(`[Langfuse Hook v${HOOK_HANDLER_VERSION}] Flush error: ${err.message}`);
    })
    .finally(() => {
      // Allow 200ms for network buffering, then exit
      setTimeout(() => {
        console.error(`[Hook] Exiting...`);
        process.exit(0);
      }, 200);
    });
}

// Run the main function
main();