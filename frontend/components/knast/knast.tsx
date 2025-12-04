import { useWorkstationExists, useWorkstationJobs, useWorkstationOptions } from "./queries";
import { Loader } from "@navikt/ds-react";
import CreateKnastForm from "./createKnastForm";
import { ManageKnastPage } from "./manageKnast";

const Knast = () => {
    const workstationExists = useWorkstationExists()
    const workstationJobs = useWorkstationJobs()
    const knastOptions = useWorkstationOptions()

    const page = workstationExists.isLoading || workstationJobs.isLoading
        ? "loading"
        : !workstationExists.data
            ? "create"
            : "manage"

    return (<div>
        {page === "loading" ? <div>Laster knast informasjon... <Loader /></div>
            : page === "create" ? <CreateKnastForm options={knastOptions.data} />
                : <ManageKnastPage />}
    </div>
    )
}

export default Knast
