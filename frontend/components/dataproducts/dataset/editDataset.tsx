import { useGetDataset } from "../../../lib/rest/dataproducts"
import ErrorStripe from "../../lib/errorStripe"
import LoaderSpinner from "../../lib/spinner"
import EditDatasetForm from "./editDatasetForm"

interface EditDatasetProps {
    datasetID: string
    setEdit: (val: boolean) => void
}

const EditDataset = ({datasetID, setEdit}: EditDatasetProps) => {
    const { data: dataset, isLoading: loading, error } = useGetDataset(datasetID)

    if (error) return <ErrorStripe error={error} />
    if (loading || !dataset) return <LoaderSpinner />

    return (
        <EditDatasetForm dataset={dataset} setEdit={setEdit}/>
    )
}

export default EditDataset;
