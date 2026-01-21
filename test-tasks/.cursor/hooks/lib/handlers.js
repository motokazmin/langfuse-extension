/**
 * Hook Handlers Module
 * Contains handlers for all Cursor hook events.
 */

import { calculateEditStats, formatDuration } from "./utils.js";
import { addCompletionScores } from "./langfuse-client.js";

/**
 * Handle beforeSubmitPrompt hook
 * Records the user's prompt as input
 */
export function handleBeforeSubmitPrompt(trace, input) {
  trace.update({ input: input.prompt });
  
  trace.generation({
    name: "User Prompt",
    input: input.prompt,
    model: input.model
  });
  
  return { continue: true };
}

/**
 * Handle afterAgentResponse hook
 * Records the agent's response as output
 */
export function handleAfterAgentResponse(trace, input) {
  trace.update({ output: input.text });
  return null;
}

/**
 * Handle afterAgentThought hook
 * Records agent thinking process
 */
export function handleAfterAgentThought(trace, input) {
  trace.span({
    name: "Thinking",
    output: input.text,
    metadata: { duration: formatDuration(input.duration_ms) }
  }).end();
  
  return null;
}

/**
 * Handle afterFileEdit hook
 * Records file edits with statistics
 */
export function handleAfterFileEdit(trace, input) {
  const stats = calculateEditStats(input.edits);
  
  trace.span({
    name: `File Edit: ${input.file_path?.split('/').pop()}`,
    input: { file: input.file_path },
    output: stats
  }).end();
  
  return null;
}

/**
 * Route hook events to appropriate handlers
 * @param {string} hookName - Name of the hook event
 * @param {object} trace - Langfuse trace object
 * @param {object} input - Hook input data
 * @returns {object|null} Response object for Cursor or null
 */
export function routeHookHandler(hookName, trace, input) {
  const handlers = {
    beforeSubmitPrompt: handleBeforeSubmitPrompt,
    afterAgentResponse: handleAfterAgentResponse,
    afterAgentThought: handleAfterAgentThought,
    afterFileEdit: handleAfterFileEdit,
    stop: (t, i) => { 
      addCompletionScores(t, i); 
      return {}; // Return empty object for Cursor
    }
  };
  
  const handler = handlers[hookName];
  return handler ? handler(trace, input) : null;
}