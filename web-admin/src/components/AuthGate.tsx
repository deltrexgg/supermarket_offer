"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { api, type Me } from "@/lib/api";
import { supabase } from "@/lib/supabase";

const NAV: [string, string][] = [
  ["/dashboard", "Dashboard"],
  ["/supermarkets", "Supermarkets"],
  ["/lists", "Areas & categories"],
  ["/offers", "Offers"],
];

// Wraps the signed-in pages: no session, or not an admin account, goes to /login.
export default function AuthGate({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [me, setMe] = useState<Me | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function check() {
      const { data } = await supabase().auth.getSession();
      if (!data.session) {
        router.replace("/login");
        return;
      }
      try {
        const m = await api<Me>("/auth/me");
        if (cancelled) return;
        if (m.role !== "admin") {
          await supabase().auth.signOut();
          router.replace("/login?reason=role");
          return;
        }
        setMe(m);
      } catch {
        if (!cancelled) {
          await supabase().auth.signOut();
          router.replace("/login?reason=account");
        }
      }
    }
    check();
    const { data: sub } = supabase().auth.onAuthStateChange((event) => {
      if (event === "SIGNED_OUT") router.replace("/login");
    });
    return () => {
      cancelled = true;
      sub.subscription.unsubscribe();
    };
  }, [router]);

  if (!me) return <p className="muted center">Loading…</p>;

  return (
    <>
      <header className="topbar">
        <strong>Platform Admin</strong>
        <nav>
          {NAV.map(([href, label]) => (
            <Link key={href} href={href} className={pathname.startsWith(href) ? "active" : ""}>{label}</Link>
          ))}
        </nav>
        <span className="spacer" />
        <span className="muted">{me.email}</span>
        <button className="link" onClick={() => supabase().auth.signOut()}>Sign out</button>
      </header>
      <main className="content">{children}</main>
    </>
  );
}
