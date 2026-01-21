console.log("AI-Analyzer Background: Service worker started");

/**
 * Типы сообщений для обмена данными
 */
interface AnalyzeTraceMessage {
  type: "ANALYZE_TRACE";
  traceId: string;
  timestamp: string;
}

interface AnalyzeTraceResponse {
  data?: {
    status: string;
    analyzedTraceId: string;
    timestamp: string;
  };
  error?: string;
}

/**
 * Обработчик сообщений от content scripts
 */
chrome.runtime.onMessage.addListener(
  (
    message: AnalyzeTraceMessage,
    sender: chrome.runtime.MessageSender,
    sendResponse: (response: AnalyzeTraceResponse) => void
  ): boolean => {
    console.log("AI-Analyzer Background: Message received", message);
    console.log("AI-Analyzer Background: Sender info", sender);

    // Валидация типа сообщения
    if (message.type === "ANALYZE_TRACE") {
      console.log("AI-Analyzer Background: Processing ANALYZE_TRACE request");
      
      // Извлекаем данные из сообщения
      const { traceId, timestamp } = message;

      // Валидация данных
      if (!traceId) {
        console.error("AI-Analyzer Background: TraceId is missing");
        sendResponse({
          error: "TraceId отсутствует в запросе"
        });
        return false;
      }

      console.log(`AI-Analyzer Background: Analyzing trace: ${traceId}`);
      console.log(`AI-Analyzer Background: Request timestamp: ${timestamp}`);

      // Отправляем запрос к Go backend
      fetch("http://localhost:8080/analyze", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ traceId })
      })
        .then(response => {
          console.log("AI-Analyzer Background: Backend response status:", response.status);
          
          // Сохраняем статус для обработки
          const status = response.status;
          
          return response.json().then(data => ({ status, data }));
        })
        .then(({ status, data }) => {
          console.log("AI-Analyzer Background: Backend response data:", data);
          
          // Обработка ошибок с специальными статусами
          if (status === 429) {
            const retryAfter = data.retryAfter || 10;
            sendResponse({
              error: `⏱️ Слишком много запросов. Пожалуйста, подождите ${retryAfter} секунд и попробуйте снова.`
            });
            return;
          }
          
          if (status === 402) {
            sendResponse({
              error: "💳 Недостаточно кредитов для AI анализа. Пополните баланс на OpenRouter."
            });
            return;
          }
          
          if (status !== 200) {
            throw new Error(data.error || `HTTP ${status}`);
          }
          
          // Отправляем успешный ответ
          const response: AnalyzeTraceResponse = {
            data: {
              status: "Анализ завершён успешно",
              analyzedTraceId: traceId,
              timestamp: new Date().toISOString(),
              ...data.data // Добавляем данные от AI
            }
          };

          console.log("AI-Analyzer Background: Sending response to content script", response);
          sendResponse(response);
        })
        .catch(error => {
          console.error("AI-Analyzer Background: Error calling backend:", error);
          
          // Отправляем ответ с ошибкой
          sendResponse({
            error: `Ошибка связи с backend: ${error.message}`
          });
        });

      // Возвращаем true для асинхронной отправки ответа
      return true;
    }

    // Неизвестный тип сообщения
    console.warn("AI-Analyzer Background: Unknown message type", message.type);
    sendResponse({
      error: `Неизвестный тип сообщения: ${message.type}`
    });
    
    return false;
  }
);

/**
 * Обработчик установки расширения
 */
chrome.runtime.onInstalled.addListener((details) => {
  console.log("AI-Analyzer Background: Extension installed", details);
  
  if (details.reason === "install") {
    console.log("AI-Analyzer Background: First time installation");
  } else if (details.reason === "update") {
    console.log("AI-Analyzer Background: Extension updated");
  }
});

/**
 * Обработчик запуска расширения
 */
chrome.runtime.onStartup.addListener(() => {
  console.log("AI-Analyzer Background: Extension started");
});

console.log("AI-Analyzer Background: Event listeners registered");