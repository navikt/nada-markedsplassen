import { Alert, Heading, VStack } from "@navikt/ds-react";
import { ExclamationmarkTriangleIcon } from "@navikt/aksel-icons";

interface GlobalDenyListProps {
    globalDenyList: string[]
}

export const GlobalDenyList = ({ globalDenyList }: GlobalDenyListProps) => {
    if (!globalDenyList || globalDenyList.length === 0) {
        return null;
    }

    return (
        <div className="w-[50rem] mt-6">
            <Heading size="small" level="3" className="mb-3 flex items-center gap-2">
                <ExclamationmarkTriangleIcon className="text-ax-warning-600" fontSize="1.5rem" />
                Globalt blokkerte URLer
            </Heading>
            <Alert variant="warning" size="small" className="mb-4">
                Disse URLene er globalt blokkert og kan ikke aksesseres fra arbeidsstasjoner.
            </Alert>
            <div className="pl-3">
                <VStack gap="space-8">
                    {globalDenyList.map((url, index) => (
                        <div key={index} className="text-sm text-ax-neutral-800 bg-ax-danger-100 p-2 rounded border-l-4 border-ax-danger-500">
                            {url}
                        </div>
                    ))}
                </VStack>
            </div>
        </div>
    );
};
