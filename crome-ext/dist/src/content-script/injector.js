(function(){console.log("AI-Analyzer: Injector script running.");console.log("Current URL:",location.href);console.log("Current pathname:",location.pathname);const i="ai-analyzer-react-root";let y=location.href,s=null;const m=()=>{const n=window.location.href;console.log("AI-Analyzer: Extracting traceId from URL:",n);try{const t=new URL(n);console.log("AI-Analyzer: Parsed URL object:",{pathname:t.pathname,search:t.search,searchParams:Array.from(t.searchParams.entries())});const r=t.searchParams.get("peek");if(console.log("AI-Analyzer: Peek param value:",r),r&&r.trim()!==""){const o=r.trim();return console.log("AI-Analyzer: TraceId found in peek param:",o),s=o,o}const e=n.match(/\/traces\/([a-zA-Z0-9_-]+)/);if(e&&e[1]){const o=e[1];return console.log("AI-Analyzer: TraceId found in path:",o),s=o,o}return s?(console.log("AI-Analyzer: Using cached traceId:",s),s):(console.warn("AI-Analyzer: TraceId not found in URL"),console.warn("AI-Analyzer: URL pathname:",t.pathname),console.warn("AI-Analyzer: URL search params:",t.search),null)}catch(t){return console.error("AI-Analyzer: Error extracting traceId:",t),null}},b=()=>{const n=document.getElementById("ai-analyzer-progress");n&&n.remove();const t=document.createElement("div");return t.id="ai-analyzer-progress",t.style.position="fixed",t.style.top="160px",t.style.right="20px",t.style.zIndex="9999",t.style.width="250px",t.style.backgroundColor="#ffffff",t.style.border="2px solid #6d28d9",t.style.borderRadius="8px",t.style.padding="16px",t.style.boxShadow="0 4px 12px rgba(0,0,0,0.15)",t.style.fontFamily="system-ui, -apple-system, sans-serif",t.innerHTML=`
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
  `,document.body.appendChild(t),{update:r=>{const e=document.getElementById("progress-step");e&&(e.textContent=r)},remove:()=>{t.remove()}}},h=(n,t)=>{const r=document.getElementById("ai-analyzer-modal");r&&r.remove();const e=document.createElement("div");e.id="ai-analyzer-modal",e.style.position="fixed",e.style.top="0",e.style.left="0",e.style.width="100%",e.style.height="100%",e.style.backgroundColor="rgba(0, 0, 0, 0.5)",e.style.zIndex="10000",e.style.display="flex",e.style.alignItems="center",e.style.justifyContent="center";const o=document.createElement("div");o.style.backgroundColor="white",o.style.borderRadius="12px",o.style.padding="24px",o.style.maxWidth="600px",o.style.maxHeight="80vh",o.style.overflow="auto",o.style.boxShadow="0 4px 20px rgba(0,0,0,0.3)";const g=n.data||n,d=g.analysisSummary||{},a=g.detailedAnalysis||{},f={HEALTHY:"#10b981",WARNING:"#f59e0b",ERROR:"#ef4444"}[d.overallStatus]||"#6b7280";o.innerHTML=`
    <div style="margin-bottom: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h2 style="margin: 0; font-size: 24px; color: #1f2937;">🤖 AI Анализ Трейса</h2>
        <button id="ai-analyzer-close" style="background: none; border: none; font-size: 24px; cursor: pointer; color: #6b7280;">×</button>
      </div>
      
      <div style="background-color: #f3f4f6; padding: 12px; border-radius: 8px; margin-bottom: 16px;">
        <div style="font-size: 12px; color: #6b7280; margin-bottom: 4px;">Trace ID</div>
        <div style="font-family: monospace; font-size: 14px; color: #1f2937; word-break: break-all;">${t}</div>
      </div>
    </div>

    <div style="margin-bottom: 20px;">
      <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 12px;">
        <div style="width: 12px; height: 12px; border-radius: 50%; background-color: ${f};"></div>
        <h3 style="margin: 0; font-size: 18px; color: #1f2937;">Статус: ${d.overallStatus||"N/A"}</h3>
      </div>
      <p style="margin: 0; color: #4b5563; line-height: 1.5;">${d.keyFinding||"Анализ завершён"}</p>
    </div>

    ${a.anomalyType&&a.anomalyType!=="NONE"?`
      <div style="border-top: 1px solid #e5e7eb; padding-top: 20px;">
        <h3 style="margin: 0 0 12px 0; font-size: 16px; color: #1f2937;">
          ⚠️ Обнаружена аномалия: <span style="color: #ef4444;">${a.anomalyType}</span>
        </h3>
        
        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; color: #374151; margin-bottom: 4px;">Описание:</div>
          <p style="margin: 0; color: #4b5563; line-height: 1.5;">${a.description||"Нет описания"}</p>
        </div>

        <div style="margin-bottom: 16px;">
          <div style="font-weight: 600; color: #374151; margin-bottom: 4px;">Первопричина:</div>
          <p style="margin: 0; color: #4b5563; line-height: 1.5;">${a.rootCause||"Не определена"}</p>
        </div>

        <div style="background-color: #ecfdf5; padding: 12px; border-radius: 8px; border-left: 4px solid #10b981;">
          <div style="font-weight: 600; color: #065f46; margin-bottom: 4px;">💡 Рекомендация:</div>
          <p style="margin: 0; color: #047857; line-height: 1.5;">${a.recommendation||"Нет рекомендаций"}</p>
        </div>
      </div>
    `:""}

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
  `,e.appendChild(o),document.body.appendChild(e);const c=()=>e.remove();document.getElementById("ai-analyzer-close")?.addEventListener("click",c),document.getElementById("ai-analyzer-ok")?.addEventListener("click",c),e.addEventListener("click",A=>{A.target===e&&c()});const l=document.getElementById("ai-analyzer-ok");l&&(l.addEventListener("mouseenter",()=>{l.style.backgroundColor="#5b21b6"}),l.addEventListener("mouseleave",()=>{l.style.backgroundColor="#6d28d9"}))},x=async n=>{console.log("AI-Analyzer: Sending analyze request for traceId:",n);const t=b();t.update("🔄 Отправка запроса на сервер...");try{const r={type:"ANALYZE_TRACE",traceId:n,timestamp:new Date().toISOString()};console.log("AI-Analyzer: Message payload:",r),setTimeout(()=>t.update("📡 Получение данных из Langfuse..."),500),setTimeout(()=>t.update("🤖 AI анализирует трейс..."),2e3),chrome.runtime.sendMessage(r,e=>{if(t.remove(),chrome.runtime.lastError){console.error("AI-Analyzer: Chrome runtime error:",chrome.runtime.lastError.message),alert(`❌ Ошибка связи с расширением: ${chrome.runtime.lastError.message}`);return}e?(console.log("AI-Analyzer: Received response from background:",e),e.data?(h(e,n),console.log("AI-Analyzer: Analysis completed successfully")):e.error&&(console.error("AI-Analyzer: Error from background:",e.error),alert(`❌ Ошибка: ${e.error}`))):(console.warn("AI-Analyzer: No response received from background"),alert("⚠️ Нет ответа от фонового скрипта"))})}catch(r){console.error("AI-Analyzer: Exception in sendAnalyzeRequest:",r),alert(`❌ Исключение: ${r instanceof Error?r.message:String(r)}`)}},I=async()=>{console.log("AI-Analyzer: Analyze button clicked!"),console.log("AI-Analyzer: Current location.href:",location.href),console.log("AI-Analyzer: Current location.pathname:",location.pathname),console.log("AI-Analyzer: Current location.search:",location.search);const n=document.querySelector(`#${i} button`);if(n){const t=n.textContent;n.textContent="⏳ Анализ...",n.disabled=!0;try{const e=document.getElementById(i)?.dataset.traceId;if(console.log("AI-Analyzer: Stored trace ID from button:",e),!e){console.error("AI-Analyzer: No stored traceId in button!"),alert(`❌ Не удалось определить Trace ID.

Попробуйте обновить страницу.`);return}console.log("AI-Analyzer: Using stored TraceId:",e),await x(e)}catch(r){console.error("AI-Analyzer: Error in handleAnalyzeClick:",r),alert(`❌ Произошла ошибка: ${r instanceof Error?r.message:String(r)}`)}finally{n&&(n.textContent=t,n.disabled=!1)}}},p=()=>{if(console.log("AI-Analyzer: tryInjectApp called. Current pathname:",location.pathname),location.pathname.includes("/traces")){console.log("AI-Analyzer: Trace page detected!");const n=m(),t=n!==null;console.log("AI-Analyzer: Has trace ID:",t,n);const r=document.getElementById(i);if(!t){r&&(console.log("AI-Analyzer: No trace ID, removing button"),r.remove());return}if(r)console.log("AI-Analyzer: App root already exists, skipping injection");else{console.log("AI-Analyzer: App root not found, injecting...");const e=document.createElement("div");e.id=i,e.dataset.traceId=n,console.log("AI-Analyzer: Stored trace ID in button:",n),e.style.position="fixed",e.style.top="100px",e.style.right="20px",e.style.zIndex="9999",e.style.width="auto",e.style.height="auto",e.style.minWidth="250px",e.style.backgroundColor="#f9f9f9",e.style.border="2px solid #6d28d9",e.style.borderRadius="8px",e.style.padding="12px",e.style.boxShadow="0 2px 8px rgba(0,0,0,0.1)",document.body.appendChild(e),console.log("AI-Analyzer: Container added to DOM");const o=document.createElement("button");o.textContent="🤖 AI-Анализ",o.style.backgroundColor="#6d28d9",o.style.color="white",o.style.border="none",o.style.padding="8px 16px",o.style.borderRadius="6px",o.style.fontSize="14px",o.style.fontWeight="500",o.style.cursor="pointer",o.style.transition="background-color 0.2s",o.style.width="100%",o.onmouseenter=()=>{o.disabled||(o.style.backgroundColor="#5b21b6")},o.onmouseleave=()=>{o.disabled||(o.style.backgroundColor="#6d28d9")},o.onclick=I,e.appendChild(o),console.log("AI-Analyzer: Button added with message exchange logic")}}else{console.log("AI-Analyzer: Not on trace page");const n=document.getElementById(i);n&&(n.remove(),console.log("AI-Analyzer: Left trace page, removing app container."))}},u=()=>{if(!document.body){setTimeout(u,50);return}new MutationObserver(()=>{location.href!==y&&(y=location.href,console.log("AI-Analyzer: URL changed to:",y),m(),setTimeout(p,100))}).observe(document.body,{childList:!0,subtree:!0}),console.log("AI-Analyzer: MutationObserver initialized")};u();console.log("AI-Analyzer: Initial injection attempt");setTimeout(p,100);setTimeout(()=>{console.log("AI-Analyzer: Secondary injection attempt"),p()},1e3);
})()