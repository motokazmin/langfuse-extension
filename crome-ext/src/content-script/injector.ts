console.log("AI-Analyzer: Injector script running.");
console.log("Current URL:", location.href);
console.log("Current pathname:", location.pathname);

const APP_ROOT_ID = 'ai-analyzer-react-root';
let lastUrl = location.href;
let cachedTraceId: string | null = null; // Кэш для trace ID

/**
 * Извлекает traceId из URL страницы Langfuse
 * Поддерживает форматы:
 * - .../traces/TRACE_ID
 * - .../traces?peek=TRACE_ID
 */
const extractTraceId = (): string | null => {
  const url = window.location.href;
  console.log("AI-Analyzer: Extracting traceId from URL:", url);

  try {
    // Сначала пробуем URLSearchParams (более надежный способ)
    const urlObj = new URL(url);
    console.log("AI-Analyzer: Parsed URL object:", {
      pathname: urlObj.pathname,
      search: urlObj.search,
      searchParams: Array.from(urlObj.searchParams.entries())
    });
    
    // Вариант 1: traceId в query параметре peek (.../traces?peek=TRACE_ID)
    const peekParam = urlObj.searchParams.get('peek');
    console.log("AI-Analyzer: Peek param value:", peekParam);
    
    if (peekParam && peekParam.trim() !== '') {
      const traceId = peekParam.trim();
      console.log("AI-Analyzer: TraceId found in peek param:", traceId);
      cachedTraceId = traceId; // Сохраняем в кэш
      return traceId;
    }

    // Вариант 2: traceId в пути URL (.../traces/TRACE_ID)
    const pathMatch = url.match(/\/traces\/([a-zA-Z0-9_-]+)/);
    if (pathMatch && pathMatch[1]) {
      const traceId = pathMatch[1];
      console.log("AI-Analyzer: TraceId found in path:", traceId);
      cachedTraceId = traceId; // Сохраняем в кэш
      return traceId;
    }

    // Вариант 3: используем кэшированное значение если URL изменился
    if (cachedTraceId) {
      console.log("AI-Analyzer: Using cached traceId:", cachedTraceId);
      return cachedTraceId;
    }

    console.warn("AI-Analyzer: TraceId not found in URL");
    console.warn("AI-Analyzer: URL pathname:", urlObj.pathname);
    console.warn("AI-Analyzer: URL search params:", urlObj.search);
    return null;
    
  } catch (error) {
    console.error("AI-Analyzer: Error extracting traceId:", error);
    return null;
  }
};

/**
 * Показывает прогресс анализа
 */
const showProgressIndicator = (): { update: (step: string) => void; remove: () => void } => {
  // Удаляем предыдущий индикатор если есть
  const existing = document.getElementById('ai-analyzer-progress');
  if (existing) existing.remove();

  const progressDiv = document.createElement('div');
  progressDiv.id = 'ai-analyzer-progress';
  progressDiv.style.position = 'fixed';
  progressDiv.style.top = '160px';
  progressDiv.style.right = '20px';
  progressDiv.style.zIndex = '9999';
  progressDiv.style.width = '250px';
  progressDiv.style.backgroundColor = '#ffffff';
  progressDiv.style.border = '2px solid #6d28d9';
  progressDiv.style.borderRadius = '8px';
  progressDiv.style.padding = '16px';
  progressDiv.style.boxShadow = '0 4px 12px rgba(0,0,0,0.15)';
  progressDiv.style.fontFamily = 'system-ui, -apple-system, sans-serif';

  progressDiv.innerHTML = `
    <div style="display: flex; align-items: center; gap: 12px; margin-bottom: 12px;">
      <div class="spinner" style="
        width: 20px;
        height: 20px;
        border: 3px solid #e5e7eb;
        border-top-color: #6d28d9;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
      "></div>
      <div style="font-weight: 600; color: #1f2937; font-size: 14px;">Анализ трейса...</div>
    </div>
    <div id="progress-step" style="font-size: 12px; color: #6b7280; line-height: 1.5;"></div>
    <style>
      @keyframes spin {
        to { transform: rotate(360deg); }
      }
    </style>
  `;

  document.body.appendChild(progressDiv);

  return {
    update: (step: string) => {
      const stepDiv = document.getElementById('progress-step');
      if (stepDiv) stepDiv.textContent = step;
    },
    remove: () => {
      progressDiv.remove();
    }
  };
};

/**
 * Отображает результаты анализа в красивом модальном окне
 */
