import { useRouter } from "next/router"
import { EditStoryMetadataForm } from "../../../components/stories/editStoryMetadata";
import LoaderSpinner from "../../../components/lib/spinner";
import { fetchStoryMetadataURL, fetchTemplate } from "../../../lib/rest/restApi";
import { useEffect, useState } from "react";
import ErrorMessage from "../../../components/lib/error";

const getStoryMetadata = async (id: string) => {
    const url = fetchStoryMetadataURL(id);
    return fetchTemplate(url)
}

export const useGetStoryMetadata = (id: string)=>{
    const [storyMetadata, setStoryMetadata] = useState<any>(null)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)


    useEffect(()=>{
        if(!id) return
        getStoryMetadata(id).then((res)=> res.json())
        .then((story)=>
        {
            setError(null)
            setStoryMetadata(story)
        })
        .catch((err)=>{
            setError(err)
            setStoryMetadata(null)            
        }).finally(()=>{
            setLoading(false)
        })
    }, [id])

    return {storyMetadata, loading, error}
}


const EditStoryPage = ()=>{
    const router = useRouter()
    const id = router.query.id;
    const data = useGetStoryMetadata(id as string)

    if (data.error) return <ErrorMessage error={data.error} />
    if (data.loading || !data.storyMetadata)
      return <LoaderSpinner />

    const story = data.storyMetadata

    return <div>
        <EditStoryMetadataForm 
            id={id as string} 
            name={story.name} 
            description={story.description} 
            keywords={story.keywords} 
            teamkatalogenURL={story.teamkatalogenURL || ""} 
            group={story.group} />
    </div>
}

export default EditStoryPage;