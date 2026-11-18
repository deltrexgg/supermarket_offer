import type { Metadata } from "next";
import "../globals.css";
import Shell from "@/components/Shell";
import { metadataBase } from "@/lib/seo";
import { strings } from "@/lib/i18n";

export const metadata: Metadata = {
  metadataBase,
  title: { default: strings("en").metaTitle, template: `%s | ${strings("en").siteName}` },
  description: strings("en").metaDesc,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" dir="ltr">
      <body>
        <Shell lang="en">{children}</Shell>
      </body>
    </html>
  );
}
