import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: { default: "Dubai Supermarket Offers", template: "%s | Dubai Supermarket Offers" },
  description: "Today's supermarket offers in Dubai: compare prices, discounts and branches.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header className="top">
          <div className="wrap">
            <Link href="/" className="logo"><i>%</i> Dubai Offers</Link>
          </div>
        </header>
        <main className="wrap">{children}</main>
        <footer className="wrap muted foot">Prices in AED. Offers are valid on the dates shown, Dubai time.</footer>
      </body>
    </html>
  );
}
