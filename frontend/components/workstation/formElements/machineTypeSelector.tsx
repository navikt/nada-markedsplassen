import {Select} from "@navikt/ds-react";
import {forwardRef} from "react";
import {WorkstationMachineType} from "../../../lib/rest/generatedDto";
import {useWorkstationMine, useWorkstationOptions} from "../queries";

export interface MachineTypeSelectorProps {
    initialMachineType?: string;
    handleSetMachineType?: (machineType: string) => void;
}

export const MachineTypeSelector = forwardRef<HTMLSelectElement, MachineTypeSelectorProps>((props, ref) => {
    const {initialMachineType, handleSetMachineType} = props

    const {data: workstationOptions, isLoading: optionsLoading} = useWorkstationOptions()
    const {data: workstation, isLoading: workstationLoading} = useWorkstationMine()
    const machineTypes: WorkstationMachineType[] = workstationOptions?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    const defaultMachineType: string = initialMachineType ?? (workstation?.config?.machineType && workstation.config.machineType !== "" ?
        workstation.config.machineType :
        (machineTypes[0]?.machineType ?? ""))

    if (optionsLoading || workstationLoading) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    const onChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        if (handleSetMachineType) {
            handleSetMachineType(event.target.value);
        }
    }

    return (
        <Select defaultValue={defaultMachineType} label="Velg maskintype" ref={ref} onChange={onChange}>
            {machineTypes.map((type) => (
                <option key={type.machineType} value={type.machineType}>
                    {type.machineType} ({type.vCPU} virtuelle kjerner, {type.memoryGB}GB minne)
                </option>
            ))}
        </Select>
    )
})

MachineTypeSelector.displayName = "MachineTypeSelector";

export default MachineTypeSelector;
