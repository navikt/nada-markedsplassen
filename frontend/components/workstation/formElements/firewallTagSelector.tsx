import { UNSAFE_Combobox } from "@navikt/ds-react";
import { useWorkstation } from "../WorkstationStateProvider";
import {forwardRef, useState} from "react";

export const FirewallTagSelector = forwardRef<HTMLInputElement, {}>(({}, ref) => {
    const {workstation, workstationOptions} = useWorkstation()

    const defaultFirewallTags = workstation?.config?.firewallRulesAllowList ?? [];
    const firewallTags = workstationOptions?.firewallTags ?? [];

    const {selectedFirewallTags, setSelectedFirewallTags} = useState(new Set(defaultFirewallTags))

    const handleFirewallTagChange = (tagValue: string, isSelected: boolean) => {
        if (isSelected) {
            setSelectedFirewallTags(new Set(selectedFirewallTags.add(tagValue)))
            return
        }
        selectedFirewallTags.delete(tagValue)

        setSelectedFirewallTags(new Set(selectedFirewallTags))
    }

    return (
        <UNSAFE_Combobox
            ref={ref}
            label="Velg hvilke onprem-kilder du trenger åpninger mot"
            options={firewallTags.map((o) => ({ label: o.name, value: o.name }))}
            selectedOptions={defaultFirewallTags}
            isMultiSelect
            onToggleSelected={handleFirewallTagChange}
        />
    );
})

FirewallTagSelector.displayName = "FirewallTagSelector";

export default FirewallTagSelector;
