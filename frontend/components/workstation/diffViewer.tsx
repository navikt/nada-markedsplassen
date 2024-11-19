import {
    Diff,
    WorkstationDiffContainerImage, WorkstationDiffDisableGlobalURLAllowList,
    WorkstationDiffMachineType, WorkstationDiffOnPremAllowList,
    WorkstationDiffURLAllowList
} from "../../lib/rest/generatedDto";
import {Heading} from "@navikt/ds-react";
import {MinusCircleIcon, PlusCircleIcon} from "@navikt/aksel-icons";

export const WorkstationDiffDescriptions: { [key: string]: string } = {
    [WorkstationDiffDisableGlobalURLAllowList]: "Skru av globale åpninger",
    [WorkstationDiffContainerImage]: "Utviklingsmiljø",
    [WorkstationDiffMachineType]: "Maskintype",
    [WorkstationDiffURLAllowList]: "Tillate URL-er",
    [WorkstationDiffOnPremAllowList]: "On-prem kilder",
};

interface DiffViewerProps {
    diff: { [key: string]: Diff | undefined };
}

const DiffViewerComponent: React.FC<DiffViewerProps> = ({diff}) => {
    if (!diff || Object.keys(diff).length === 0) {
        return <div>Ingen endringer å vise.</div>;
    }

    return (
        <div className="diff-viewer">
            {Object.entries(diff).map(([key, value]) => {
                return (
                    <div key={key}>
                        <Heading size="xsmall">{WorkstationDiffDescriptions[key]}</Heading>
                        <div>
                            {(value?.added?.length ?? 0) > 0 && (
                                <div><PlusCircleIcon title="lagt til" fontSize="1.5rem"/><p
                                    style={{color: 'green'}}>{value?.added.join(', ')}</p></div>
                            )}
                            {(value?.removed?.length ?? 0) > 0 && (
                                <div><MinusCircleIcon title="fjernet" fontSize="1.5rem"/><p
                                    style={{color: 'red'}}>{value?.removed.join(', ')}</p></div>
                            )}
                        </div>
                    </div>
                );
            })}
        </div>
    );
};

export default DiffViewerComponent;
