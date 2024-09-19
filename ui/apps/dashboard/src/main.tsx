import React, { useEffect } from 'react';
import ReactDOM from 'react-dom/client';

import i18nInstance, { getLang } from '@/utils/i18n';
import { initReactI18next } from 'react-i18next';
import { loader } from '@monaco-editor/react';
import * as monaco from 'monaco-editor';
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
// https://github.com/remcohaszing/monaco-yaml/issues/150
import yamlWorker from '@/utils/workaround-yaml.worker?worker';
import enTexts from '../locales/en-US.json';
import zhTexts from '../locales/zh-CN.json';
import { initRoute } from '@/routes/route.tsx';
import * as Sentry from '@sentry/react';
import {
  createRoutesFromChildren,
  matchRoutes,
  useLocation,
  useNavigationType,
} from 'react-router-dom';

console.log('init sentry start');
Sentry.init({
  dsn: 'https://5edc679922b621e4ea52b1f66f24f48d@o4507968956268544.ingest.us.sentry.io/4507968984711168',
  integrations: [
    Sentry.reactRouterV6BrowserTracingIntegration({
      useEffect,
      useLocation,
      useNavigationType,
      createRoutesFromChildren,
      matchRoutes,
    }),
    // Sentry.browserTracingIntegration(),
    Sentry.replayIntegration(),
    Sentry.captureConsoleIntegration({
      levels: ['error'],
    }),
  ],
  // Tracing
  tracesSampleRate: 1.0, //  Capture 100% of the transactions
  // Set 'tracePropagationTargets' to control for which URLs distributed tracing should be enabled
  tracePropagationTargets: [/^\//, /^https:\/\/yourserver\.io\/api/],
  // Session Replay
  replaysSessionSampleRate: 0.1, // This sets the sample rate at 10%. You may want to change it to 100% while in development and then sample at a lower rate in production.
  replaysOnErrorSampleRate: 1.0, // If you're not already sampling the entire session, change the sample rate to 100% when sampling sessions where errors occur.
});
window.addEventListener('error', (e) => {
  console.log('[main.tsx]error', e);
  Sentry.captureException(e);
});
window.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === 'yaml') {
      return new yamlWorker();
    }
    return new editorWorker();
  },
};
loader.config({ monaco });
import App from './App.tsx';
i18nInstance
  .use(initReactI18next) // passes i18n down to react-i18next
  .init({
    debug: true,
    lng: getLang(), // if you're using a language detector, do not define the lng option
    fallbackLng: ['zh-CN'],
    resources: {
      zh: {
        translation: zhTexts,
      },
      en: {
        translation: enTexts,
      },
    },
    interpolation: {
      escapeValue: false, // react already safes from xss => https://www.i18next.com/translation-function/interpolation#unescape
    },
    saveMissing: true, // send not translated keys to endpoint,
    react: {
      useSuspense: false,
    },
  })
  .then(() => {
    initRoute();
    ReactDOM.createRoot(document.getElementById('root')!).render(
      <React.StrictMode>
        <App />
      </React.StrictMode>,
    );
  })
  .catch(() => {});
