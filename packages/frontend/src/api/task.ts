import request from '@/utils/request'

export interface Task {
  id: string
  user_id: string
  title: string
  description: string
  status: 'todo' | 'in_progress' | 'done'
  priority: 'low' | 'important' | 'urgent' | 'critical' | 'routine'
  due_at: string | null
  created_at: string
  updated_at: string
}

export interface ListTasksReq {
  status?: string
  priority?: string
  limit?: number
  cursor?: string
}

export interface ListTasksResp {
  items: Task[]
  next_cursor: string
}

export interface CreateTaskReq {
  title: string
  description?: string
  priority?: string
  due_at?: string
}

export interface UpdateTaskReq {
  title?: string
  status?: string
  priority?: string
  due_at?: string
  description?: string
}

export const taskApi = {
  getTasks(params?: ListTasksReq): Promise<ListTasksResp> {
    return request.get('/tasks', { params }) as Promise<any>
  },
  
  getTaskById(id: string): Promise<Task> {
    return request.get(`/tasks/${id}`) as Promise<any>
  },

  createTask(data: CreateTaskReq): Promise<Task> {
    return request.post('/tasks', data) as Promise<any>
  },

  updateTask(id: string, data: UpdateTaskReq): Promise<void> {
    return request.patch(`/tasks/${id}`, data) as Promise<any>
  },

  deleteTask(id: string): Promise<void> {
    return request.delete(`/tasks/${id}`) as Promise<any>
  }
}
