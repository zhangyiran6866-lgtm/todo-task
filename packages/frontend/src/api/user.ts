import request from '@/utils/request'

export interface UserInfo {
  id: string
  email: string
  nickname: string
  language: string
  theme: string
  created_at?: string
}

export interface ChangePasswordReq {
  old_password: string
  new_password: string
}

export interface UpdateProfileReq {
  nickname?: string
  language?: 'zh' | 'en'
  theme?: 'cyan' | 'purple' | 'green' | 'pink'
}

export const userApi = {
  getMe(): Promise<UserInfo> {
    return request.get('/users/me') as unknown as Promise<UserInfo>
  },

  updateMe(data: UpdateProfileReq): Promise<UserInfo> {
    return request.patch('/users/me', data) as unknown as Promise<UserInfo>
  },

  changePassword(data: ChangePasswordReq): Promise<void> {
    return request.put('/users/me/password', data) as unknown as Promise<void>
  }
}
