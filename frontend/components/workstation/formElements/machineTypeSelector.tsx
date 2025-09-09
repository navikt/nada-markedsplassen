import { Select } from "@navikt/ds-react";
import { useState } from "react";
import { WorkstationMachineType } from "../../../lib/rest/generatedDto";
import { useWorkstationOptions } from "../queries";

export interface MachineTypeSelectorProps {
    initialMachineType: string | undefined;
    handleSetMachineType: (machineType: string) => void;
}

const MachineTypeSelector = (props: MachineTypeSelectorProps) => {
    const { initialMachineType, handleSetMachineType } = props;
    const options = useWorkstationOptions();

    const machineTypes: WorkstationMachineType[] = options.data?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    if (options.isLoading) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    return (
        <Select
            label="Velg maskintype"
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

const useMachineTypeSelector = (defaultValue: string) => {
    const { data: machineTypes, isLoading, error } = useWorkstationOptions()
    const [selectedMachineType, setSelectedMachineType] = useState<string>(defaultValue)

    return {
        machineTypes,
        isLoading,
        error,
        selectedMachineType,
        setSelectedMachineType,
        MachineTypeSelector: ()=>(<MachineTypeSelector initialMachineType={selectedMachineType} handleSetMachineType={setSelectedMachineType}/>)
    }
}


export default useMachineTypeSelector;
