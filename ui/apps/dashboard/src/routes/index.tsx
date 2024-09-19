import { createBrowserRouter, RouterProvider } from 'react-router-dom';

import { routes } from './route.tsx';
import { routerBase } from '@/services/base';
import * as Sentry from '@sentry/react';

const sentryCreateBrowserRouter =
  Sentry.wrapCreateBrowserRouter(createBrowserRouter);
const router = sentryCreateBrowserRouter(routes, {
  basename: routerBase,
});
export default function Router() {
  return <RouterProvider router={router} />;
}
