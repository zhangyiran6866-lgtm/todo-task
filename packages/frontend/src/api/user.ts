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

export const userApi = {
  getMe(): Promise<UserInfo> {
    return request.get('/users/me') as unknown as Promise<UserInfo>
  },

  changePassword(data: ChangePasswordReq): Promise<void> {
    return request.put('/users/me/password', data) as unknown as Promise<void>
  }
}
