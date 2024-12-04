import {
    WorkstationDiffContainerImage,
    WorkstationDiffMachineType, WorkstationDiffOnPremAllowList,
    WorkstationDiffURLAllowList,
    WorkstationJob
} from "../../lib/rest/generatedDto";
import {Heading} from "@navikt/ds-react";
import {WorkstationDiffDescriptions} from "./DiffViewerComponent";

interface JobViewerProps {
    job: WorkstationJob | undefined;
}

const JobViewerComponent: React.FC<JobViewerProps> = ({job}) => {
    if (!job) {
        return;
    }

    return (
        <div>
            {job.machineType && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffMachineType]}</Heading>
                    <p>{job.machineType}</p>
                </>
            )}
            {job.containerImage && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffContainerImage]}</Heading>
                    <p>{job.containerImage}</p>
                </>
            )}
            {job.urlAllowList && job.urlAllowList.length > 0 && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffURLAllowList]}</Heading>
                    <p>{job.urlAllowList.join(', ')}</p>
                </>
            )}
            {job.onPremAllowList && job.onPremAllowList.length > 0 && (
                <>
                    <Heading size="xsmall">{WorkstationDiffDescriptions[WorkstationDiffOnPremAllowList]}</Heading>
                    <p>{job.onPremAllowList.join(', ')}</p>
                </>
            )}
        </div>
    );
};

export default JobViewerComponent;
