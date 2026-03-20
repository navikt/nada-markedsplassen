import { BodyShort, VStack } from '@navikt/ds-react';
import { InboxIcon } from '@navikt/aksel-icons';

interface EmptyStateProps {
  title: string
  description?: string
  icon?: React.ComponentType<{ className?: string; fontSize?: string; 'aria-hidden'?: boolean }>
  className?: string
}

export const EmptyState = ({
  title,
  description,
  icon: Icon = InboxIcon,
  className
}: EmptyStateProps) => {
  return (
    <section
      role="status"
      aria-live="polite"
      className={`py-8 text-center ${className || ''}`}
    >
      <VStack gap="space-16" align="center">
        <Icon
          fontSize="3rem"
          className="text-ax-text-neutral-subtle"
        />
        <BodyShort size="large">
          {title}
        </BodyShort>
        {description && (
          <BodyShort className="text-ax-text-neutral-subtle">
            {description}
          </BodyShort>
        )}
      </VStack>
    </section>
  );
};

export default EmptyState;
