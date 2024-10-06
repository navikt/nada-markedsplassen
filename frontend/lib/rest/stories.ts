import { UploadFile } from "../../components/stories/newStory"
import { NewStory, UpdateStoryDto } from "./generatedDto"
import { createStoryUrl, deleteTemplate, postTemplate, putTemplate, updateStoryUrl } from "./restApi"

export const createStory = (newStory: NewStory, files: UploadFile[]) => {
    const formData = new FormData()
    files.forEach((file) => {
        formData.append(file.path, file.file)
    })
    formData.append('nada-backend-new-story', JSON.stringify(newStory))
    return fetch(createStoryUrl(), {
        method: 'POST',
        credentials: 'include',
        body: formData,
      }).then(res => {
        if (!res.ok) {
          throw new Error(res.statusText)
        }
        return res.json()
      })    
}

export const updateStory =(storyId: string, updatedStory: UpdateStoryDto) => {
    return putTemplate(updateStoryUrl(storyId), updatedStory).then((res) => res.json())
}

export const deleteStory = (storyId: string) => {
    return deleteTemplate(updateStoryUrl(storyId), {isDeleted: true}).then((res) => res.json())
}
