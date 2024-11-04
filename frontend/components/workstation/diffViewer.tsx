import {
    Diff,
    WorkstationDiffContainerImage,
    WorkstationDiffMachineType, WorkstationDiffOnPremAllowList,
    WorkstationDiffURLAllowList
} from "../../lib/rest/generatedDto";
import {Heading} from "@navikt/ds-react";
import {MinusCircleIcon, PlusCircleIcon} from "@navikt/aksel-icons";

export const WorkstationDiffDescriptions: { [key: string]: string } = {
    [WorkstationDiffContainerImage]: "Kjøremiljø",
    [WorkstationDiffMachineType]: "Maskin type",
    [WorkstationDiffURLAllowList]: "URL Filter",
    [WorkstationDiffOnPremAllowList]: "On-prem kilder",
};

interface DiffViewerProps {
    diff: { [key: string]: Diff | undefined };
}

const DiffViewerComponent: React.FC<DiffViewerProps> = ({diff}) => {
    console.log(diff)
    if (!diff || Object.keys(diff).length === 0) {
        return <div>Ingen endringer å vise.</div>;
    }

    return (
        <div className="diff-viewer">
            {Object.entries(diff).map(([key, value]) => {
                return (
                    <div key={key}>
                        <Heading size="xsmall">{WorkstationDiffDescriptions[key]}</Heading>
                        {value?.value ? (
                            <p>{value.value}</p>
                        ) : (
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
                        )}
                    </div>
                );
            })}
        </div>
    );
};

export default DiffViewerComponent;
