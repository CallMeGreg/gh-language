package cmd

type languageDetails struct {
	Size      int    `json:"size"`
	CreatedAt string `json:"createdAt"`
	Node      struct {
		Name string `json:"name"`
	} `json:"node"`
}

type Repo struct {
	Languages     []languageDetails `json:"languages"`
	CreatedAt     string            `json:"createdAt"`
	NameWithOwner string            `json:"nameWithOwner"`
}

type languageCount struct {
	name  string
	count int
}

func Red(s string) string {
	return "\x1b[31m" + s + "\x1b[m"
}

func Green(s string) string {
	return "\x1b[32m" + s + "\x1b[m"
}

func Yellow(s string) string {
	return "\x1b[33m" + s + "\x1b[m"
}

func Blue(s string) string {
	return "\x1b[34m" + s + "\x1b[m"
}

func Gray(s string) string {
	return "\x1b[90m" + s + "\x1b[m"
}
