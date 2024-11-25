import {useEffect, useState} from "react";
import {useQuery} from "react-query";
import {getWorkstationJobs} from "../../../lib/rest/workstation";
import {
    WorkstationJob,
    WorkstationJobStateFailed,
    WorkstationJobs, WorkstationJobStateCompleted
} from "../../../lib/rest/generatedDto";
import {useWorkstationDispatch} from "../WorkstationStateProvider";
import {HttpError} from "../../../lib/rest/request";

const usePollingWorkstationJobs = (callbacks: { onComplete?: () => void } = {}) => {
    const dispatch = useWorkstationDispatch()
    const [polling, setPolling] = useState(false);

    const {data, refetch} = useQuery<WorkstationJobs, HttpError>(['workstationJobs'], () => getWorkstationJobs(), {
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
            const allJobsCompleted = data?.jobs.filter((job): job is WorkstationJob => job !== undefined).every(
                (job: WorkstationJob) =>
                    job.state === WorkstationJobStateCompleted || job.state === WorkstationJobStateFailed
            );

            if (allJobsCompleted) {
                setPolling(false);

                if (callbacks.onComplete) {
                    callbacks.onComplete();
                }
            }

            dispatch({type: 'SET_WORKSTATION_JOBS', payload: data});
        }
    }, [data, refetch]);

    const startPolling = () => setPolling(true);
    const stopPolling = () => setPolling(false);

    return {startPolling, stopPolling};
};

export default usePollingWorkstationJobs;
