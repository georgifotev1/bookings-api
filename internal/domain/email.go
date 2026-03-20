package domain

type EmailData struct {
	From     string
	To       []string
	Subject  string
	Template string
	Data     any
}
