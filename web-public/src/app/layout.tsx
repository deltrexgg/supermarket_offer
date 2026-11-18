import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Supermarket Offers",
  description: "Browse the latest supermarket deals",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
