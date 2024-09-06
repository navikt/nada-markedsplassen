import { useGetDataset } from "../../../lib/rest/dataproducts"
import ErrorMessage from "../../lib/error"
import LoaderSpinner from "../../lib/spinner"
import EditDatasetForm from "./editDatasetForm"

interface EditDatasetProps {
    datasetID: string
    setEdit: (val: boolean) => void
}

const EditDataset = ({datasetID, setEdit}: EditDatasetProps) => {
    const { dataset, loading, error } = useGetDataset(datasetID)

    if (error) return <ErrorMessage error={error} />
    if (loading || !dataset) return <LoaderSpinner />

    return (
        <EditDatasetForm dataset={dataset} setEdit={setEdit}/>
    )
}

export default EditDataset;
