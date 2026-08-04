import instance from '@/utils/request.ts'

export interface UploadReply {
  id: string
  url: string
}

/** 上传图片（multipart），返回文件记录 id 与下载路径 */
export async function uploadImage(file: File): Promise<UploadReply> {
  const form = new FormData()
  form.append('file', file)
  return await instance.post('/file/upload/raw/v1', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
