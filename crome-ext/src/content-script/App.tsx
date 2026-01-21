import React, { useRef, useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import styles from './styles.css?inline';

const ShadowDomWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const hostRef = useRef<HTMLDivElement>(null);
  const [shadowRoot, setShadowRoot] = useState<ShadowRoot | null>(null);

  useEffect(() => {
    if (hostRef.current && !shadowRoot) {
      const newShadowRoot = hostRef.current.attachShadow({ mode: 'open' });
      const styleElement = document.createElement('style');
      styleElement.textContent = styles;
      newShadowRoot.appendChild(styleElement);
      setShadowRoot(newShadowRoot);
    }
  }, [shadowRoot]);

  return (
    <div ref={hostRef}>
      {shadowRoot && createPortal(children, shadowRoot)}
    </div>
  );
};

const App: React.FC = () => {
  const handleAnalyzeClick = () => {
    // ==================================================
    // ВОТ НАШЕ ПРИВЕТСТВИЕ!
    // ==================================================
    alert("Привет! Кнопка от AI-Аналитика работает!");
    // ==================================================
  };

  return (
    <ShadowDomWrapper>
      <div className="ai-analyzer-container">
        <button className="analyze-button" onClick={handleAnalyzeClick}>
          🤖 AI-Анализ (Тестовая кнопка)
        </button>
      </div>
    </ShadowDomWrapper>
  );
};

export default App;
