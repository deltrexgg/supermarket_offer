import { useSyncExternalStore } from "react";

const noop = () => () => {};

// False while rendering on the server and during hydration, true afterwards.
export function useHydrated(): boolean {
  return useSyncExternalStore(noop, () => true, () => false);
}

// True when the visitor's system asks for less motion.
export function usePrefersReducedMotion(): boolean {
  return useSyncExternalStore(
    (cb) => {
      const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
      mq.addEventListener("change", cb);
      return () => mq.removeEventListener("change", cb);
    },
    () => window.matchMedia("(prefers-reduced-motion: reduce)").matches,
    () => false,
  );
}
