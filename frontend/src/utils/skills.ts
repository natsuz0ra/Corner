import type { SkillItem } from '@/types/settings'

export function sortSkillRows(skills: SkillItem[]) {
  return [...skills].sort((a, b) => {
    if ((a.provider || '') !== (b.provider || '')) {
      return (a.provider || '').localeCompare(b.provider || '')
    }
    const aTime = new Date(a.uploadedAt || 0).getTime()
    const bTime = new Date(b.uploadedAt || 0).getTime()
    if (aTime !== bTime) return bTime - aTime
    return a.id.localeCompare(b.id)
  })
}

export function canDeleteSkill(skill: Pick<SkillItem, 'readOnly'>) {
  return !skill.readOnly
}
