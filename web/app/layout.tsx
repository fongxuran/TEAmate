import './globals.css';

import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'TEAmate',
  description: 'Local MVP UI scaffolding',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <main className="container">{children}</main>
      </body>
    </html>
  );
}
