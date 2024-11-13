import React, {useState} from "react";
import {BodyShort, Select} from "@navikt/ds-react";
import {WorkstationContainer as DTOWorkstationContainer} from "../../lib/rest/generatedDto";

interface ContainerImageSelectorProps {
    containerImages: DTOWorkstationContainer[];
    defaultValue?: string;
    onChange: (event: React.ChangeEvent<HTMLSelectElement>) => void;
}

const ContainerImageSelector: React.FC<ContainerImageSelectorProps> = ({containerImages, defaultValue, onChange}) => {
    const [selectedImage, setSelectedImage] = useState<DTOWorkstationContainer | undefined>(
        containerImages.find(image => image.image === defaultValue)
    );

    const handleChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        const selected = containerImages.find(image => image.image === event.target.value);
        setSelectedImage(selected);
        onChange(event);
    };
    return (
        <div className="flex flex-col">
            <Select defaultValue={defaultValue} label="Velg utviklingsmiljø" onChange={onChange}>
                {containerImages.map((image) => (
                    <option key={image.image} value={image.image}>
                        {image.labels?.['org.opencontainers.image.title'] || image.description}
                    </option>
                ))}
            </Select>
            {selectedImage && (
                <BodyShort size="medium" className="pt-2 --a-surface-info-subtle">
                    <b>Beskrivelse:</b> <i>{selectedImage.labels?.['org.opencontainers.image.description']}</i>
                </BodyShort>
            )}
        </div>
    );
};

export default ContainerImageSelector;
