import { Select } from "@navikt/ds-react";
import { useState } from "react";
import { WorkstationMachineType, WorkstationOptions } from "../../../lib/rest/generatedDto";
import { useWorkstationOptions } from "../queries";

export interface MachineTypeSelectorProps {
    initialMachineType: string | undefined;
    options?: WorkstationOptions;
    handleSetMachineType: (machineType: string) => void;
    disabled?: boolean;
}

const MachineTypeSelector = (props: MachineTypeSelectorProps) => {
    const { initialMachineType, options, handleSetMachineType, disabled } = props;

    const machineTypes: WorkstationMachineType[] = options?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    if (!options) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    return (
        <Select
            label=""
            value={initialMachineType}
            onChange={e => handleSetMachineType(e.target.value)}
            disabled={disabled}
        >
            <option value="" disabled>Velg maskintype</option>
            {machineTypes.map((type, index) => {
                const dailyCost = (type.hourlyCost * 24).toFixed(0);
                const description = `${type.vCPU} vCPU, ${type.memoryGB} GB minne, kr ${dailyCost},-/døgn`;
                return (
                    <option value={type.machineType} key={index}>
                        {type.machineType} ({description})
                    </option>
                );
            })}
        </Select>
    );
};

const useMachineTypeSelector = (defaultValue: string, options?: WorkstationOptions) => {
    const [selectedMachineType, setSelectedMachineType] = useState<string>(defaultValue)

    return {
        selectedMachineType,
        setSelectedMachineType,
        MachineTypeSelector: ({disabled}: {disabled?: boolean})=>(<MachineTypeSelector initialMachineType={selectedMachineType} handleSetMachineType={setSelectedMachineType} options={options} disabled={disabled}/>)
    }
}


export default useMachineTypeSelector;
