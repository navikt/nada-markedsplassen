import {Link, List, Select} from "@navikt/ds-react";
import { WorkstationContainer as DTOWorkstationContainer } from "../../lib/rest/generatedDto";
import React, {useState} from "react";

interface ContainerImageSelectorProps {
    containerImages: DTOWorkstationContainer[];
    defaultValue?: string;
    onChange: (event: React.ChangeEvent<HTMLSelectElement>) => void;
    onDocumentationLinkClick: () => void;
}

const ContainerImageSelector: React.FC<ContainerImageSelectorProps> = ({containerImages, defaultValue, onChange, onDocumentationLinkClick}) => {
    const [selectedImage, setSelectedImage] = useState<DTOWorkstationContainer | undefined>(
        containerImages.find(image => image.image === defaultValue)
    );

    const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        const selected = containerImages.find(image => image.image === event.target.value);
        setSelectedImage(selected);
        onChange(event);
    };

    return (
        <>
            <Select defaultValue={defaultValue} label="Velg utviklingsmiljø" onChange={handleChange}>
                {containerImages.map((image) => (
                    <option key={image.image} value={image.image}>
                        {image.labels?.['org.opencontainers.image.title'] || image.description}
                    </option>
                ))}
            </Select>
            {selectedImage && (
                <List as="ul" size="small">
                    <List.Item>
                        <Link href={selectedImage.labels?.['org.opencontainers.image.source']}>{selectedImage.labels?.['org.opencontainers.image.description']}</Link>
                    </List.Item>
                    <List.Item>
                        <Link onClick={onDocumentationLinkClick}>Detaljert dokumentasjon for valgt utviklingsmiljø</Link>
                    </List.Item>
                </List>
            )}
        </>
    );
};

export default ContainerImageSelector;
