import {useEffect, useState} from "react";
import {useQuery} from "react-query";
import {getWorkstationZonalTagBindingJobs} from "../../../lib/rest/workstation";
import {
    WorkstationJobStateFailed,
    WorkstationJobStateCompleted,
    WorkstationZonalTagBindingJobs,
    WorkstationZonalTagBindingJob
} from "../../../lib/rest/generatedDto";
import {useWorkstationDispatch} from "../WorkstationStateProvider";
import {HttpError} from "../../../lib/rest/request";

const usePollingWorkstationZonalTagBindingJobs = (callbacks: { onComplete?: () => void } = {})=> {
    const dispatch = useWorkstationDispatch()
    const [polling, setPolling] = useState(false);

    const {
        data,
        refetch
    } = useQuery<WorkstationZonalTagBindingJobs, HttpError>(['workstationStartJobs'], () => getWorkstationZonalTagBindingJobs(), {
        enabled: false
    })

    useEffect(() => {
        if (polling) {
            const interval = setInterval(() => {
                refetch();
            }, 1000); // Poll every 1 seconds, until jobs are completed

            return () => clearInterval(interval);
        }
    }, [polling, refetch]);

    useEffect(() => {
        if (data) {
            const allJobsCompleted = data?.jobs?.filter((job): job is WorkstationZonalTagBindingJob => job !== undefined).every(
                (job: WorkstationZonalTagBindingJob) =>
                    job.state === WorkstationJobStateCompleted || job.state === WorkstationJobStateFailed
            );

            if (allJobsCompleted) {
                setPolling(false);

                if (callbacks.onComplete) {
                    callbacks.onComplete();
                }
            }

            dispatch({type: 'SET_WORKSTATION_ZONAL_TAG_BINDING_JOBS', payload: data});
        }
    }, [data, refetch]);

    const startPolling = () => setPolling(true);
    const stopPolling = () => setPolling(false);

    return {startPolling, stopPolling};
};

export default usePollingWorkstationZonalTagBindingJobs;
