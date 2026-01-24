import { calculateEditStats, formatDuration } from "./utils.js";
import { addCompletionScores } from "./langfuse-client.js";

export function handleBeforeSubmitPrompt(trace, input) {
  // Здесь мы записываем входные данные
  trace.update({ input: input.prompt });
  
  trace.generation({
    name: "User Prompt",
    input: input.prompt,
    model: input.model
  });
  return { continue: true };
}

export function handleAfterAgentResponse(trace, input) {
  trace.update({ output: input.text });
  return null;
}

export function handleAfterAgentThought(trace, input) {
  trace.span({
    name: "Thinking",
    output: input.text,
    metadata: { duration: formatDuration(input.duration_ms) }
  }).end();
  return null;
}

export function handleAfterFileEdit(trace, input) {
  const stats = calculateEditStats(input.edits);
  trace.span({
    name: `File Edit: ${input.file_path?.split('/').pop()}`,
    input: { file: input.file_path },
    output: stats
  }).end();
  return null;
}

export function routeHookHandler(hookName, trace, input) {
  const handlers = {
    beforeSubmitPrompt: handleBeforeSubmitPrompt,
    afterAgentResponse: handleAfterAgentResponse,
    afterAgentThought: handleAfterAgentThought,
    afterFileEdit: handleAfterFileEdit,
    stop: (t, i) => { 
      addCompletionScores(t, i); 
      return {}; // Возвращаем пустой объект для Cursor
    }
  };
  const h = handlers[hookName];
  return h ? h(trace, input) : null;
}