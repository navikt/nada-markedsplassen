import {createQueryKeyStore} from '@lukemorales/query-key-factory';
import {useQuery, useMutation, useQueryClient} from 'react-query';
import {
    getWorkstation,
    getWorkstationOptions,
    getWorkstationJobs,
    getWorkstationLogs,
    getWorkstationZonalTagBindingJobs,
    getWorkstationStartJobs,
    createWorkstationJob,
    getWorkstationZonalTagBindings,
    createWorkstationZonalTagBindingJob,
    stopWorkstation, startWorkstation, getWorkstationStartJob
} from '../../lib/rest/workstation';
import {
    EffectiveTags, KindNotExist,
    Workstation_STATE_RUNNING,
    WorkstationJobs,
    WorkstationLogs, WorkstationOptions,
    WorkstationOutput, WorkstationStartJob, WorkstationStartJobs, WorkstationZonalTagBindingJobs
} from "../../lib/rest/generatedDto";
import {HttpError} from "../../lib/rest/request";

export const queries = createQueryKeyStore({
    workstations: {
        mine: null,
        options: null,
        job: (id: string) => [id],
        jobs: null,
        startJob: (id: string) => [id],
        startJobs: null,
        logs: null,
        zonalTagBindingJobs: null,
        effectiveTags: null,
    }
});

export function useWorkstationMine() {
    return useQuery<WorkstationOutput, HttpError>({
        ...queries.workstations.mine,
        queryFn: getWorkstation,
    });
}

export function useWorkstationExists() {
    return useQuery<boolean, HttpError>({
        ...queries.workstations.mine,
        queryFn: async () => {
            try {
                await getWorkstation();
                return true;
            } catch (error) {
                const httpError = error as HttpError;

                if (httpError.kind === KindNotExist) {
                    return false;
                }

                throw error;
            }
        },
        refetchInterval: false,
    });
}

export function useWorkstationOptions() {
    return useQuery<WorkstationOptions, HttpError>({
        ...queries.workstations.options,
        queryFn: getWorkstationOptions,
    });
}

export function useWorkstationJobs() {
    return useQuery<WorkstationJobs, HttpError>({
        ...queries.workstations.jobs,
        queryFn: getWorkstationJobs,
    });
}

export function useWorkstationStartJobs() {
    return useQuery<WorkstationStartJobs, HttpError>({
        ...queries.workstations.startJobs,
        queryFn: getWorkstationStartJobs,
    });
}

export function useWorkstationLogs() {
    const {data: workstationData} = useWorkstationMine();
    const isRunning = workstationData?.state === Workstation_STATE_RUNNING;

    return useQuery<WorkstationLogs, HttpError>({
        ...queries.workstations.logs,
        queryFn: getWorkstationLogs,
        refetchInterval: 5000,
        enabled: isRunning,
    });
}

export function useWorkstationZonalTagBindingJobs() {
    const {data: workstationData} = useWorkstationMine();
    const isRunning = workstationData?.state === Workstation_STATE_RUNNING;

    return useQuery<WorkstationZonalTagBindingJobs, HttpError>({
        ...queries.workstations.zonalTagBindingJobs,
        queryFn: getWorkstationZonalTagBindingJobs,
        refetchInterval: 5000,
        enabled: isRunning,
    });
}

export function useWorkstationEffectiveTags() {
    const {data: workstationData} = useWorkstationMine();
    const isRunning = workstationData?.state === Workstation_STATE_RUNNING;

    return useQuery<EffectiveTags, HttpError>({
        ...queries.workstations.effectiveTags,
        queryFn: getWorkstationZonalTagBindings,
        refetchInterval: 5000,
        enabled: isRunning,
    });
}

export const useStartWorkstation = () => {
    const queryClient = useQueryClient();

    return useMutation<WorkstationStartJob, HttpError>({
        mutationFn: startWorkstation,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.startJobs).then(r => console.log(r));
        },
    });
};

export const useStopWorkstation = () => {
    const queryClient = useQueryClient();

    return useMutation<void, HttpError>({
        mutationFn: stopWorkstation,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.startJobs).then(r => console.log(r));
        },
    });
};

export const useCreateZonalTagBindingJob = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: createWorkstationZonalTagBindingJob,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.zonalTagBindingJobs).then(r => console.log(r));
        },
    });
};

export const useCreateWorkstationJob = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: createWorkstationJob,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.jobs).then(r => console.log(r));
        },
    });
}

export const usePollWorkstationStartJob = (jobId: string) => {
    return useQuery({
        ...queries.workstations.startJob(jobId),
        queryFn: () => getWorkstationStartJob(jobId),
        refetchInterval: 1000,
        enabled: !!jobId,
    });
};
