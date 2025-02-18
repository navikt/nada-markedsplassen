import {Radio, RadioGroup, Select} from "@navikt/ds-react";
import {WorkstationMachineType} from "../../../lib/rest/generatedDto";
import {useWorkstationOptions} from "../queries";
import React, {useEffect, useRef} from "react";

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
        <RadioGroup legend="Velg maskintype" className="machine-selector">
            {machineTypes.map((type) => {
                const dailyCost = (type.hourlyCost * 24).toFixed(0);
                const description = type.vCPU + " virtuelle kjerner, " + type.memoryGB + " GB minne, kr " + dailyCost + ",-/døgn";
                return <Radio value={type.machineType} description={description}>{type.machineType}</Radio>
            })}
            <style>
                {`
                    .machine-selector > .navds-radio-buttons {
                        display: flex;
                        flex-wrap: wrap;
                    }
                    .machine-selector > .navds-radio-buttons > .navds-radio {
                        width: 50%;
                        flex-grow: 1;
                    }`}
            </style>
        </RadioGroup>
    )
}


export default MachineTypeSelector;
