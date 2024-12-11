import {Select} from "@navikt/ds-react";
import {WorkstationMachineType} from "../../../lib/rest/generatedDto";
import {useWorkstationOptions} from "../queries";
import {useEffect, useRef} from "react";

export interface MachineTypeSelectorProps {
    initialMachineType: string | undefined;
    handleSetMachineType: (machineType: string) => void;
}

export const MachineTypeSelector = (props: MachineTypeSelectorProps) => {
    const {initialMachineType, handleSetMachineType} = props

    const selectedMachineTypeRef = useRef<HTMLSelectElement>(null);

    useEffect(() => {
        handleSetMachineType(selectedMachineTypeRef.current?.value || '')
    }, [selectedMachineTypeRef]);

    const options = useWorkstationOptions()

    const machineTypes: WorkstationMachineType[] = options.data?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    if (options.isLoading) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    const onChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        handleSetMachineType(event.target.value);
    }

    return (
        <Select ref={selectedMachineTypeRef} value={initialMachineType} label="Velg maskintype" onChange={onChange}>
            {machineTypes.map((type) => (
                <option key={type.machineType} value={type.machineType}>
                    {type.machineType} ({type.vCPU} virtuelle kjerner, {type.memoryGB}GB minne)
                </option>
            ))}
        </Select>
    )
}


export default MachineTypeSelector;
