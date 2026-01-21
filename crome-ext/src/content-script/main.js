console.log("AI-Analyzer: Main script loading...");

// Ждём пока React будет доступен в глобальном скопе
// (он загружается injector.ts перед этим скриптом)
if (typeof window.React === 'undefined') {
  console.error("AI-Analyzer: React not found in window");
  // Повторяем попытку через 500мс
  setTimeout(() => {
    if (typeof window.React !== 'undefined') {
      loadApp();
    } else {
      console.error("AI-Analyzer: React still not available after 500ms");
    }
  }, 500);
} else {
  loadApp();
}

function loadApp() {
  const React = window.React;
  const ReactDOM = window.ReactDOM;

  console.log("AI-Analyzer: React and ReactDOM available, loading app");

  // Встраиваем стили напрямую
  const styles = `
    .ai-analyzer-container {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      margin: 0;
      padding: 0.75rem;
      border: 2px solid #6d28d9;
      border-radius: 8px;
      background-color: #f9f9f9;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .analyze-button {
      background-color: #6d28d9;
      color: white;
      border: none;
      padding: 8px 16px;
      border-radius: 6px;
      font-size: 14px;
      font-weight: 500;
      cursor: pointer;
      transition: background-color 0.2s;
      white-space: nowrap;
    }

    .analyze-button:hover {
      background-color: #5b21b6;
    }

    .chat-window {
      margin-top: 0.5rem;
      padding: 1rem;
      border: 1px solid #ccc;
      border-radius: 6px;
      background-color: white;
      color: #333;
      font-size: 12px;
      max-height: 200px;
      overflow-y: auto;
    }

    .chat-window p {
      margin: 0;
    }
  `;

  // Shadow DOM компонент
  const ShadowDomWrapper = ({ children }) => {
    const hostRef = React.useRef(null);
    const [shadowRoot, setShadowRoot] = React.useState(null);

    React.useEffect(() => {
      if (hostRef.current && !shadowRoot) {
        const newShadowRoot = hostRef.current.attachShadow({ mode: 'open' });
        const styleElement = document.createElement('style');
        styleElement.textContent = styles;
        newShadowRoot.appendChild(styleElement);
        setShadowRoot(newShadowRoot);
        console.log("AI-Analyzer: Shadow DOM created");
      }
    }, [shadowRoot]);

    return React.createElement(
      'div',
      { ref: hostRef },
      shadowRoot && ReactDOM.createPortal(children, shadowRoot)
    );
  };

  // App компонент
  const App = () => {
    const [isChatOpen, setChatOpen] = React.useState(false);

    const handleAnalyzeClick = () => {
      console.log("🤖 AI-Анализ кнопка нажата!");
      setChatOpen(!isChatOpen);
    };

    return React.createElement(
      ShadowDomWrapper,
      null,
      React.createElement(
        'div',
        { className: 'ai-analyzer-container' },
        React.createElement(
          'button',
          { className: 'analyze-button', onClick: handleAnalyzeClick },
          isChatOpen ? '✨ Закрыть' : '🤖 AI-Анализ'
        ),
        isChatOpen && React.createElement(
          'div',
          { className: 'chat-window' },
          React.createElement('p', null, 'Окно анализа... Скоро здесь будут результаты.')
        )
      )
    );
  };

  // Монтируем приложение
  const rootElement = document.getElementById('ai-analyzer-react-root');

  if (rootElement) {
    console.log("AI-Analyzer: Root element found, mounting React app...");
    const root = ReactDOM.createRoot(rootElement);
    root.render(
      React.createElement(
        React.StrictMode,
        null,
        React.createElement(App)
      )
    );
    console.log("AI-Analyzer: React app mounted successfully!");
  } else {
    console.error("AI-Analyzer: Root element not found!");
  }
}