const displayAnalysisResults = (response: any, traceId: string): void => {
  // Удаляем предыдущее модальное окно если есть
  const existingModal = document.getElementById('ai-analyzer-modal');
  if (existingModal) {
    existingModal.remove();
  }

  // Создаём модальное окно
  const modal = document.createElement('div');
  modal.id = 'ai-analyzer-modal';
  modal.style.position = 'fixed';
  modal.style.top = '0';
  modal.style.left = '0';
  modal.style.width = '100%';
  modal.style.height = '100%';
  modal.style.backgroundColor = 'rgba(0, 0, 0, 0.5)';
  modal.style.zIndex = '10000';
  modal.style.display = 'flex';
  modal.style.alignItems = 'center';
  modal.style.justifyContent = 'center';

  // Создаём контент модального окна
  const modalContent = document.createElement('div');
  modalContent.style.backgroundColor = 'white';
  modalContent.style.borderRadius = '12px';
  modalContent.style.padding = '24px';
  modalContent.style.maxWidth = '600px';
  modalContent.style.maxHeight = '80vh';
  modalContent.style.overflow = 'auto';
  modalContent.style.boxShadow = '0 4px 20px rgba(0,0,0,0.3)';

  // Извлекаем данные анализа
  const data = response.data || response;
  const analysisSummary = data.analysisSummary || {};
  const detailedAnalysis = data.detailedAnalysis || {};
  
  // Определяем цвет статуса
  const statusColors: Record<string, string> = {
    'HEALTHY': '#10b981',
    'WARNING': '#f59e0b',
    'ERROR': '#ef4444'
  };
  const statusColor = statusColors[analysisSummary.overallStatus] || '#6b7280';

  // Создаём HTML содержимое
  modalContent.innerHTML = `
    <div style="margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h2 style="margin: 0; font-size: 24px; color: #1f2937;">🤖 AI Анализ Трейса</h2>
        <button id="ai-analyzer-close" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #6b7280;">×</button>
      </div>
      
      <div style="background-color: #f3f4f6; padding: 12px; border-radius: 8px; margin-bottom: 16px;">
        <div style="font-size: 12px; color: #6b7280; margin-bottom: 4px;">Trace ID</div>
        <div style="font-family: monospace; font-size: 14px; color: #1f2937; word-break: break-all;">${traceId}</div>
      </div>
    </div>

    <div style="margin-bottom: 20px;">
      <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 12px;">
        <div style="width: 12px; height: 12px; border-radius: 50%; background-color: ${statusColor};"></div>
        <h3 style="margin: 0; font-size: 18px; color: #1f2937;">Статус: ${analysisSummary.overallStatus || 'N/A'}</h3>
      </div>
      <p style="margin: 0; color: #4b5563; line-height: 1.5;">${analysisSummary.keyFinding || 'Анализ завершён'}</p>
    </div>

    ${detailedAnalysis.anomalyType && detailedAnalysis.anomalyType !== 'NONE' ? `
      <div style="border-top: 1px solid #e5e7eb; padding-top: 20px;">
        <h3 style="margin: 0 0 12px 0; font-size: 16px; color: #1f2937;">
          ⚠️ Обнаружена аномалия: <span style="color: #ef4444;">${detailedAnalysis.anomalyType}</span>
        </h3>
        
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; color: #374151; margin-bottom: 4px;">Описание:</div>
          <p style="margin: 0; color: #4b5563; line-height: 1.5;">${detailedAnalysis.description || 'Нет описания'}</p>
        </div>

        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; color: #374151; margin-bottom: 4px;">Первопричина:</div>
          <p style="margin: 0; color: #4b5563; line-height: 1.5;">${detailedAnalysis.rootCause || 'Не определена'}</p>
        </div>

        <div style="background-color: #ecfdf5; padding: 12px; border-radius: 8px; border-left: 4px solid #10b981;">
          <div style="font-weight: 600; color: #065f46; margin-bottom: 4px;">💡 Рекомендация:</div>
          <p style="margin: 0; color: #047857; line-height: 1.5;">${detailedAnalysis.recommendation || 'Нет рекомендаций'}</p>
        </div>
      </div>
    ` : ''}

    <div style="margin-top: 24px; display: flex; justify-content: flex-end;">
      <button id="ai-analyzer-ok" style="
        background-color: #6d28d9;
        color: white;
        border: none;
        padding: 10px 24px;
        border-radius: 6px;
        font-size: 14px;
        font-weight: 500;
        cursor: pointer;
        transition: background-color 0.2s;
      ">Закрыть</button>
    </div>
  `;

  modal.appendChild(modalContent);
  document.body.appendChild(modal);

  // Обработчики закрытия
  const closeModal = () => modal.remove();
  document.getElementById('ai-analyzer-close')?.addEventListener('click', closeModal);
  document.getElementById('ai-analyzer-ok')?.addEventListener('click', closeModal);
  modal.addEventListener('click', (e) => {
    if (e.target === modal) closeModal();
  });

  // Hover эффект для кнопки
  const okButton = document.getElementById('ai-analyzer-ok');
  if (okButton) {
    okButton.addEventListener('mouseenter', () => {
      okButton.style.backgroundColor = '#5b21b6';
    });
    okButton.addEventListener('mouseleave', () => {
      okButton.style.backgroundColor = '#6d28d9';
    });
  }
};

