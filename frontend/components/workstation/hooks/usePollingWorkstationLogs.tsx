import {useEffect, useState} from "react";
import {useQuery} from "react-query";
import {getWorkstationLogs} from "../../../lib/rest/workstation";
import {WorkstationLogs} from "../../../lib/rest/generatedDto";
import {useWorkstationDispatch} from "../WorkstationStateProvider";
import {HttpError} from "../../../lib/rest/request";

const usePollingWorkstationLogs = () => {
    const dispatch = useWorkstationDispatch()
    const [polling, setPolling] = useState(false);

    const {data, refetch} = useQuery<WorkstationLogs, HttpError>(['workstationLogs'], () => getWorkstationLogs(), {
        enabled: false
    })

    useEffect(() => {
        if (polling) {
            const interval = setInterval(() => {
                refetch();
            }, 1000); // Poll every 5 seconds for new logs

            return () => clearInterval(interval);
        }
    }, [polling, refetch]);

    useEffect(() => {
        if (data) {
            dispatch({type: 'SET_WORKSTATION_LOGS', payload: data});
        }
    }, [data, refetch]);

    const startPolling = () => setPolling(true);
    const stopPolling = () => setPolling(false);

    return {startPolling, stopPolling};
};

export default usePollingWorkstationLogs;
