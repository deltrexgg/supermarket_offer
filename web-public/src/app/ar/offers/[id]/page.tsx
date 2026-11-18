import type { Metadata } from "next";
import OfferView from "@/components/OfferView";
import { offerMetadata } from "@/lib/seo";

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }): Promise<Metadata> {
  const { id } = await params;
  return offerMetadata("ar", id);
}

export default function OfferPage({ params }: { params: Promise<{ id: string }> }) {
  return <OfferView lang="ar" params={params} />;
}
