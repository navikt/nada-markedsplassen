import {Select} from "@navikt/ds-react";
import {forwardRef, useEffect, useState} from "react";
import {useWorkstation} from "../WorkstationStateProvider";
import {WorkstationMachineType} from "../../../lib/rest/generatedDto";

export const MachineTypeSelector = forwardRef<HTMLSelectElement, {}>(({}, ref) => {
    const {workstation, workstationOptions} = useWorkstation();
    const [error, setError] = useState<string | null>(null);

    const machineTypes: WorkstationMachineType[] = workstationOptions?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    const defaultMachineType: string = workstation?.config?.machineType && workstation.config.machineType !== "" ?
        workstation.config.machineType :
        (machineTypes[0]?.machineType ?? "")

    useEffect(() => {
        if (defaultMachineType === "") {
            setError("Maskintype kan ikke være tom.");
        } else {
            setError(null);
        }
    }, [defaultMachineType]);

    return (
        <Select defaultValue={defaultMachineType} label="Velg maskintype" ref={ref} error={error}>
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
