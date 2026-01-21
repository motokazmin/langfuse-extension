#!/usr/bin/env node

import { readStdin } from './lib/utils.js';
import { getOrCreateTrace, flushLangfuse, HOOK_HANDLER_VERSION } from './lib/langfuse-client.js';
import { routeHookHandler } from './lib/handlers.js';

async function main() {
  let input;
  
  try {
    input = await readStdin();
    console.error(`[Hook] Processing: ${input.hook_event_name}, gen_id: ${input.generation_id?.substring(0, 8)}`);
  } catch (error) {
    console.error(`[Langfuse Hook v${HOOK_HANDLER_VERSION}] Error reading stdin: ${error.message}`);
    console.log(JSON.stringify({ continue: true }));
    return;
  }

  // ========================================
  // КРИТИЧНО: СРАЗУ отвечаем Cursor
  // ========================================
  let hookResponse = { continue: true };
  
  try {
    const trace = getOrCreateTrace(input);
    
    if (trace) {
      const response = routeHookHandler(input.hook_event_name, trace, input);
      if (response) {
        hookResponse = response;
      }
    }
  } catch (error) {
    console.error(`[Langfuse Hook v${HOOK_HANDLER_VERSION}] Error in handler: ${error.message}`);
  }

  // НЕМЕДЛЕННО возвращаем ответ в Cursor
  console.log(JSON.stringify(hookResponse));
  console.error(`[Hook] Response sent, starting flush...`);

  // ========================================
  // Langfuse flush идёт ПОСЛЕ ответа
  // ========================================
  // Теперь можем подождать flush, т.к. ответ уже отправлен
  flushLangfuse()
    .then(() => {
      console.error(`[Hook] Flush completed successfully`);
    })
    .catch(err => {
      console.error(`[Langfuse Hook v${HOOK_HANDLER_VERSION}] Flush error: ${err.message}`);
    })
    .finally(() => {
      // Даём ещё 200ms на network буферизацию, потом выходим
      setTimeout(() => {
        console.error(`[Hook] Exiting...`);
        process.exit(0);
      }, 200);
    });
}

main();