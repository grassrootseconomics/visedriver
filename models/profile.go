package models

type Profile struct {
	ProfileItems []string
	Max          int
}

func (p *Profile) InsertOrShift(index int, value string) {
	if index < len(p.ProfileItems) {
		p.ProfileItems = append(p.ProfileItems[:index], value)
	} else {
		p.ProfileItems = append(p.ProfileItems, value)
	}
}