/**
 * Отправляет запрос на анализ трейса в фоновый скрипт
 */
const sendAnalyzeRequest = async (traceId: string): Promise<void> => {
  console.log("AI-Analyzer: Sending analyze request for traceId:", traceId);

  // Показываем прогресс-индикатор
  const progress = showProgressIndicator();
  progress.update('🔄 Отправка запроса на сервер...');

  try {
    // Формируем структурированное сообщение
    const message = {
      type: "ANALYZE_TRACE",
      traceId: traceId,
      timestamp: new Date().toISOString()
    };

    console.log("AI-Analyzer: Message payload:", message);
    
    // Обновляем прогресс
    setTimeout(() => progress.update('📡 Получение данных из Langfuse...'), 500);
    setTimeout(() => progress.update('🤖 AI анализирует трейс...'), 2000);

    // Отправляем сообщение в background script
    chrome.runtime.sendMessage(message, (response) => {
      // Убираем прогресс-индикатор
      progress.remove();

      // Проверяем наличие ошибок Chrome API
      if (chrome.runtime.lastError) {
        console.error("AI-Analyzer: Chrome runtime error:", chrome.runtime.lastError.message);
        alert(`❌ Ошибка связи с расширением: ${chrome.runtime.lastError.message}`);
        return;
      }

      // Обрабатываем ответ от background script
      if (response) {
        console.log("AI-Analyzer: Received response from background:", response);
        
        if (response.data) {
          // Показываем красивое модальное окно с результатами
          displayAnalysisResults(response, traceId);
          console.log("AI-Analyzer: Analysis completed successfully");
        } else if (response.error) {
          console.error("AI-Analyzer: Error from background:", response.error);
          alert(`❌ Ошибка: ${response.error}`);
        }
      } else {
        console.warn("AI-Analyzer: No response received from background");
        alert("⚠️ Нет ответа от фонового скрипта");
      }
    });

  } catch (error) {
    console.error("AI-Analyzer: Exception in sendAnalyzeRequest:", error);
    alert(`❌ Исключение: ${error instanceof Error ? error.message : String(error)}`);
  }
};

/**
 * Обработчик клика на кнопку AI-Анализа
 */
