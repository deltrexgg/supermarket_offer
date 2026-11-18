"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import Logo from "@/components/Logo";
import type { Card } from "@/lib/api";
import { usePrefersReducedMotion } from "@/lib/hooks";
import {
  daysLeftLabel, goToSlide, offerPath, offLabel, pick, priceLabel, saveLabel, strings, type Lang,
} from "@/lib/i18n";

const INTERVAL_MS = 7000;

// Slow auto-sliding banner of the biggest discounts. It stands still while the
// pointer or keyboard focus is on it, when the visitor pauses it, and when the
// system asks for reduced motion. Phones can swipe.
export default function TopOffers({ offers, lang }: { offers: Card[]; lang: Lang }) {
  const t = strings(lang);
  const rtl = lang === "ar";
  const n = offers.length;
  const [i, setI] = useState(0);
  const [hold, setHold] = useState(false);
  const [user, setUser] = useState(false);
  const reduced = usePrefersReducedMotion();
  const startX = useRef<number | null>(null);

  useEffect(() => {
    if (n < 2 || hold || user || reduced) return;
    const id = setInterval(() => setI((v) => (v + 1) % n), INTERVAL_MS);
    return () => clearInterval(id);
  }, [n, hold, user, reduced]);

  if (n === 0) return null;
  const step = (d: number) => setI((v) => (((v + d) % n) + n) % n);

  const onUp = (x: number) => {
    if (startX.current === null) return;
    const dx = x - startX.current;
    startX.current = null;
    if (Math.abs(dx) < 50) return;
    // Dragging towards the start edge reveals the next slide.
    step((dx < 0) !== rtl ? 1 : -1);
  };

  return (
    <section
      className="slider"
      aria-roledescription="carousel"
      aria-label={t.carouselLabel}
      onMouseEnter={() => setHold(true)}
      onMouseLeave={() => setHold(false)}
      onFocus={() => setHold(true)}
      onBlur={() => setHold(false)}
      onPointerDown={(e) => { startX.current = e.clientX; }}
      onPointerUp={(e) => onUp(e.clientX)}
      onPointerCancel={() => { startX.current = null; }}
    >
      <div className="track" style={{ transform: `translateX(${(rtl ? 1 : -1) * i * 100}%)` }}>
        {offers.map((o, idx) => {
          const name = pick(lang, o.item_name, o.item_name_ar);
          const active = idx === i;
          return (
            <div className="slide" key={o.id} aria-roledescription="slide" aria-hidden={!active}>
              <div className="slide-txt">
                <span className="eyebrow">{t.topOffer}</span>
                <span className="slide-pct">
                  <span className="pctn">{offLabel(lang, o.discount_pct)}</span>
                  <span className="pctw">{t.offWord}</span>
                </span>
                <h2 dir="auto">{name}</h2>
                <span className="shopname" dir="auto">
                  <Logo name={o.supermarket_name} url={o.logo_url} />
                  {o.supermarket_name}
                </span>
                <div className="slide-price">
                  <span className="now srf">{priceLabel(lang, o.offer_price)}</span>
                  <span className="was">{priceLabel(lang, o.regular_price)}</span>
                </div>
                <span className="slide-meta">
                  {o.regular_price > o.offer_price && (
                    <span className="savetag">{saveLabel(lang, o.regular_price, o.offer_price)}</span>
                  )}
                  <span className="dleft">{daysLeftLabel(lang, o.days_left)}</span>
                </span>
                <Link className="cta" href={offerPath(lang, o.id)} tabIndex={active ? 0 : -1} draggable={false}>{t.viewOffer}</Link>
              </div>
              <div className="slide-art">
                {o.photo_url ? (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={o.photo_url} alt="" draggable={false} />
                ) : (
                  <span className="tile-letter" aria-hidden="true">{name.charAt(0).toUpperCase()}</span>
                )}
              </div>
            </div>
          );
        })}
      </div>
      {n > 1 && (
        <div className="slider-foot">
          <div className="slider-bar">
            <button type="button" className="ctl" aria-label={user ? t.playSlides : t.pauseSlides} onClick={() => setUser((v) => !v)}>
              {user ? (
                <svg className="ic" width="15" height="15" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 5l12 7-12 7z" fill="currentColor" stroke="none" /></svg>
              ) : (
                <svg className="ic" width="15" height="15" viewBox="0 0 24 24" aria-hidden="true"><path d="M9 5v14M15 5v14" /></svg>
              )}
            </button>
            <div className="dots">
              {offers.map((o, idx) => (
                <button key={o.id} type="button" className="dotbtn" aria-label={goToSlide(lang, idx + 1)} aria-current={idx === i} onClick={() => setI(idx)}>
                  <span />
                </button>
              ))}
            </div>
          </div>
          <div className="slider-bar slider-arrows">
            <button type="button" className="ctl" aria-label={t.prevSlide} onClick={() => step(-1)}>
              <svg className="ic" width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" style={{ transform: rtl ? "none" : "rotate(180deg)" }}><path d="M9 6l6 6-6 6" /></svg>
            </button>
            <button type="button" className="ctl" aria-label={t.nextSlide} onClick={() => step(1)}>
              <svg className="ic" width="16" height="16" viewBox="0 0 24 24" aria-hidden="true" style={{ transform: rtl ? "rotate(180deg)" : "none" }}><path d="M9 6l6 6-6 6" /></svg>
            </button>
          </div>
        </div>
      )}
    </section>
  );
}
