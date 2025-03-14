import {Radio, RadioGroup, Select} from "@navikt/ds-react";
import {WorkstationMachineType} from "../../../lib/rest/generatedDto";
import {useWorkstationOptions} from "../queries";
import React, {useEffect, useRef, useState} from "react";

export interface MachineTypeSelectorProps {
    initialMachineType: string | undefined;
    handleSetMachineType: (machineType: string) => void;
}

const MachineTypeSelector = (props: MachineTypeSelectorProps) => {
    const {initialMachineType, handleSetMachineType} = props

    const options = useWorkstationOptions()

    const machineTypes: WorkstationMachineType[] = options.data?.machineTypes?.filter((type): type is WorkstationMachineType => type !== undefined) ?? [];

    if (options.isLoading) {
        return <Select label="Velg maskintype" disabled>Laster...</Select>
    }

    return (
        <RadioGroup legend="Velg maskintype" className="machine-selector" value={initialMachineType} onChange={v=>{
            handleSetMachineType(v)
            }}>
            {machineTypes.map((type, index) => {
                const dailyCost = (type.hourlyCost * 24).toFixed(0);
                const description = type.vCPU + " virtuelle kjerner, " + type.memoryGB + " GB minne, kr " + dailyCost + ",-/døgn";
                return <Radio value={type.machineType} key={index} description={description}>{type.machineType}</Radio>
            })}
            <style>
                {`
                    .machine-selector > .navds-radio-buttons {
                        display: flex;
                        flex-wrap: wrap;
                    }
                    .machine-selector > .navds-radio-buttons > .navds-radio {
                        width: 50%;
                        grow: 1;
                    }`}
            </style>
        </RadioGroup>
    )
}

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
