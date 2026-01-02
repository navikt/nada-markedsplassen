import { LinkCard } from "@navikt/ds-react";
import Link from "next/link";
import { ReactNode } from "react";

interface DatasetLinkCardProps {
  href: string;
  name: string;
  description: string;
  children?: ReactNode;
}

export function DatasetLinkCard({ href, name, description, children }: DatasetLinkCardProps) {
  return (
    <LinkCard className="mb-4">
      <LinkCard.Title>
        <LinkCard.Anchor asChild>
          <Link href={href}>{name}</Link>
        </LinkCard.Anchor>
      </LinkCard.Title>
      <LinkCard.Description className="line-clamp-3">
        {description}
      </LinkCard.Description>
      {children && <LinkCard.Footer>{children}</LinkCard.Footer>}
    </LinkCard>
  );
}
