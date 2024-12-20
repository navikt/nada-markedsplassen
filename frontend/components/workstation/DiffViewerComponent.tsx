import { MinusIcon, PlusIcon } from "@navikt/aksel-icons";
import { Heading, HStack, Tag, VStack } from "@navikt/ds-react";
import React from "react";
import {
    Diff,
    WorkstationDiffContainerImage,
    WorkstationDiffDisableGlobalURLAllowList,
    WorkstationDiffMachineType, WorkstationDiffOnPremAllowList,
    WorkstationDiffURLAllowList,
} from "../../lib/rest/generatedDto";

export const WorkstationDiffDescriptions: Record<string, string> = {
    [WorkstationDiffDisableGlobalURLAllowList]: "Skru av globale åpninger",
    [WorkstationDiffContainerImage]: "Utviklingsmiljø",
    [WorkstationDiffMachineType]: "Maskintype",
    [WorkstationDiffURLAllowList]: "Tillate URLer",
    [WorkstationDiffOnPremAllowList]: "On-prem kilder",
};

interface DiffViewerProps {
    diff: Record<string, Diff | undefined>;
}

const AddedItems: React.FC<{ items?: string[] }> = ({items}) => (
    <>
        {items?.filter(item => item).map((item, index) => (
            <Tag key={index} variant="success" icon={<PlusIcon title="lagt til" fontSize="1.5rem"/>}>{item}</Tag>
        ))}
    </>
);

const RemovedItems: React.FC<{ items?: string[] }> = ({items}) => (
    <>
        {items?.filter(item => item).map((item, index) => (
            <Tag key={index} variant="error" icon={<MinusIcon title="lagt til" fontSize="1.5rem"/>}>{item}</Tag>
        ))}
    </>
);

const DiffViewerComponent: React.FC<DiffViewerProps> = ({diff}) => {
    if (!diff || Object.keys(diff).length === 0) {
        return <div>Ingen endringer fra sist kjøring.</div>;
    }

    return (
        <VStack gap="4">
            {Object.entries(diff).map(([key, value]) => (
                <VStack key={key} gap="1">
                    <Heading size="xsmall">{WorkstationDiffDescriptions[key]}</Heading>
                    <HStack gap="2">
                        <AddedItems items={value?.added}/>
                        <RemovedItems items={value?.removed}/>
                    </HStack>
                </VStack>
            ))}
        </VStack>
    );
};

export default DiffViewerComponent;
