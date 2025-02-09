import {
    WorkstationDiffContainerImage,
    WorkstationDiffMachineType,
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
        </div>
    );
};

export default JobViewerComponent;
