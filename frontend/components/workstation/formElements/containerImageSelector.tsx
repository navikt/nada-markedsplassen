import {Accordion, Link, List, Select, VStack} from "@navikt/ds-react";
import {WorkstationContainer} from "../../../lib/rest/generatedDto";
import React, {forwardRef, useState} from "react";
import {useWorkstation} from "../WorkstationStateProvider";
import Markdown from "react-markdown";

const knownLabels = new Map<string, string>([
    ['org.opencontainers.image.title', 'Tittel'],
    ['org.opencontainers.image.description', 'Beskrivelse'],
    ['org.opencontainers.image.source', 'Kilde'],
]);

export const ContainerImageSelector = forwardRef<HTMLSelectElement, {}>(({}, ref) => {
    const emptyContainerImage: WorkstationContainer = {image: '', description: '', labels: {}, documentation: ''};

    const {workstation, workstationOptions} = useWorkstation()

    const containerImagesMap = new Map<string, WorkstationContainer>(
        workstationOptions?.containerImages
            .filter((image): image is WorkstationContainer => image !== undefined)
            .map(image => [image.image, image]) || []
    );

    const defaultContainerImage: WorkstationContainer = containerImagesMap.get(workstation?.config?.image ?? '') || emptyContainerImage

    const [selectedImage, setSelectedImage] = useState(defaultContainerImage);

    const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        setSelectedImage(containerImagesMap.get(event.target.value) || emptyContainerImage);
    };

    function describeImage() {
        if (!selectedImage) {
            return <></>
        }

        const hasLabels = Object.keys(selectedImage.labels || {}).length > 0
        const hasDocumentation = selectedImage.documentation !== ''

        if (!hasLabels && !hasDocumentation) {
            return <></>
        }

        return (
            <Accordion variant="neutral" size="small" headingSize="xsmall">
                {hasLabels && <Accordion.Item defaultOpen>
                    <Accordion.Header>Flere detaljer</Accordion.Header>
                    <Accordion.Content>
                        <List>
                            {Object.entries(selectedImage.labels || {}).map(([key, value]) => (
                                <List.Item key={key}>
                                    <strong>{knownLabels.get(key) || key}:</strong> {key === 'org.opencontainers.image.source' ? <Link href={value}>{value}</Link> : value}
                                </List.Item>
                            ))}
                        </List>
                    </Accordion.Content>
                </Accordion.Item>}
                {hasDocumentation && <Accordion.Item>
                    <Accordion.Header>Dokumentasjon</Accordion.Header>
                    <Accordion.Content>
                        <Markdown>{selectedImage.documentation}</Markdown>
                    </Accordion.Content>
                </Accordion.Item>}
            </Accordion>
        )
    }

    return (
        <VStack gap="2">
            <Select ref={ref} defaultValue={defaultContainerImage.image} label="Velg utviklingsmiljø"
                    onChange={handleChange}>
                {Array.from(containerImagesMap.entries()).map(([name, image]) => (
                    <option key={name} value={name}>
                        {image.labels?.['org.opencontainers.image.title'] || image.description}
                    </option>
                ))}
            </Select>
            {describeImage()}
        </VStack>
    );
})

ContainerImageSelector.displayName = "ContainerImageSelector";

export default ContainerImageSelector;
