import {ExpansionCard, HStack, Link, List, Select, VStack} from "@navikt/ds-react";
import {WorkstationContainer} from "../../../lib/rest/generatedDto";
import React, {forwardRef, useEffect, useState} from "react";
import Markdown from "react-markdown";
import {useWorkstationMine, useWorkstationOptions} from "../queries";
import {CodeIcon} from "@navikt/aksel-icons";

const knownLabels = new Map<string, string>([
    ['org.opencontainers.image.title', 'Tittel'],
    ['org.opencontainers.image.description', 'Beskrivelse'],
    ['org.opencontainers.image.source', 'Kilde'],
]);

export interface ContainerImageSelectorProps {
    initialContainerImage?: WorkstationContainer;
    handleSetContainerImage?: (containerImage: WorkstationContainer) => void;
}

export const ContainerImageSelector = forwardRef<HTMLSelectElement, ContainerImageSelectorProps>((props, ref) => {
    const {initialContainerImage, handleSetContainerImage} = props

    const emptyContainerImage: WorkstationContainer = {image: '', description: '', labels: {}, documentation: ''};

    const {data: workstationOptions, isLoading: optionsLoading} = useWorkstationOptions()
    const {data: workstation, isLoading: workstationLoading} = useWorkstationMine()

    const [selectedImage, setSelectedImage] = useState(initialContainerImage || emptyContainerImage);

    const containerImagesMap = new Map<string, WorkstationContainer>(
        workstationOptions?.containerImages
            .filter((image): image is WorkstationContainer => image !== undefined)
            .map(image => [image.image, image]) || []
    );

    useEffect(() => {
        const initialImage = typeof initialContainerImage === 'string' ? containerImagesMap.get(initialContainerImage) : initialContainerImage;
        const theImage = initialImage || containerImagesMap.get(workstation?.config?.image ?? '') || workstationOptions?.containerImages?.find(() => true) || emptyContainerImage
        setSelectedImage(theImage);

        if (handleSetContainerImage) {
            handleSetContainerImage(theImage)
        }
    }, [optionsLoading, workstationLoading, workstation, workstationOptions]);

    if (optionsLoading || workstationLoading) {
        return <Select label="Velg utviklingsmiljø" disabled>Laster...</Select>
    }

    const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        const theImage = containerImagesMap.get(event.target.value) || workstationOptions?.containerImages?.find(() => true) || emptyContainerImage
        setSelectedImage(theImage);

        if (handleSetContainerImage) {
            handleSetContainerImage(theImage)
        }
    };

    function imageDetails() {
        console.log(selectedImage)

        if (!selectedImage) {
            return <></>
        }

        const hasLabels = Object.keys(selectedImage.labels || {}).length > 0
        const hasDocumentation = selectedImage.documentation !== ''

        if (!hasLabels && !hasDocumentation) {
            return <></>
        }

        return (
            <div className="subtle-card">
                <ExpansionCard size="small" aria-label="Utviklingsmiljø detaljer">
                    <ExpansionCard.Header>
                        <HStack wrap={false} gap="4" align="center">
                            <div>
                                <CodeIcon aria-hidden fontSize="2rem"/>
                            </div>
                            <div>
                                <ExpansionCard.Title size="small">
                                    Flere detaljer om valgt utviklingsmiljø
                                </ExpansionCard.Title>
                                <ExpansionCard.Description>
                                    {selectedImage.labels?.['org.opencontainers.image.title'] || selectedImage.description}
                                </ExpansionCard.Description>
                            </div>
                        </HStack>
                    </ExpansionCard.Header>
                    <ExpansionCard.Content>
                        <List>
                            {Object.entries(selectedImage.labels || {}).map(([key, value]) => (
                                <List.Item key={key}>
                                    <strong>{knownLabels.get(key) || key}:</strong> {key === 'org.opencontainers.image.source' ?
                                    <Link href={value}>{value}</Link> : value}
                                </List.Item>
                            ))}
                        </List>
                        <Markdown>{selectedImage.documentation}</Markdown>
                    </ExpansionCard.Content>
                </ExpansionCard>
                <style>
                    {`
                    .subtle-card {
                      --ac-expansioncard-bg: var(--a-deepblue-50);
                      --ac-expansioncard-border-open-color: var(--a-border-alt-3);
                      --ac-expansioncard-border-hover-color: var(--a-border-alt-3);
                    }`}
                </style>
            </div>
        )
    }

    return (
        <VStack gap="2">
            <Select ref={ref} defaultValue={selectedImage.image} label="Velg utviklingsmiljø"
                    onChange={handleChange}>
                {Array.from(containerImagesMap.entries()).map(([name, image]) => {
                        return (
                            <option key={name} value={name}>
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
})

ContainerImageSelector.displayName = "ContainerImageSelector";

export default ContainerImageSelector;
