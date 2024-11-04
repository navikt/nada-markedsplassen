import { Select } from "@navikt/ds-react";
import { WorkstationContainer as DTOWorkstationContainer } from "../../lib/rest/generatedDto";

interface ContainerImageSelectorProps {
    containerImages: DTOWorkstationContainer[];
    defaultValue?: string;
}

const ContainerImageSelector: React.FC<ContainerImageSelectorProps> = ({ containerImages, defaultValue }) => {
    return (
        <Select defaultValue={defaultValue} label="Velg container image" description="Du kan når som helst bytte image">
            {containerImages.map((image) => (
                <option key={image.image} value={image.image}>
                    {image.labels?.['org.opencontainers.image.title'] || image.description}
                </option>
            ))}
        </Select>
    );
};

export default ContainerImageSelector;
