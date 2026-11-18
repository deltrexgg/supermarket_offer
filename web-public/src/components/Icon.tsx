// Small inline stroke icons (no icon library needed).
const PATHS = {
  search: "M20 20l-4-4M11 18a7 7 0 1 1 0-14 7 7 0 0 1 0 14z",
  arrow: "M5 12h14M13 6l6 6-6 6",
  back: "M19 12H5M11 6l-6 6 6 6",
  chevron: "M9 6l6 6-6 6",
  pause: "M8 5v14M16 5v14",
  play: "M7 5l12 7-12 7z",
  filter: "M4 6h16M7 12h10M10 18h4",
  pin: "M12 21s7-6.2 7-11.5A7 7 0 0 0 5 9.5C5 14.8 12 21 12 21zM12 12a2.5 2.5 0 1 1 0-5 2.5 2.5 0 0 1 0 5z",
  phone: "M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z",
  calendar: "M4 8a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v9a3 3 0 0 1-3 3H7a3 3 0 0 1-3-3zM4 10h16M9 3v4M15 3v4",
  tag: "M20 12l-8 8-9-9V3h8l9 9zM7.5 7.5h.01",
  // category icons
  leaf: "M5 19c0-8 5-14 14-14 0 9-6 14-14 14zM5 19l7-7",
  milk: "M9 3h6v3l2 3v11a1 1 0 0 1-1 1H8a1 1 0 0 1-1-1V9l2-3zM7 12h10",
  meat: "M14 4a6 6 0 0 1 0 12l-3 3-2-2 1-3a6 6 0 0 1 4-10z",
  grain: "M12 3v18M12 8c-3 0-4-2-4-4 3 0 4 2 4 4zM12 8c3 0 4-2 4-4-3 0-4 2-4 4zM12 14c-3 0-4-2-4-4 3 0 4 2 4 4zM12 14c3 0 4-2 4-4-3 0-4 2-4 4z",
  cup: "M6 4h12l-1.5 16h-9zM6.5 9h11",
  home: "M4 11l8-7 8 7v9a1 1 0 0 1-1 1h-4v-6H9v6H5a1 1 0 0 1-1-1z",
  bread: "M5 11a4 4 0 0 1 3-6h8a4 4 0 0 1 3 6v8a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1z",
  grid: "M4 4h7v7H4zM13 4h7v7h-7zM4 13h7v7H4zM13 13h7v7h-7z",
  basket: "M4 9h16l-1.5 10a2 2 0 0 1-2 1.7H7.5a2 2 0 0 1-2-1.7zM8 9l3-5M16 9l-3-5",
} as const;

export type IconName = keyof typeof PATHS;

export default function Icon({ name, size = 20, className = "ic" }: { name: IconName; size?: number; className?: string }) {
  return (
    <svg className={className} width={size} height={size} viewBox="0 0 24 24" aria-hidden="true">
      <path d={PATHS[name]} />
    </svg>
  );
}

// Picks a category icon from its name; unknown names get the basket.
export function categoryIcon(name: string): IconName {
  const n = name.toLowerCase();
  if (/fruit|veg/.test(n)) return "leaf";
  if (/dairy|milk|egg/.test(n)) return "milk";
  if (/meat|fish|poultry|chicken/.test(n)) return "meat";
  if (/bever|drink|juice/.test(n)) return "cup";
  if (/house|clean/.test(n)) return "home";
  if (/bake|bread/.test(n)) return "bread";
  if (/grocer|rice|grain|oil|spice/.test(n)) return "grain";
  return "basket";
}

