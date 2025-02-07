import {UNSAFE_Combobox} from "@navikt/ds-react";
import {useEffect, useState} from "react";
import {useWorkstationMine, useWorkstationOnpremMapping, useWorkstationOptions} from "../queries";
import {useOnpremMapping} from "../../onpremmapping/queries";
import {Host, OnpremHostTypePostgres, OnpremHostTypeTNS} from "../../../lib/rest/generatedDto";

interface FirewallTagSelectorProps {
    enabled?: boolean;
}

const HostsList = ({hosts}: {hosts: Host[]}) => {
    const [selectedHosts, setSelectedHosts] = useState<Set<string>>(new Set());

    const handleChange = (host: string, isSelected: boolean) => {
        if (isSelected) {
            setSelectedHosts(new Set(selectedHosts.add(host)));
        } else {
            selectedHosts.delete(host);
            setSelectedHosts(new Set(selectedHosts));
        }
    };

    return (
        <UNSAFE_Combobox
            label="Select Postgres Host"
            options={hosts.map((host) => ({label: host.Name, value: host.Host}))}
            selectedOptions={selectedHosts ? [selectedHosts] : []}
            onToggleSelected={handleChange}
        />
    );
};

const HostsChecked = ({hosts}: {hosts: Host[]}) => {
    const [selectedHosts, setSelectedHosts] = useState<Set<string>>(new Set());

    const handleChange = (hosts: string[]) => {
        setSelectedHosts(hosts);
    };

    return (
        <div>
            {hosts.map((host) => (
                <div key={host.Host}>
                    <CheckboxGroup legend="Transportmiddel" onChange={handleChange}>
                        <Checkbox value="car">Bil</Checkbox>
                        <Checkbox value="taxi">Drosje</Checkbox>
                        <Checkbox value="public">Kollektivt</Checkbox>
                    </CheckboxGroup>
                    <label>
                        <input
                            type="checkbox"
                            checked={selectedHosts.has(host.Host)}
                            onChange={(e) => handleChange(host.Host, e.target.checked)}
                        />
                        {host.Name}
                    </label>
                </div>
            ))}
        </div>
    );
};

export const FirewallTagSelector = (props: FirewallTagSelectorProps) => {
    const onpremMapping = useOnpremMapping()
    const workstationOnpremMapping = useWorkstationOnpremMapping()

    const defaultFirewallTags = workstationOnpremMapping.data?.hosts || [];
    const onpremHosts = onpremMapping.data?.hosts

    const [selectedFirewallTags, setSelectedFirewallTags] = useState(new Set(defaultFirewallTags))

    const handleChange = (tagValue: string, isSelected: boolean) => {
        if (isSelected) {
            setSelectedFirewallTags(new Set(selectedFirewallTags.add(tagValue)))
        }

        selectedFirewallTags.delete(tagValue)
        setSelectedFirewallTags(new Set(selectedFirewallTags))
    }

    return (
    <div>
        {onpremHosts && Object.entries(onpremHosts).map(([type, hosts]) => {
            switch (type) {
                case OnpremHostTypePostgres:
                    return (
                        <div key={type}>
                            <h3>Postgres Hosts</h3>
                            <HostsList hosts={hosts} />
                        </div>
                    )
                case OnpremHostTypeTNS:
                    return (
                        <div key={type}>
                            <h3>TNS Hosts</h3>
                            <HostsChecked hosts={hosts} />
                        </div>
                    )
                // Add more cases as needed
                default:
                    return null;
            }
        })}
    </div>
        // <UNSAFE_Combobox
        //     label="Velg hvilke onprem-kilder du trenger åpninger mot"
        //     options={firewallTags.map((o) => ({label: o.name, value: o.name}))}
        //     selectedOptions={Array.from(selectedFirewallTags)}
        //     isMultiSelect
        //     onToggleSelected={(tagValue, isSelected) => handleChange(tagValue, isSelected)}
        //     enabled={enabled}
        // />
    );
}



export default FirewallTagSelector;
