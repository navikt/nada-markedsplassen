import { UNSAFE_Combobox } from "@navikt/ds-react";
import { FirewallTag } from "../../../lib/rest/generatedDto";

interface FirewallTagSelectorProps {
    firewallTags: FirewallTag[];
    selectedFirewallHosts: Set<string>;
    onToggleSelected: (tagValue: string, isSelected: boolean) => void;
}

const FirewallTagSelector: React.FC<FirewallTagSelectorProps> = ({ firewallTags, selectedFirewallHosts, onToggleSelected }) => {
    return (
        <UNSAFE_Combobox
            label="Velg hvilke onprem-kilder du trenger åpninger mot"
            options={firewallTags.map((o) => ({ label: o.name, value: o.name }))}
            selectedOptions={Array.from(selectedFirewallHosts)}
            isMultiSelect
            onToggleSelected={onToggleSelected}
        />
    );
};

export default FirewallTagSelector;
