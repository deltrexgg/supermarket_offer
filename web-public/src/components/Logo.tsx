// A supermarket's own logo, so shoppers recognise the shop at a glance.
// Without one, the shop's first letter stands in.
export default function Logo({ name, url }: { name: string; url: string | null }) {
  if (!url) {
    return <span className="shoplogo letter" aria-hidden="true">{name.charAt(0).toUpperCase()}</span>;
  }
  // eslint-disable-next-line @next/next/no-img-element
  return <img className="shoplogo" src={url} alt="" loading="lazy" />;
}
