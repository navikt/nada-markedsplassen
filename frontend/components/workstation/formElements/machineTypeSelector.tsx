import {Select} from "@navikt/ds-react";
import {forwardRef} from "react";
import {WorkstationMachineType} from "../../../lib/rest/generatedDto";
import {useWorkstationMine, useWorkstationOptions} from "../../knast/queries";

export const MachineTypeSelector = forwardRef<HTMLSelectElement, {}>(({}, ref) => {
    const {data: workstationOptions, isLoading: optionsLoading} = useWorkstationOptions()
    const {data: workstation, isLoading: workstationLoading} = useWorkstationMine()
    const machineTypes: WorkstationMachineType[] = workstationOptions?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    const defaultMachineType: string = workstation?.config?.machineType && workstation.config.machineType !== "" ?
        workstation.config.machineType :
        (machineTypes[0]?.machineType ?? "")

    if (optionsLoading || workstationLoading) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    return (
        <Select defaultValue={defaultMachineType} label="Velg maskintype" ref={ref}>
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
