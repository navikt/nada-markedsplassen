import { ExpansionCard, HStack, Link, List, Select, VStack } from "@navikt/ds-react";
import { WorkstationContainer } from "../../../lib/rest/generatedDto";
import React, { useEffect, useRef, useState } from "react";
import Markdown from "react-markdown";
import { useWorkstationOptions } from "../queries";
import { CodeIcon, ExternalLinkIcon } from "@navikt/aksel-icons";
import { ColorAuxText } from "../designTokens";

const IMAGE_LABEL_TITLE = 'org.opencontainers.image.title';
const IMAGE_LABEL_DESCRIPTION = 'org.opencontainers.image.description';
const IMAGE_LABEL_SOURCE = 'org.opencontainers.image.source';

export interface ContainerImageSelectorProps {
    initialContainerImage: string;
    handleSetContainerImage: (containerImage: string) => void;
}

export const ContainerImageSelector = (props: ContainerImageSelectorProps) => {
    const options = useWorkstationOptions()

    const selectedImageRef = useRef<HTMLSelectElement>(null);

    useEffect(() => {
        props.handleSetContainerImage(selectedImageRef.current?.value || '')
    }, [selectedImageRef, props]);

    const containerImagesMap = new Map<string, WorkstationContainer>(
        options.data?.containerImages
            .filter((image): image is WorkstationContainer => image !== undefined)
            .map(image => [image.image, image]) || []
    );

    if (options.isLoading) {
        return <Select label="Velg utviklingsmiljø" disabled>Laster...</Select>
    }

    const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        props.handleSetContainerImage(event.target.value)
    };

    function imageDetails() {
        const image = containerImagesMap.get(selectedImageRef.current?.value || '')
        if (!image) {
            return <></>
        }

        const hasLabels = Object.keys(image.labels || {}).length > 0
        const hasDocumentation = image.documentation !== ''

        if (!hasLabels && !hasDocumentation) {
            return <></>
        }

        return (
            <div className="mt-2 ml-2 flex flex-row gap-4">
                <p className="text-sm" style={{
                    color: ColorAuxText
                }}>{image.labels[IMAGE_LABEL_TITLE]}: {image.labels[IMAGE_LABEL_DESCRIPTION]}</p>
                <Link className="text-sm" href={image.labels[IMAGE_LABEL_SOURCE]}>Imagelenke<ExternalLinkIcon /></Link>
            </div>
        )
    }

    return (
        <VStack>
            <Select className="!p-0" ref={selectedImageRef} value={props.initialContainerImage} label="" onChange={handleChange}>
                {Array.from(containerImagesMap.entries()).map(([name, image]) => {
                    return (
                        <option key={image.image} value={image.image}>
                            {image.labels?.['org.opencontainers.image.title'] || image.description}
                        </option>
                    )
                }
                )
                }
            </Select>
            {imageDetails()}
        </VStack>
    );
}

export default ContainerImageSelector;
