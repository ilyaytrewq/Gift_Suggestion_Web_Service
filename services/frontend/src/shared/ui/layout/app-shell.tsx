import type { PropsWithChildren } from 'react';

import { Footer } from './footer';
import { Header } from './header';

export function AppShell({ children }: PropsWithChildren): JSX.Element {
  return (
    <div className="app-shell">
      <Header />
      <main className="app-main">{children}</main>
      <Footer />
    </div>
  );
}
