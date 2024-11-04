import { Select } from "@navikt/ds-react";
import { WorkstationContainer as DTOWorkstationContainer } from "../../lib/rest/generatedDto";

interface ContainerImageSelectorProps {
    containerImages: DTOWorkstationContainer[];
    defaultValue?: string;
    onChange: (event: React.ChangeEvent<HTMLSelectElement>) => void;
}

const ContainerImageSelector: React.FC<ContainerImageSelectorProps> = ({ containerImages, defaultValue, onChange }) => {
    return (
        <Select defaultValue={defaultValue} label="Velg container image" description="Du kan når som helst bytte image" onChange={onChange}>
            {containerImages.map((image) => (
                <option key={image.image} value={image.image}>
                    {image.labels?.['org.opencontainers.image.title'] || image.description}
                </option>
            ))}
        </Select>
    );
};

export default ContainerImageSelector;
