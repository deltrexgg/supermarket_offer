"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { api, type Me } from "@/lib/api";
import { supabase } from "@/lib/supabase";

// Wraps the signed-in pages: no session, or not a supermarket account, goes to /login.
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
        if (m.role !== "supermarket") {
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
        <strong>Supermarket Portal</strong>
        <nav>
          <Link href="/offers" className={pathname.startsWith("/offers") ? "active" : ""}>Offers</Link>
          <Link href="/branches" className={pathname.startsWith("/branches") ? "active" : ""}>Branches</Link>
        </nav>
        <span className="spacer" />
        <span className="muted">{me.email}</span>
        <button className="link" onClick={() => supabase().auth.signOut()}>Sign out</button>
      </header>
      <main className="content">{children}</main>
    </>
  );
}
