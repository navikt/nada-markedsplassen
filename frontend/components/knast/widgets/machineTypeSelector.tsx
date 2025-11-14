import { Select } from "@navikt/ds-react";
import { useState } from "react";
import { WorkstationMachineType, WorkstationOptions } from "../../../lib/rest/generatedDto";
import { useWorkstationOptions } from "../queries";

export interface MachineTypeSelectorProps {
    initialMachineType: string | undefined;
    options?: WorkstationOptions;
    handleSetMachineType: (machineType: string) => void;
}

const MachineTypeSelector = (props: MachineTypeSelectorProps) => {
    const { initialMachineType, options, handleSetMachineType } = props;

    const machineTypes: WorkstationMachineType[] = options?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    console.log(options)
    if (!options) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    return (
        <Select
            label=""
            value={initialMachineType}
            onChange={e => handleSetMachineType(e.target.value)}
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
        MachineTypeSelector: ()=>(<MachineTypeSelector initialMachineType={selectedMachineType} handleSetMachineType={setSelectedMachineType} options={options}/>)
    }
}


export default useMachineTypeSelector;
