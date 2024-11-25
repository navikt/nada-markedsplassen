import {useEffect, useState} from "react";
import {useQuery} from "react-query";
import {getWorkstationStartJobs} from "../../../lib/rest/workstation";
import {
    WorkstationJobStateFailed,
    WorkstationJobStateCompleted, WorkstationStartJobs, WorkstationStartJob
} from "../../../lib/rest/generatedDto";
import {useWorkstationDispatch} from "../WorkstationStateProvider";
import {HttpError} from "../../../lib/rest/request";

const usePollingWorkstationStartJobs = (callbacks: { onComplete?: () => void } = {}) => {
    const dispatch = useWorkstationDispatch()
    const [polling, setPolling] = useState(false);

    const {
        data,
        refetch
    } = useQuery<WorkstationStartJobs, HttpError>(['workstationStartJobs'], () => getWorkstationStartJobs(), {
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
            const allJobsCompleted = data?.jobs?.filter((job): job is WorkstationStartJob => job !== undefined).every(
                (job: WorkstationStartJob) =>
                    job.state === WorkstationJobStateCompleted || job.state === WorkstationJobStateFailed
            );

            if (allJobsCompleted) {
                setPolling(false);

                if (callbacks.onComplete) {
                    callbacks.onComplete();
                }
            }

            dispatch({type: 'SET_WORKSTATION_START_JOBS', payload: data});
        }
    }, [data, refetch]);

    const startPolling = () => setPolling(true);
    const stopPolling = () => setPolling(false);

    return {startPolling, stopPolling};
};

export default usePollingWorkstationStartJobs;
