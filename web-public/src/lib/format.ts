export const aed = (n: number) => `AED ${n.toFixed(2)}`;

export function daysLeftLabel(days: number): string {
  if (days <= 0) return "Ends today";
  if (days === 1) return "1 day left";
  return `${days} days left`;
}

// Colour class for the "how soon" chip.
export function urgency(days: number): "bad" | "warn" | "ok" {
  return days <= 0 ? "bad" : days <= 2 ? "warn" : "ok";
}

export function longDate(iso: string): string {
  const [y, m, d] = iso.split("-").map(Number);
  const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
  return `${d} ${months[m - 1]} ${y}`;
}
