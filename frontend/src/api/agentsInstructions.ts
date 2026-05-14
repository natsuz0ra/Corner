import { apiClient } from './client'

export interface AgentsInstructionsFile {
  content: string
  path: string
}

export const agentsInstructionsAPI = {
  get: async (): Promise<AgentsInstructionsFile> => {
    return (await apiClient.get<AgentsInstructionsFile>('/api/agents-instructions')).data
  },
  update: async (content: string) => {
    return apiClient.put('/api/agents-instructions', { content })
  },
}
