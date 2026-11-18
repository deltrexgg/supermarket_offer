import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  robots: { index: false, follow: false },
  title: "Supermarket Portal",
  description: "Merchant dashboard for supermarket offers",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
