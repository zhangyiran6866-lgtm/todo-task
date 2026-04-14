import { defineStore } from 'pinia'
import { ref } from 'vue'
import { taskApi, type Task, type CreateTaskReq, type UpdateTaskReq } from '@/api/task'

export const useTaskStore = defineStore('task', () => {
  const tasks = ref<Task[]>([])
  const nextCursor = ref<string>('')
  const isLoading = ref<boolean>(false)
  const isRefreshing = ref<boolean>(false)

  // 筛选相关的状态
  const filterStatus = ref<string>('')
  const filterPriority = ref<string>('')

  // 获取任务列表（支持下拉加载与重新刷新）
  async function fetchTasks(loadMore = false) {
    if (isLoading.value || isRefreshing.value) return
    
    if (loadMore && !nextCursor.value) return // 没有更多了

    if (!loadMore) {
      isRefreshing.value = true
    } else {
      isLoading.value = true
    }
    
    try {
      const resp = await taskApi.getTasks({
        status: filterStatus.value || undefined,
        priority: filterPriority.value || undefined,
        limit: 20,
        cursor: loadMore ? nextCursor.value : undefined
      })
      
      if (loadMore) {
        tasks.value.push(...resp.items)
      } else {
        tasks.value = resp.items
      }
      nextCursor.value = resp.next_cursor || ''
    } catch (e) {
      console.error('Failed to fetch tasks:', e)
    } finally {
      isLoading.value = false
      isRefreshing.value = false
    }
  }

  // 重置条件并刷新列表
  async function applyFilters(status: string, priority: string) {
    filterStatus.value = status
    filterPriority.value = priority
    await fetchTasks(false)
  }

  // 创建任务：乐观更新
  async function createTask(payload: CreateTaskReq) {
    // 乐观地发送请求，拿到新建的任务模型直接塞到数组最前面
    const newTask = await taskApi.createTask(payload)
    tasks.value.unshift(newTask)
    return newTask
  }

  // 更新任务：悲观或乐观更新
  async function updateTask(id: string, payload: UpdateTaskReq) {
    await taskApi.updateTask(id, payload)
    
    // 成功后，同步本地状态
    const target = tasks.value.find(t => t.id === id)
    if (target) {
      Object.assign(target, payload)
    }
  }

  // 删除任务：从列表中移除
  async function deleteTask(id: string) {
    await taskApi.deleteTask(id)
    tasks.value = tasks.value.filter(t => t.id !== id)
  }

  return {
    tasks,
    nextCursor,
    isLoading,
    isRefreshing,
    filterStatus,
    filterPriority,
    fetchTasks,
    applyFilters,
    createTask,
    updateTask,
    deleteTask
  }
})
