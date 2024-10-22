import { UploadFile } from "../../components/stories/newStory"
import { buildUrl } from "./apiUrl"
import { NewStory, UpdateStoryDto } from "./generatedDto"
import { deleteTemplate, fetchTemplate, postTemplate, putTemplate } from "./request"

const storyPath = buildUrl('stories')
const buildCreateStoryUrl = () => storyPath('new')()
const buildUpdateStoryUrl = (id: string) => storyPath(id)()
const buildDeleteStoryUrl = (id: string) => storyPath(id)()
const buildFetchStoryMetadataUrl = (id: string) => storyPath(id)()

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
