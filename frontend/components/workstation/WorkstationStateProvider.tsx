import React, {createContext, useReducer, useContext, ReactNode} from 'react';
import {
    WorkstationOutput,
    WorkstationOptions,
    WorkstationLogs,
    WorkstationJobs,
    WorkstationStartJobs,
    WorkstationZonalTagBindingJobs,
    EffectiveTags, EffectiveTag, WorkstationJob, WorkstationStartJob, WorkstationZonalTagBindingJob
} from '../../lib/rest/generatedDto';

interface WorkstationState {
    workstation?: WorkstationOutput;
    workstationOptions?: WorkstationOptions | null;
    workstationLogs?: WorkstationLogs | null;
    workstationJobs?: WorkstationJob[] | null;
    workstationStartJobs?: WorkstationStartJob[] | null;
    workstationZonalTagBindingJobs?: WorkstationZonalTagBindingJob[] | null;
    effectiveTags?: EffectiveTag[] | null;
}

type WorkstationAction =
    | { type: 'SET_WORKSTATION'; payload: WorkstationOutput }
    | { type: 'SET_WORKSTATION_OPTIONS'; payload: WorkstationOptions | null }
    | { type: 'SET_WORKSTATION_LOGS'; payload: WorkstationLogs | null }
    | { type: 'SET_WORKSTATION_JOBS'; payload: WorkstationJobs | null }
    | { type: 'SET_WORKSTATION_START_JOBS'; payload: WorkstationStartJobs | null }
    | { type: 'SET_WORKSTATION_ZONAL_TAG_BINDING_JOBS'; payload: WorkstationZonalTagBindingJobs | null }
    | { type: 'SET_EFFECTIVE_TAGS'; payload: EffectiveTags | null };

export const WorkstationStateContext = createContext<WorkstationState | undefined>(undefined);
export const WorkstationStateDispatchContext = createContext<React.Dispatch<WorkstationAction> | undefined>(undefined);

function workstationReducer(state: WorkstationState, action: WorkstationAction): WorkstationState {
    switch (action.type) {
        case 'SET_WORKSTATION':
            state.workstation = action.payload;
            return state
        case 'SET_WORKSTATION_OPTIONS':
            state.workstationOptions = action.payload;
            return state
        case 'SET_WORKSTATION_LOGS':
            state.workstationLogs = action.payload;
            return state
        case 'SET_WORKSTATION_JOBS':
            state.workstationJobs = action.payload?.jobs.filter((job): job is WorkstationJob => job !== undefined);
            return state
        case 'SET_WORKSTATION_START_JOBS':
            state.workstationStartJobs = action.payload?.jobs.filter((job): job is WorkstationStartJob => job !== undefined);
            return state
        case 'SET_WORKSTATION_ZONAL_TAG_BINDING_JOBS':
            state.workstationZonalTagBindingJobs = action.payload?.jobs.filter((job): job is WorkstationZonalTagBindingJob => job !== undefined);
            return state
        case 'SET_EFFECTIVE_TAGS':
            state.effectiveTags = action.payload?.tags?.filter((tag): tag is EffectiveTag => tag !== undefined);
            return state
        default:
            return state;
    }
}

export const WorkstationStateProvider = ({children, initialState = {}}: { children: ReactNode, initialState: WorkstationState }) => {
    const [state, dispatch] = useReducer(workstationReducer, initialState);

    return (
        <WorkstationStateContext.Provider value={state}>
            <WorkstationStateDispatchContext.Provider value={dispatch}>
                {children}
            </WorkstationStateDispatchContext.Provider>
        </WorkstationStateContext.Provider>
    );
};

export const useWorkstation = () => {
    const context = useContext(WorkstationStateContext);
    if (context === undefined) {
        throw new Error('useWorkstation must be used within a WorkstationProvider');
    }

    return context;
};

export const useWorkstationDispatch = () => {
    const context = useContext(WorkstationStateDispatchContext);
    if (context === undefined) {
        throw new Error('useWorkstationDispatch must be used within a WorkstationProvider');
    }

    return context;
};
