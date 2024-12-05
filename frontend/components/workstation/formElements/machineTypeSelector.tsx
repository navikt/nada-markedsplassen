import {Select} from "@navikt/ds-react";
import {useState} from "react";
import {WorkstationMachineType} from "../../../lib/rest/generatedDto";
import {useWorkstationMine, useWorkstationOptions} from "../queries";

export interface MachineTypeSelectorProps {
    initialMachineType: string | undefined;
    handleSetMachineType?: (machineType: string) => void;
}

export const MachineTypeSelector = (props: MachineTypeSelectorProps) => {
    const {initialMachineType, handleSetMachineType} = props

    const options= useWorkstationOptions()
    const workstation = useWorkstationMine()

    const machineTypes: WorkstationMachineType[] = options.data?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    const [selectedMachineType, setSelectedMachineType] = useState<string>(
        initialMachineType || workstation.data?.config?.machineType || machineTypes[0]?.machineType || ""
    );

    if (options.isLoading || workstation.isLoading) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    const onChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        setSelectedMachineType(event.target.value);

        if (handleSetMachineType) {
            handleSetMachineType(event.target.value);
        }
    }

    return (
        <Select value={selectedMachineType} label="Velg maskintype" onChange={onChange}>
            {machineTypes.map((type) => (
                <option key={type.machineType} value={type.machineType}>
                    {type.machineType} ({type.vCPU} virtuelle kjerner, {type.memoryGB}GB minne)
                </option>
            ))}
        </Select>
    )
}


export default MachineTypeSelector;