const handleAnalyzeClick = async (): Promise<void> => {
  console.log("AI-Analyzer: Analyze button clicked!");
  console.log("AI-Analyzer: Current location.href:", location.href);
  console.log("AI-Analyzer: Current location.pathname:", location.pathname);
  console.log("AI-Analyzer: Current location.search:", location.search);

  // Показываем индикатор загрузки
  const button = document.querySelector(`#${APP_ROOT_ID} button`) as HTMLButtonElement;
  if (button) {
    const originalText = button.textContent;
    button.textContent = '⏳ Анализ...';
    button.disabled = true;

    try {
      // Получаем traceId из data-атрибута (сохранённый при создании кнопки)
      const appRoot = document.getElementById(APP_ROOT_ID);
      const traceId = appRoot?.dataset.traceId;
      
      console.log("AI-Analyzer: Stored trace ID from button:", traceId);

      if (!traceId) {
        console.error("AI-Analyzer: No stored traceId in button!");
        alert("❌ Не удалось определить Trace ID.\n\nПопробуйте обновить страницу.");
        return;
      }

      console.log("AI-Analyzer: Using stored TraceId:", traceId);
      // Отправляем запрос
      await sendAnalyzeRequest(traceId);

    } catch (error) {
      console.error("AI-Analyzer: Error in handleAnalyzeClick:", error);
      alert(`❌ Произошла ошибка: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      // Восстанавливаем кнопку
      if (button) {
        button.textContent = originalText;
        button.disabled = false;
      }
    }
  }
};

/**
 * Функция для внедрения UI приложения
 */
const tryInjectApp = (): void => {
  console.log("AI-Analyzer: tryInjectApp called. Current pathname:", location.pathname);
  
  // Проверяем, что мы на странице трейса
  if (location.pathname.includes('/traces')) {
    console.log("AI-Analyzer: Trace page detected!");
    
    // Проверяем есть ли trace ID (в URL или в кэше)
    const currentTraceId = extractTraceId();
    const hasTraceId = currentTraceId !== null;
    console.log("AI-Analyzer: Has trace ID:", hasTraceId, currentTraceId);
    
    const existingRoot = document.getElementById(APP_ROOT_ID);
    
    // Если нет trace ID, удаляем кнопку если она была
    if (!hasTraceId) {
      if (existingRoot) {
        console.log("AI-Analyzer: No trace ID, removing button");
        existingRoot.remove();
      }
      return;
    }
    
    // Если нашего UI еще нет, создаем его
    if (!existingRoot) {
      console.log("AI-Analyzer: App root not found, injecting...");
      
      const appRoot = document.createElement('div');
      appRoot.id = APP_ROOT_ID;
      
      // ВАЖНО: Сохраняем trace ID в data-атрибуте при создании
      appRoot.dataset.traceId = currentTraceId;
      console.log("AI-Analyzer: Stored trace ID in button:", currentTraceId);
      
      // Стили для позиционирования
      appRoot.style.position = 'fixed';
      appRoot.style.top = '100px';
      appRoot.style.right = '20px';
      appRoot.style.zIndex = '9999';
      appRoot.style.width = 'auto';
      appRoot.style.height = 'auto';
      appRoot.style.minWidth = '250px';
      appRoot.style.backgroundColor = '#f9f9f9';
      appRoot.style.border = '2px solid #6d28d9';
      appRoot.style.borderRadius = '8px';
      appRoot.style.padding = '12px';
      appRoot.style.boxShadow = '0 2px 8px rgba(0,0,0,0.1)';
      
      // Внедряем контейнер в body
      document.body.appendChild(appRoot);
      console.log("AI-Analyzer: Container added to DOM");

      // Создаём кнопку
      const button = document.createElement('button');
      button.textContent = '🤖 AI-Анализ';
      button.style.backgroundColor = '#6d28d9';
      button.style.color = 'white';
      button.style.border = 'none';
      button.style.padding = '8px 16px';
      button.style.borderRadius = '6px';
      button.style.fontSize = '14px';
      button.style.fontWeight = '500';
      button.style.cursor = 'pointer';
      button.style.transition = 'background-color 0.2s';
      button.style.width = '100%';
      
      // Эффекты при наведении
      button.onmouseenter = () => {
        if (!button.disabled) {
          button.style.backgroundColor = '#5b21b6';
        }
      };
      button.onmouseleave = () => {
        if (!button.disabled) {
          button.style.backgroundColor = '#6d28d9';
        }
      };
      
      // Обработчик клика с новой логикой
      button.onclick = handleAnalyzeClick;
      
      appRoot.appendChild(button);
      console.log("AI-Analyzer: Button added with message exchange logic");
      
    } else {
      console.log("AI-Analyzer: App root already exists, skipping injection");
    }
  } else {
    console.log("AI-Analyzer: Not on trace page");
    
    // Если мы не на странице трейса, удаляем наш UI
    const existingRoot = document.getElementById(APP_ROOT_ID);
    if (existingRoot) {
      existingRoot.remove();
      console.log("AI-Analyzer: Left trace page, removing app container.");
    }
  }
};

// Слушатель для отслеживания навигации внутри SPA
const initMutationObserver = () => {
  if (!document.body) {
    // body ещё не готов, пробуем позже
    setTimeout(initMutationObserver, 50);
    return;
  }
  
  new MutationObserver(() => {
    if (location.href !== lastUrl) {
      lastUrl = location.href;
      console.log("AI-Analyzer: URL changed to:", lastUrl);
      
      // Обновляем кэш trace ID при изменении URL
      extractTraceId();
      
      setTimeout(tryInjectApp, 100);
    }
  }).observe(document.body, { childList: true, subtree: true });
  
  console.log("AI-Analyzer: MutationObserver initialized");
};

// Инициализируем observer
initMutationObserver();

// Первая попытка внедрения при загрузке скрипта
console.log("AI-Analyzer: Initial injection attempt");
setTimeout(tryInjectApp, 100);

// Дополнительная попытка через 1 секунду
setTimeout(() => {
  console.log("AI-Analyzer: Secondary injection attempt");
  tryInjectApp();
}, 1000);