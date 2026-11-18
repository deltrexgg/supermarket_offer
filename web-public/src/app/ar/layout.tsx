import type { Metadata } from "next";
import "../globals.css";
import Shell from "@/components/Shell";
import { metadataBase } from "@/lib/seo";
import { strings } from "@/lib/i18n";

export const metadata: Metadata = {
  metadataBase,
  title: { default: strings("ar").metaTitle, template: `%s | ${strings("ar").siteName}` },
  description: strings("ar").metaDesc,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ar" dir="rtl">
      <body>
        <Shell lang="ar">{children}</Shell>
      </body>
    </html>
  );
}
