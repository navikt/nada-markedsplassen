import { Select } from "@navikt/ds-react";
import { WorkstationMachineType } from "../../lib/rest/generatedDto";

interface MachineTypeSelectorProps {
    machineTypes: WorkstationMachineType[];
    defaultValue?: string;
    onChange: (event: React.ChangeEvent<HTMLSelectElement>) => void;
}

const MachineTypeSelector: React.FC<MachineTypeSelectorProps> = ({ machineTypes, defaultValue, onChange }) => {
    return (
        <Select defaultValue={defaultValue} label="Velg maskintype" description="Du kan når som helst bytte maskintype" onChange={onChange}>
            {machineTypes.map((type) => (
                <option key={type.machineType} value={type.machineType}>
                    {type.machineType} (vCPU: {type.vCPU}, memoryGB: {type.memoryGB})
                </option>
            ))}
        </Select>
    );
};

export default MachineTypeSelector;
