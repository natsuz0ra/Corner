package domain

// SkillStore reads and deletes skill metadata from the configured skills directory.
type SkillStore interface {
	ListSkills() ([]Skill, error)
	GetSkillByName(name string) (*Skill, error)
	GetSkillByID(id string) (*Skill, error)
	CreateSkill(item Skill) (*Skill, error)
	DeleteSkill(id string) error
}
