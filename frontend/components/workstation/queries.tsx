import {createQueryKeyStore} from '@lukemorales/query-key-factory';
import {useQuery, useMutation, useQueryClient} from '@tanstack/react-query';
import {
    getWorkstation,
    getWorkstationOptions,
    getWorkstationJobs,
    getWorkstationLogs,
    createWorkstationJob,
    getWorkstationZonalTagBindings,
    stopWorkstation,
    startWorkstation,
    updateWorkstationUrlAllowList,
    getWorkstationURLList,
    getWorkstationOnpremMapping,
    updateWorkstationOnpremMapping,
    createWorkstationConnectivityWorkflow,
    getWorkstationConnectivityWorkflow,
    restartWorkstation,
    getWorkstationResyncJobs,
    getConfigWorkstationSSHJob,
} from '../../lib/rest/workstation'
import {
    ConfigWorkstationSSHJob,
    EffectiveTags,
    Workstation_STATE_RUNNING, WorkstationConnectivityWorkflow,
    WorkstationJobs,
    WorkstationLogs, WorkstationOnpremAllowList, WorkstationOptions,
    WorkstationOutput, WorkstationResyncJobs, WorkstationStartJob, WorkstationURLList
} from '../../lib/rest/generatedDto'
import {HttpError} from "../../lib/rest/request";

export const queries = createQueryKeyStore({
    workstations: {
        mine: null,
        exists: null,
        options: null,
        job: (id: string) => [id],
        jobs: null,
        resyncJobs: null,
        restart: null,
        startJob: (id: string) => [id],
        startJobs: null,
        logs: null,
        zonalTagBindingsJobs: null,
        effectiveTags: null,
        urlList: null,
        connectivity: null,
        onpremMapping: null,
        updateUrlAllowList: (urls: string[], disableGlobalURLList: boolean) => [urls, disableGlobalURLList],
        updateOnpremMapping: (mapping: string[]) => [mapping],
        configSSHjob: null,
    }
});

export function useCreateWorkstationConnectivityWorkflow() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: createWorkstationConnectivityWorkflow,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.connectivity).then(r => console.log(r));
        },
    });
}

export function useWorkstationConnectivityWorkflow() {
    return useQuery<WorkstationConnectivityWorkflow, HttpError>({
        ...queries.workstations.connectivity,
        queryFn: getWorkstationConnectivityWorkflow,
        refetchInterval: 5000,
    });
}

export function useWorkstationURLList() {
    return useQuery<WorkstationURLList, HttpError>({
        ...queries.workstations.urlList,
        queryFn: getWorkstationURLList,
    });
}

export function useWorkstationOnpremMapping() {
    return useQuery<WorkstationOnpremAllowList, HttpError>({
        ...queries.workstations.onpremMapping,
        queryFn: getWorkstationOnpremMapping,
    });
}

export function useUpdateWorkstationOnpremMapping() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: updateWorkstationOnpremMapping,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.onpremMapping).then(r => console.log(r));
        },
    });
}

export function useWorkstationExists() {
    return useQuery<boolean, HttpError>({
        ...queries.workstations.exists,
        queryFn: async () => {
            try {
                await getWorkstation();
                return true;
            } catch (error) {
                if ((error as HttpError).statusCode === 404) {
                    return false;
                }

                throw error;
            }
        },
        retry: false,
    });
}

export function useWorkstationMine() {
    return useQuery<WorkstationOutput, HttpError>({
        ...queries.workstations.mine,
        queryFn: getWorkstation,
        refetchInterval: 5000,
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
        refetchInterval: 5000,
    });
}

export function useConfigWorkstationSSHJobs() {
    return useQuery<ConfigWorkstationSSHJob, HttpError>({
        ...queries.workstations.configSSHjob,
        queryFn: getConfigWorkstationSSHJob,
        refetchInterval: 3000,
    });
}

export function useWorkstationResyncJobs() {
  return useQuery<WorkstationResyncJobs, HttpError>({
      ...queries.workstations.resyncJobs,
      queryFn: getWorkstationResyncJobs,
      refetchInterval: 5000,
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

export const useRestartWorkstation = () => {
    const queryClient = useQueryClient();

    return useMutation<void, HttpError>({
        mutationFn: restartWorkstation,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.mine).then(r => console.log(r));
        },
    });
}

export const useStartWorkstation = () => {
    const queryClient = useQueryClient();

    return useMutation<WorkstationStartJob, HttpError>({
        mutationFn: startWorkstation,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.startJobs).then(r => console.log(r));
            queryClient.invalidateQueries(queries.workstations.mine).then(r => console.log(r));
        },
    });
};

export const useStopWorkstation = () => {
    const queryClient = useQueryClient();

    return useMutation<void, HttpError>({
        mutationFn: stopWorkstation,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.startJobs).then(r => console.log(r));
            queryClient.invalidateQueries(queries.workstations.mine).then(r => console.log(r));
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

export const useUpdateUrlAllowList = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: updateWorkstationUrlAllowList,
        onSuccess: () => {
            queryClient.invalidateQueries(queries.workstations.urlList).then(r => console.log(r));
        },
    })
};
