import type { Option } from "@/lib/api";

// A plain GET form: it works without JavaScript and keeps every filter in the URL.
export default function Filters({
  areas, categories, supermarkets, current,
}: {
  areas: Option[]; categories: Option[]; supermarkets: Option[];
  current: { q?: string; area?: string; category?: string; supermarket?: string; sort?: string };
}) {
  const select = (name: string, label: string, options: Option[], value?: string) => (
    <div className="field">
      <label htmlFor={name}>{label}</label>
      <select id={name} name={name} defaultValue={value ?? ""}>
        <option value="">All</option>
        {options.map((o) => <option key={o.id} value={o.id}>{o.name}</option>)}
      </select>
    </div>
  );
  return (
    <form method="get" action="/" className="filterform">
      <div className="search">
        <input type="search" name="q" defaultValue={current.q ?? ""} placeholder="Search items or brands" aria-label="Search offers" />
        <button className="primary" type="submit">Search</button>
      </div>
      <div className="filters">
        {select("area", "Area", areas, current.area)}
        {select("category", "Category", categories, current.category)}
        {select("supermarket", "Supermarket", supermarkets, current.supermarket)}
        <div className="field">
          <label htmlFor="sort">Sort by</label>
          <select id="sort" name="sort" defaultValue={current.sort ?? "highest_discount"}>
            <option value="highest_discount">Highest discount</option>
            <option value="lowest_price">Lowest price</option>
            <option value="ending_soon">Ending soon</option>
          </select>
        </div>
      </div>
    </form>
  );
}
