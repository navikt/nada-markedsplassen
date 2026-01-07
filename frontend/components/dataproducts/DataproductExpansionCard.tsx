import { ExpansionCard } from "@navikt/ds-react";
import { ReactNode } from "react";

interface DataproductExpansionCardProps {
  name: string;
  description: string;
  children: ReactNode;
  level: "1" | "2" | "3" | "4" 
}

export function DataproductExpansionCard({ name, description, children, level }: DataproductExpansionCardProps) {
  return (
    <ExpansionCard aria-label={`Dataprodukt: ${name}`}>
      <ExpansionCard.Header>
        <ExpansionCard.Title as={`h${level}`}>{name}</ExpansionCard.Title>
        <ExpansionCard.Description>{description}</ExpansionCard.Description>
      </ExpansionCard.Header>
      <ExpansionCard.Content>
        {children}
      </ExpansionCard.Content>
    </ExpansionCard>
  );
}

