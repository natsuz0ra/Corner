import { apiClient } from './client'
import type { UpdateCheckResult, UpdateJobStatus } from '@/types/update'
import { normalizeUpdateCheck, normalizeUpdateJob } from '@/utils/updateStatus'

export const updateAPI = {
  check: async (force = false): Promise<UpdateCheckResult> => {
    const data = (await apiClient.get<Partial<UpdateCheckResult>>('/api/update/check', {
      params: force ? { force: '1' } : undefined,
    })).data
    return normalizeUpdateCheck(data)
  },
  job: async (): Promise<UpdateJobStatus> => {
    const data = (await apiClient.get<Partial<UpdateJobStatus>>('/api/update/job')).data
    return normalizeUpdateJob(data)
  },
  apply: async (targetVersion?: string): Promise<UpdateJobStatus> => {
    const data = (await apiClient.post<Partial<UpdateJobStatus>>('/api/update/apply', {
      targetVersion: targetVersion || '',
    })).data
    return normalizeUpdateJob(data)
  },
}
