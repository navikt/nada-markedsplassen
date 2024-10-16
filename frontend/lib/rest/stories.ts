import { UploadFile } from "../../components/stories/newStory"
import { buildCreateStoryUrl, buildDeleteStoryUrl, buildFetchStoryMetadataUrl, buildUpdateStoryUrl } from "./apiUrl"
import { NewStory, UpdateStoryDto } from "./generatedDto"
import { deleteTemplate, fetchTemplate, postTemplate, putTemplate } from "./request"

export const fetchStoryMetadata = (id: string)=>
    fetchTemplate(buildFetchStoryMetadataUrl(id))

export const createStory = (newStory: NewStory, files: UploadFile[]) => {
    const formData = new FormData()
    files.forEach((file) => {
        formData.append(file.path, file.file)
    })
    formData.append('nada-backend-new-story', JSON.stringify(newStory))
    return postTemplate(buildCreateStoryUrl(), formData)    
}

export const updateStory =(storyId: string, updatedStory: UpdateStoryDto) => 
    putTemplate(buildUpdateStoryUrl(storyId), updatedStory)

export const deleteStory = (storyId: string) => 
    deleteTemplate(buildDeleteStoryUrl(storyId))
