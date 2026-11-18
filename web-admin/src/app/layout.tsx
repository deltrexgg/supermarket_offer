import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Platform Admin",
  description: "Platform owner console for the supermarket offers platform",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
