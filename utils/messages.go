package utils

type FocusMsg struct {
	ID string
}

type BlurMsg struct {
	ID string
}

type UpdateSizeMsg struct {
	ID     string
	Width  int
	Height int
}
