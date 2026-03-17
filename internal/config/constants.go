package config

const (
	Marker           = "AGENTFLOW_ROUTER:"
	PluginDirName    = "agentflow"
	DefaultProfile   = "full"
	RulesSourceFile  = "AGENTS.md"
	SkillSourceFile  = "SKILL.md"
	ModuleSourceDir  = "agentflow"
	GlobalRulesDir   = ".agentflow" // ~/.agentflow/ centralized rules directory
	GlobalRulesFile  = "AGENTS.md"  // ~/.agentflow/AGENTS.md full compiled rules
	LangZH           = "zh"
	LangEN           = "en"
	DefaultLang      = LangZH
	TimeStampLayout  = "20060102150405"
	TempFilePattern  = ".agentflow_*.tmp"
	RenameAsideToken = "._agentflow_old_"
	BackupToken      = "_bak"
	markerScanLimit  = 1024
)

// ValidLang reports whether lang is a supported rules language.
func ValidLang(lang string) bool {
	return lang == LangZH || lang == LangEN
}
