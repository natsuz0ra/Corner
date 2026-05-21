import { apiClient } from './client'
import type { MemorySnapshot, MemoryTarget } from '@/types/settings'

export const memoryAPI = {
  get: async (): Promise<MemorySnapshot> => {
    return (await apiClient.get<MemorySnapshot>('/api/memory')).data
  },
  clear: async (target: MemoryTarget | 'all') => {
    return apiClient.delete(`/api/memory/${target}`)
  },
  deleteEntry: async (target: MemoryTarget, index: number) => {
    return apiClient.delete(`/api/memory/${target}/entries/${index}`)
  },
}
