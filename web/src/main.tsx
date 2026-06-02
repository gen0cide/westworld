import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from './App'
import { UIProvider } from './ui'
import { DragProvider } from './drag'
import './styles.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <UIProvider>
      <DragProvider>
        <App />
      </DragProvider>
    </UIProvider>
  </StrictMode>,
)
