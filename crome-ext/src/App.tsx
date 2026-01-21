import { useState } from 'react'
import './App.css'

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <h1>AI-Анализатор Langfuse</h1>
      <div className="card">
        <button onClick={() => setCount((count) => count + 1)}>
          Анализировано: {count}
        </button>
        <p>
          Откройте страницу с трейсом на cloud.langfuse.com для использования расширения
        </p>
      </div>
    </>
  )
}

export default App
