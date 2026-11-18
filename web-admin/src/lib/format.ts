export const aed = (n: number) => `AED ${n.toFixed(2)}`;

// Today's date in Dubai as YYYY-MM-DD. Call from effects or handlers only
// (not while rendering), so pages can still be prerendered.
export function dubaiToday(): string {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Dubai" }).format(new Date());
}

export function statusLabel(status: string, daysLeft: number): string {
  if (status === "active") return daysLeft === 0 ? "Ends today" : `${daysLeft} days left`;
  if (status === "upcoming") return "Upcoming";
  return "Expired";
}
