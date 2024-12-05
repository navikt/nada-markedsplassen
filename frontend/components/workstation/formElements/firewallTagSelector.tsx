import {UNSAFE_Combobox} from "@navikt/ds-react";
import {useState} from "react";
import {useWorkstationMine, useWorkstationOptions} from "../queries";

interface FirewallTagSelectorProps {
    initialFirewallTags?: string[];
    onFirewallChange: (tags: string[]) => void;
}

export const FirewallTagSelector = (props: FirewallTagSelectorProps) => {
    const {onFirewallChange, initialFirewallTags} = props;

    const {data: workstationOptions, isLoading: optionsLoading} = useWorkstationOptions()
    const {data: workstation, isLoading: workstationLoading} = useWorkstationMine()

    const defaultFirewallTags = initialFirewallTags || workstation?.config?.firewallRulesAllowList || [];
    const firewallTags = workstationOptions?.firewallTags?.filter(tag => tag !== undefined) ?? [];
    const [selectedFirewallTags, setSelectedFirewallTags] = useState(new Set(defaultFirewallTags))

    const handleChange = (tagValue: string, isSelected: boolean) => {
        if (isSelected) {
            setSelectedFirewallTags(new Set(selectedFirewallTags.add(tagValue)))
            onFirewallChange(Array.from(selectedFirewallTags))
            return
        }

        selectedFirewallTags.delete(tagValue)
        setSelectedFirewallTags(new Set(selectedFirewallTags))
        onFirewallChange(Array.from(selectedFirewallTags))
    }

    if (optionsLoading || workstationLoading) {
        return <UNSAFE_Combobox label="Velg hvilke onprem-kilder du trenger åpninger mot" options={["Laster..."]}
                                selectedOptions={["Laster..."]} isMultiSelect/>
    }

    return (
        <UNSAFE_Combobox
            label="Velg hvilke onprem-kilder du trenger åpninger mot"
            options={firewallTags.map((o) => ({label: o.name, value: o.name}))}
            selectedOptions={Array.from(selectedFirewallTags)}
            isMultiSelect
            onToggleSelected={(tagValue, isSelected) => handleChange(tagValue, isSelected)}
        />
    );
}

export default FirewallTagSelector;
