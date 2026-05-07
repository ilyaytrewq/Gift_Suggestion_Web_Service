import '@fontsource/fraunces/700.css';
import '@fontsource/ibm-plex-sans/600.css';
import '@fontsource/manrope/400.css';
import '@fontsource/manrope/600.css';

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { QueryProvider } from './app/providers/query-provider';
import { AppRouter } from './app/router/router';
import './app/styles/index.css';
import { AuthProvider } from './shared/auth/auth-context';
import { AuthBootstrap } from './shared/auth/auth-bootstrap';
import { ToastProvider } from './shared/ui/toast/toast-provider';

const rootElement = document.getElementById('root');

if (!rootElement) {
  throw new Error('Root element was not found');
}

createRoot(rootElement).render(
  <StrictMode>
    <QueryProvider>
      <AuthProvider>
        <ToastProvider>
          <AuthBootstrap>
            <AppRouter />
          </AuthBootstrap>
        </ToastProvider>
      </AuthProvider>
    </QueryProvider>
  </StrictMode>,
);
